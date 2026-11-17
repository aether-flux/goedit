package main

import (
	"bufio"
	"io"
	"os"
)

func ctrl(c byte) byte {
	return (c & 31)
}

const (
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
	case ctrl('q'):
		safeExit(nil)
	case 'h', 'j', 'k', 'l', 'G', ctrl('g'), '0', '$':
		moveCursor(k)
	}

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
		if seq[1] >= '0' && seq[1] >= '9' {
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

	return -1
}

func handleSpecialKeys(key int) {
	switch key {
	case ARROW_LEFT:
		if globalState.cx >= 1 {
			globalState.cx -= 1
		}
	case ARROW_DOWN:
		if globalState.cy <= globalState.srows {
			globalState.cy += 1
		}
	case ARROW_UP:
		if globalState.cy >= 1 {
			globalState.cy -= 1
		}
	case ARROW_RIGHT:
		if globalState.cx <= globalState.scols {
			globalState.cx += 1
		}
	case PAGE_DOWN:
		globalState.cy = globalState.srows - 1
	case PAGE_UP:
		globalState.cy = 0
	case HOME:
		globalState.cx = 0
	case END:
		globalState.cx = globalState.scols - 1
	}
}

func moveCursor(k byte) {
	switch k {
	case 'h':
		if globalState.cx >= 1 {
			globalState.cx -= 1
		}
	case 'j':
		if globalState.cy <= globalState.srows {
			globalState.cy += 1
		}
	case 'k':
		if globalState.cy >= 1 {
			globalState.cy -= 1
		}
	case 'l':
		if globalState.cx <= globalState.scols {
			globalState.cx += 1
		}
	case 'G':
		globalState.cy = globalState.srows - 1
	case ctrl('g'):
		globalState.cy = 0
	case '0':
		globalState.cx = 0
	case '$':
		globalState.cx = globalState.scols - 1
	}
}
