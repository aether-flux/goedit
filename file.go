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
		// for i, line := range lines {
		// 	lines[i] = strings.TrimRight(line, "\r")
		// }

		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}

		// globalState.trow = lines
		for i, line := range lines {
			line = strings.TrimRight(line, "\r")
			globalState.trow[i].str = line
		}
		globalState.nrows = len(lines)
	}
}
