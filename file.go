package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

var find = struct {
	lastMatch, lastCol, direction int
}{
	-1, -1, 1,
}

func editorOpen(filename string) {
	if len(filename) == 0 {
		return
	} else {
		data, err := os.ReadFile(filename)
		if err != nil && err != os.ErrNotExist {
			safeExit(err)
		}

		lines := strings.Split(string(data[:]), "\n")

		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}

		globalState.filename = filename
		globalState.trow = make([]erow, 0, len(lines))

		for _, line := range lines {
			line = strings.TrimRight(line, "\r")
			appendRow(line)
		}
	}

	globalState.dirty = 0
}

func editorSave() {
	dirty := globalState.dirty
	globalState.dirty = 0

	if globalState.filename == "" {
		globalState.filename = editorPrompt("Save as: %s", nil)
		if globalState.filename == "" {
			setStatusMessage("Save aborted")
			return
		}
	}

	fd, err := os.OpenFile(globalState.filename, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		saveErr(fd)
		safeExit(err)
	}

	strContent := rowsToString()
	if err = fd.Truncate(int64(len(strContent))); err != nil {
		saveErr(fd)
		safeExit(err)
	}

	if n, err := fd.Write([]byte(strContent)); err != nil {
		saveErr(fd)
		safeExit(err)
	} else {
		msg := fmt.Sprintf("%d bytes written to disk (%d changes)", n, dirty)
		setStatusMessage(msg)
	}

	if err = fd.Close(); err != nil {
		safeExit(err)
	}
}

func saveErr(fd *os.File) {
	setStatusMessage("ERR: Can't save to disk")
	if e := fd.Close(); e != nil {
		safeExit(e)
	}
}

func editorFind() {
	savedCx := globalState.cx
	savedCy := globalState.cy
	savedColoff := globalState.coloff
	savedRowoff := globalState.rowoff

	query := editorPrompt("Search: %s (ESC cancel,   next,   prev)", editorFindCb)
	if query == "" {
		globalState.cx = savedCx
		globalState.cy = savedCy
		globalState.coloff = savedColoff
		globalState.rowoff = savedRowoff
	}
}

func editorFindCb(query string, c int) {
	if c == '\r' || c == '\x1b' {
		find.lastMatch = -1
		find.lastCol = -1
		find.direction = 1
		return
	}

	switch c {
	case ARROW_RIGHT, ARROW_DOWN:
		find.direction = 1
	case ARROW_LEFT, ARROW_UP:
		find.direction = -1
	default:
		find.lastMatch = -1
		find.lastCol = -1
		find.direction = 1
	}

	n := globalState.nrows
	if query == "" || n == 0 {
		return
	}
	if find.lastMatch == -1 {
		find.direction = 1
	}
	cur, col := find.lastMatch, find.lastCol
	if cur == -1 {
		cur = 0
	}

	for i := 0; i <= n; i++ {
		row := &globalState.trow[cur]
		var idx int

		if find.direction == 1 {
			after := -1
			if i == 0 {
				after = col
			}
			idx = indexAfter(row.render, query, after)
		} else {
			before := len(row.render)
			if i == 0 {
				before = col
			}
			idx = lastIndexBefore(row.render, query, before)
		}

		if idx != -1 {
			find.lastMatch = cur
			find.lastCol = idx
			globalState.cy = cur
			globalState.cx = rxToCx(row, idx)
			globalState.rowoff = n
			return
		}

		cur += find.direction
		if cur < 0 {
			cur = n - 1
		} else if cur >= n {
			cur = 0
		}
	}
}

// First occurrence of q in s strictly after column 'after'
func indexAfter(s, q string, after int) int {
	start := after + 1
	if start > len(s) {
		return -1
	}

	i := strings.Index(s[start:], q)
	if i == -1 {
		return i
	}
	return start + i
}

// Last occurrence of q in s strictly before column 'before'
func lastIndexBefore(s, q string, before int) int {
	limit := min(before+len(q)-1, len(s))
	if limit < len(q) {
		return -1
	}

	return strings.Index(s[:limit], q)
}

func rowsToString() string {
	out := strings.Builder{}
	for i := range globalState.nrows {
		out.Write([]byte(globalState.trow[i].str))
		out.Write([]byte("\n"))
	}

	return out.String()
}

func appendRow(line string) {
	globalState.trow = append(globalState.trow, erow{str: line})
	pos := len(globalState.trow)
	updateRow(&globalState.trow[pos-1])
	globalState.nrows++
	globalState.dirty++
}

func insertRow(line string, pos int) {
	if pos < 0 || pos > globalState.nrows {
		return
	}

	globalState.trow = slices.Insert(globalState.trow, pos, erow{str: line})
	updateRow(&globalState.trow[pos])
	globalState.nrows++
	globalState.dirty++
}

func rowInsertChar(r *erow, pos int, c byte) {
	if pos < 0 || pos > len(r.str) {
		pos = len(r.str)
	}

	strByte := []byte(r.str)
	strByte = slices.Insert(strByte, pos, c)

	r.str = string(strByte)
	updateRow(r)

	globalState.dirty++
}

func rowDelChar(r *erow, pos int) {
	if pos < 0 || pos > len(r.str) {
		return
	}

	strByte := []byte(r.str)
	strByte = slices.Delete(strByte, pos, pos+1)

	r.str = string(strByte)
	updateRow(r)

	globalState.dirty++
}

func rowAppendString(r *erow, s string) {
	r.str += s
	updateRow(r)
	globalState.dirty++
}

func delRow(pos int) {
	if pos < 0 || pos > globalState.nrows {
		return
	}
	globalState.trow = slices.Delete(globalState.trow, pos, pos+1)
	globalState.nrows--
	globalState.dirty++
}

func insertChar(c byte) {
	if globalState.cy == globalState.nrows {
		appendRow("")
	}

	rowInsertChar(&globalState.trow[globalState.cy], globalState.cx, c)
	globalState.cx++
}

func insertNewline() {
	if globalState.cx == 0 {
		insertRow("", globalState.cy)
	} else {
		row := &globalState.trow[globalState.cy]
		insertRow(row.str[globalState.cx:], globalState.cy+1)
		row = &globalState.trow[globalState.cy]
		row.str = row.str[:globalState.cx]
		updateRow(row)
	}

	globalState.cy++
	globalState.cx = 0
}

func delChar() {
	if globalState.cy == globalState.nrows {
		return
	}
	if globalState.cx == 0 && globalState.cy == 0 {
		return
	}

	if globalState.cx > 0 {
		rowDelChar(&globalState.trow[globalState.cy], globalState.cx-1)
		globalState.cx--
	} else if globalState.cx == 0 {
		globalState.cx = len(globalState.trow[globalState.cy-1].str)
		rowAppendString(&globalState.trow[globalState.cy-1], globalState.trow[globalState.cy].str)
		delRow(globalState.cy)
		globalState.cy--
	}
}
