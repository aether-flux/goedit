package main

import (
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
	return &Screen{rows, cols}
}

func (s *Screen) drawRows() {
	for y := range s.Rows {
		_, _ = fmt.Fprintf(os.Stdout, "~")

		if y < s.Rows-1 {
			_, _ = fmt.Fprintf(os.Stdout, "\r\n")
		}
	}
}

func refreshScreen(s *Screen) {
	fmt.Fprint(os.Stdout, "\x1b[2J")
	fmt.Fprint(os.Stdout, "\x1b[H")

	s.drawRows()

	fmt.Fprint(os.Stdout, "\x1b[H")
}
