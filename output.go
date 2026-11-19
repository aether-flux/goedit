package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

const tabstop = 8

type Screen struct {
	Rows int
	Cols int
}

func getTerminalSize() *Screen {
	ws, err := unix.IoctlGetWinsize(unix.Stdout, unix.TIOCGWINSZ)
	if err != nil {
		safeExit(err)
	} else if ws.Col == 0 {
		safeExit(nil)
	}

	return NewScreen(int(ws.Row), int(ws.Col))
}

func NewScreen(rows, cols int) *Screen {
	globalState.srows = rows - 2
	globalState.scols = cols
	return &Screen{rows, cols}
}

func (s *Screen) drawRows(buf *bytes.Buffer) {
	for y := range s.Rows {
		filerow := y + globalState.rowoff
		if filerow >= globalState.nrows {
			if globalState.nrows == 0 && y == s.Rows/2 {
				str := "Goedit -- v0.1"

				padding := (s.Cols - len(str)) / 2
				if padding > 0 {
					BufAppend(buf, "~")
					padding -= 1
				}
				for ; padding > 0; padding -= 1 {
					BufAppend(buf, " ")
				}

				BufAppend(buf, str)
			} else {
				BufAppend(buf, "~")
			}
		} else {
			line := renderString(globalState.trow[filerow].str)
			start := min(globalState.coloff, len(line))
			line = line[start:]
			if len(line) > s.Cols {
				line = line[:s.Cols]
			}
			BufAppend(buf, string(line))
		}

		BufAppend(buf, "\x1b[K")
		BufAppend(buf, "\r\n")
	}
}

func renderString(line string) []rune {
	out := make([]rune, 0, len(line))
	for _, r := range line {
		if r == '\t' {
			out = append(out, ' ')
			for len(out)%tabstop != 0 {
				out = append(out, ' ')
			}
		} else {
			out = append(out, r)
		}
	}

	return out
}

func cxToRx(trow *erow, cx int) int {
	rx := 0

	for i := range cx {
		if trow.str[i] == '\t' {
			rx += (tabstop - 1) - (rx % tabstop)
		}
		rx += 1
	}

	return rx
}

func (s *Screen) drawStatusBar(buf *bytes.Buffer) {
	BufAppend(buf, "\x1b[7m")

	filename := "[Untitled]"
	if globalState.filename != "" {
		filename = globalState.filename
	}

	status := fmt.Sprintf("%.20s - %d lines", filename, globalState.nrows)
	l := len(status)
	start := min(globalState.coloff, l)
	status = status[start:]
	if len(status) > s.Cols {
		status = status[:s.Cols]
	}
	BufAppend(buf, status)

	// perc := ((globalState.cy + 1) / (globalState.nrows - 2)) * 100
	lineInfo := fmt.Sprintf("%d / %d", globalState.cy+1, globalState.nrows-2)

	for ; l < globalState.scols; l += 1 {
		if globalState.scols-l == len(lineInfo) {
			BufAppend(buf, lineInfo)
			break
		} else {
			BufAppend(buf, " ")
		}
	}

	BufAppend(buf, "\x1b[m")
	BufAppend(buf, "\r\n")
}

func (s *Screen) drawMessageBar(buf *bytes.Buffer) {
	BufAppend(buf, "\x1b[K")
	if time.Until(globalState.statustime).Seconds() < 5.00 {
		BufAppend(buf, globalState.statusmsg)
	}
}

func editorScroll() {
	globalState.rx = 0
	if globalState.cy < globalState.nrows {
		globalState.rx = cxToRx(&globalState.trow[globalState.cy], globalState.cx)
	}

	if globalState.cy < globalState.rowoff {
		globalState.rowoff = globalState.cy
	}

	if globalState.cy >= globalState.rowoff+globalState.srows {
		globalState.rowoff = globalState.cy - globalState.srows + 1
	}

	if globalState.rx < globalState.coloff {
		globalState.coloff = globalState.rx
	}

	if globalState.rx >= globalState.coloff+globalState.scols {
		globalState.coloff = globalState.rx - globalState.scols + 1
	}
}

func refreshScreen() {
	s := getTerminalSize()
	editorScroll()

	buf := bytes.NewBufferString("")

	BufAppend(buf, "\x1b[?25l")
	BufAppend(buf, "\x1b[H")

	s.drawRows(buf)
	s.drawStatusBar(buf)
	s.drawMessageBar(buf)

	BufAppend(buf, fmt.Sprintf("\x1b[%d;%dH", (globalState.cy-globalState.rowoff)+1, (globalState.rx-globalState.coloff)+1))
	BufAppend(buf, "\x1b[?25h")

	if _, err := os.Stdout.Write(buf.Bytes()); err != nil {
		safeExit(err)
	}
}
