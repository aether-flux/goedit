package main

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

var globalState = struct {
	cx, cy          int
	srows, scols    int
	trow            []string
	nrows, rowoff   int
	restoreTerminal func()
	orig_termios    unix.Termios
}{
	0, 0,
	0, 0,
	[]string{},
	0, 0,
	nil,
	unix.Termios{},
}

func BufAppend(buf *bytes.Buffer, str string) {
	_, _ = buf.WriteString(str)
}

func safeExit(err error) {
	_, _ = fmt.Fprint(os.Stdout, "\x1b[2J")
	_, _ = fmt.Fprint(os.Stdout, "\x1b[H")

	if globalState.restoreTerminal != nil {
		globalState.restoreTerminal()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\r\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func rawMode() (func(), error) {
	raw, err := unix.IoctlGetTermios(unix.Stdin, unix.TCGETS)
	if err != nil {
		return nil, fmt.Errorf("rawMode: error getting terminal flags: %w", err)
	}

	globalState.orig_termios = *raw
	raw.Oflag &^= unix.OPOST
	raw.Iflag &^= unix.IXON | unix.ICRNL | unix.BRKINT | unix.INPCK | unix.ISTRIP
	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.ISIG | unix.IEXTEN
	raw.Cflag |= unix.CS8
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 1

	if err = unix.IoctlSetTermios(unix.Stdin, unix.TCSETS, raw); err != nil {
		return nil, fmt.Errorf("rawMode: error setting terminal flags: %w", err)
	}

	return func() {
		if err := unix.IoctlSetTermios(unix.Stdin, unix.TCSETS, &globalState.orig_termios); err != nil {
			fmt.Fprintf(os.Stderr, "disableRawMode: error restoring original terminal flags: %s\r\n", err)
		}
	}, nil
}

func main() {
	disableRawMode, err := rawMode()
	if err != nil {
		safeExit(err)
	}
	globalState.restoreTerminal = disableRawMode
	filename := ""
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}
	editorOpen(filename)
	defer safeExit(nil)

	kr := NewKeyReader()

	for {
		s := getTerminalSize()
		refreshScreen(s)

		if err := processKey(kr); err != nil {
			safeExit(err)
		}
	}
}
