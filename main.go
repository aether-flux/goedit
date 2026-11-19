package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"golang.org/x/sys/unix"
)

type erow struct {
	str string
	// render string
}

var globalState = struct {
	cx, cy          int
	rx              int
	srows, scols    int
	trow            []erow
	filename        string
	statusmsg       string
	statustime      time.Time
	nrows           int
	rowoff, coloff  int
	restoreTerminal func()
	orig_termios    unix.Termios
}{
	0, 0,
	0,
	0, 0,
	[]erow{},
	"",
	"HELP: CTRL-Q = Quit",
	time.Now(),
	0,
	0, 0,
	nil,
	unix.Termios{},
}

func BufAppend(buf *bytes.Buffer, str string) {
	_, _ = buf.WriteString(str)
}

func safeExit(err error) {
	// _, _ = fmt.Fprint(os.Stdout, "\x1b[2J")
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

	fmt.Fprint(os.Stdout, "\x1b[?7l")

	return func() {
		fmt.Fprint(os.Stdout, "\x1b[?7h")
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

	defer func() {
		_, _ = fmt.Fprint(os.Stdout, "\x1b[H")

		if r := recover(); r != nil {
			if globalState.restoreTerminal != nil {
				globalState.restoreTerminal()
			}

			fmt.Fprintf(os.Stderr, "panic: %v\r\n%s\r\n", r, debug.Stack())
			os.Exit(1)
		}
	}()

	editorOpen(filename)
	kr := NewKeyReader()

	for {
		refreshScreen()

		if err := processKey(kr); err != nil {
			safeExit(err)
		}
	}
}
