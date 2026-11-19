package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

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
		globalState.nrows = len(lines)
	}

	globalState.dirty = 0
}

func editorSave() {
	dirty := globalState.dirty
	globalState.dirty = 0

	if globalState.filename == "" {
		return
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
	globalState.dirty++
}

func insertRow(line string, pos int) {
	if pos < 0 || pos > globalState.nrows {
		return
	}

	globalState.trow = append(globalState.trow, erow{str: line})
	updateRow(&globalState.trow[pos])
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

// func insertNewline() {
// 	if globalState.cx == 0 {
// 		insertRow("", 0)
// 	} else {
// 		row := globalState.trow[globalState.cy]
// 		insertRow(, globalState.cy+1)
// 	}
// }

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
