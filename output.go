package main

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

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
	globalState.srows = rows
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
			BufAppend(buf, globalState.trow[filerow])
		}

		BufAppend(buf, "\x1b[K")
		if y < s.Rows-1 {
			BufAppend(buf, "\r\n")
		}
	}
}

func editorScroll() {
	if globalState.cy < globalState.rowoff {
		globalState.rowoff = globalState.cy
	}

	if globalState.cy >= globalState.rowoff+globalState.srows {
		globalState.rowoff = globalState.cy - globalState.srows + 1
	}
}

func refreshScreen(s *Screen) {
	editorScroll()

	buf := bytes.NewBufferString("")

	BufAppend(buf, "\x1b[?25l")
	BufAppend(buf, "\x1b[H")

	s.drawRows(buf)

	BufAppend(buf, fmt.Sprintf("\x1b[%d;%dH", (globalState.cy-globalState.rowoff)+1, globalState.cx+1))
	// BufAppend(buf, "\x1b[H")
	BufAppend(buf, "\x1b[?25h")

	if _, err := os.Stdout.Write(buf.Bytes()); err != nil {
		safeExit(err)
	}
}
