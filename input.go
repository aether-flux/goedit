package main

import (
	"bufio"
	"io"
	"os"
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
	case ctrl('q'):
		safeExit(nil)
	}

	return nil
}

func ctrl(c byte) byte {
	return (c & 31)
}
