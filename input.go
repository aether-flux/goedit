package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

const quit_limit = 2

func ctrl(c byte) byte {
	return (c & 31)
}

const (
	BACKSPACE  = 127
	ARROW_LEFT = 1000 + iota
	ARROW_RIGHT
	ARROW_UP
	ARROW_DOWN
	PAGE_UP
	PAGE_DOWN
	HOME
	END
	DELETE
)

type KeyReader struct {
	*bufio.Reader
}

func NewKeyReader() *KeyReader {
	return &KeyReader{
		bufio.NewReader(os.Stdin),
	}
}

func (kr *KeyReader) ReadKey() (byte, error) {
	return kr.ReadByte()
}

var quittimes int = quit_limit - 1

func editorPrompt(prompt string) string {
	var out []byte
	kr := NewKeyReader()

	for {
		setStatusMessage(fmt.Sprintf(prompt, string(out)))
		refreshScreen()

		c, err := kr.ReadKey()
		if err != nil {
			safeExit(err)
		}

		if c == '\r' {
			if len(out) > 0 {
				setStatusMessage("")
				return string(out)
			}
		} else if c == BACKSPACE {
			if len(out) > 0 {
				out = out[:len(out)-1]
			}
		} else if c == '\x1b' {
			setStatusMessage("")
			return ""
		} else if c < 128 {
			out = append(out, c)
		}
	}
}

func processKey(kr *KeyReader) error {
	k, err := kr.ReadKey()

	if err == io.EOF {
		safeExit(nil)
	} else if err != nil {
		safeExit(err)
	}

	switch k {
	case '\x1b':
		key := readSpecialKeys(kr)
		handleSpecialKeys(key)
	case BACKSPACE:
		handleSpecialKeys(int(k))
	case ctrl('q'):
		if globalState.dirty > 0 && quittimes > 0 {
			msg := fmt.Sprintf("WARN: Unsaved changes, press <C-q> %d more times to quit without saving", quittimes)
			setStatusMessage(msg)
			quittimes -= 1
			return nil
		}
		_, _ = fmt.Fprint(os.Stdout, "\x1b[2J")
		// _, _ = fmt.Fprint(os.Stdout, "\x1b[H")
		safeExit(nil)
	// case 'h', 'j', 'k', 'l', 'G', ctrl('g'), '0', '$':
	// 	moveCursor(int(k))
	case ctrl('s'):
		editorSave()
	case '\r':
		insertNewline()
	default:
		insertChar(k)
	}

	var row string
	if globalState.cy >= globalState.nrows {
		row = ""
	} else {
		row = globalState.trow[globalState.cy].str
	}
	if globalState.cx > len(row) {
		globalState.cx = len(row)
	}

	quittimes = quit_limit - 1
	return nil
}

func readSpecialKeys(kr *KeyReader) int {
	var seq [3]byte
	var err error

	if seq[0], err = kr.ReadKey(); err != nil {
		safeExit(nil)
	}
	if seq[1], err = kr.ReadKey(); err != nil {
		safeExit(nil)
	}

	if seq[0] == '[' {
		if seq[1] >= '0' && seq[1] <= '9' {
			if seq[2], err = kr.ReadKey(); err != nil {
				safeExit(nil)
			}

			if seq[2] == '~' {
				switch seq[1] {
				case '5':
					return PAGE_UP
				case '6':
					return PAGE_DOWN
				case '3':
					return DELETE
				case '1', '7':
					return HOME
				case '4', '8':
					return END
				}
			}
		} else {
			switch seq[1] {
			case 'A':
				return ARROW_UP
			case 'B':
				return ARROW_DOWN
			case 'C':
				return ARROW_RIGHT
			case 'D':
				return ARROW_LEFT
			case 'H':
				return HOME
			case 'F':
				return END
			}
		}
	} else if seq[0] == 'O' {
		switch seq[1] {
		case 'H':
			return HOME
		case 'F':
			return END
		}
	}

	return '\x1b'
}

func handleSpecialKeys(key int) {
	switch key {
	case ARROW_LEFT, ARROW_DOWN, ARROW_UP, ARROW_RIGHT, HOME, END:
		moveCursor(key)
	case PAGE_DOWN, PAGE_UP:
		if key == PAGE_UP {
			globalState.cy = globalState.rowoff
		} else if key == PAGE_DOWN {
			globalState.cy = min(globalState.nrows, globalState.rowoff+globalState.srows-1)
		}

		for times := globalState.srows; times >= 0; times -= 1 {
			moveCursor(key)
		}
	case BACKSPACE:
		delChar()
	case DELETE:
		moveCursor(ARROW_RIGHT)
		delChar()
		// case '\x1b':
		// 	TODO
	}
}

func moveCursor(k int) {
	onRow := globalState.cy < globalState.nrows
	row := ""
	if onRow {
		row = globalState.trow[globalState.cy].str
	}

	switch k {
	case 'h', ARROW_LEFT:
		if globalState.cx >= 1 {
			globalState.cx -= 1
		} else if globalState.cy > 0 {
			globalState.cy -= 1
			globalState.cx = len(globalState.trow[globalState.cy].str)
		}
	case 'j', ARROW_DOWN, PAGE_DOWN:
		if globalState.cy < globalState.nrows-1 {
			globalState.cy++
		}
	case 'k', ARROW_UP, PAGE_UP:
		if globalState.cy >= 1 {
			globalState.cy -= 1
		}
	case 'l', ARROW_RIGHT:
		if onRow && globalState.cx < len(row) {
			globalState.cx++
		} else if onRow && globalState.cx == len(row) {
			globalState.cx = 0
			globalState.cy++
		}
	case 'G':
		globalState.cy = globalState.nrows - 1
	case int(ctrl('g')):
		globalState.cy = 0
	case 'L':
		globalState.cy = min(globalState.nrows, globalState.rowoff+globalState.srows-1)
	case 'H':
		globalState.cy = min(globalState.nrows, globalState.rowoff)
	case '0', HOME:
		globalState.cx = 0
	case '$', END:
		globalState.cx = len(row)
	}
}
