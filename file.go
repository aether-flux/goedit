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

		globalState.filename = filename
		globalState.trow = make([]erow, 0, len(lines))

		globalState.trow = append(globalState.trow, erow{str: ""})
		globalState.trow = append(globalState.trow, erow{str: ""})

		for _, line := range lines {
			line = strings.TrimRight(line, "\r")
			globalState.trow = append(globalState.trow, erow{str: line})
		}
		globalState.nrows = len(lines) + 2
	}
}
