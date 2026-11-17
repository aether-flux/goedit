package main

import (
	"os"
	"strings"
)

func editorOpen(filename string) {
	if len(filename) == 0 {
		return
	} else {
		data, err := os.ReadFile(filename)
		if err != nil {
			safeExit(err)
		}

		lines := strings.Split(string(data[:]), "\n")
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}

		globalState.trow = lines
		globalState.nrows = len(lines)
	}
}
