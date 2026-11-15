package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

func main() {
	disableRawMode, err := rawMode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "main: error setting terminal to raw mode")
		os.Exit(1)
	}
	reader := bufio.NewReader(os.Stdin)

	// for {
	// 	s, _ := reader.ReadString('\n')
	// 	s = strings.TrimSpace(s)
	// 	if s == "q" {
	// 		break
	// 	}
	// 	fmt.Println("Pressed:", s)
	// }

	for {
		s, err := reader.ReadByte()
		// _, err = reader.ReadByte()

		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Reading key from stdin: %s\n", err)
			os.Exit(1)
		}

		if s == 'q' {
			break
		}
		fmt.Println("Pressed: ", s)
	}

	disableRawMode()
}

func rawMode() (func(), error) {
	raw, err := unix.IoctlGetTermios(unix.Stdin, unix.TCGETS)
	if err != nil {
		return nil, fmt.Errorf("rawMode: error getting terminal flags: %w", err)
	}

	copy := raw
	raw.Lflag = raw.Lflag &^ (unix.ECHO | unix.ICANON)

	if err = unix.IoctlSetTermios(unix.Stdin, unix.TCSETS, raw); err != nil {
		return nil, fmt.Errorf("rawMode: error setting terminal flags: %w", err)
	}

	return func() {
		if err := unix.IoctlSetTermios(unix.Stdin, unix.TCSETA, copy); err != nil {
			fmt.Fprintf(os.Stderr, "disableRawMode: error restoring original terminal flags: %s", err)
		}
	}, nil
}
