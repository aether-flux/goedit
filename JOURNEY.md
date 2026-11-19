# 1. Setup
# 2. Entering Raw Mode (Reading user keypresses)

By default, the terminal opens in canonical (also called cooked) mode. To make it truly raw mode, we would need much more flags.

For now however, we can already track user keypresses, although it requires a delimiter (currently set to `\n`) which marks what the pressed key is, thus enabling the code to also register multiple keys together.

The code snippet is:

```go
package main

import (
  "fmt"
  "bufio"
  "os"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		s, err := reader.ReadByte()

		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Reading key from stdin: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("Pressed: ", s)
	}
}
```

**Output:**

```sh
a
Pressed:  97
Pressed:  10
g
Pressed:  103
Pressed:  10
sf
Pressed:  115
Pressed:  102
Pressed:  10

^Csignal: interrupt
```

- `Ctrl-C` or `Ctrl-D` will stop/terminate the program.
- `bufio` creates the NewReader using the Stdin from module `os`.

## Exit when pressing `q`

To implement this feature, all we have to do is check if the captured string (s) is 'q', and if yes, break the loop.

```go
  for {
		s, err := reader.ReadByte()

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
```

**Output:**

```sh
l
Pressed:  108
Pressed:  10
t
Pressed:  116
Pressed:  10
q
(...exited...)
```

## Disable ECHO

The ECHO flag enables terminal to display the characters typed in output. We can disable this with the following function:

```go
func rawMode() (func(), error) {
	raw, err := unix.IoctlGetTermios(unix.Stdin, unix.TCGETS)
	if err != nil {
		return nil, fmt.Errorf("rawMode: error getting terminal flags: %w", err)
	}

	copy := raw
	raw.Lflag &^= unix.ECHO

	if err = unix.IoctlSetTermios(unix.Stdin, unix.TCSETS, raw); err != nil {
		return nil, fmt.Errorf("rawMode: error setting terminal flags: %w", err)
	}

	return func() {
		if err := unix.IoctlSetTermios(unix.Stdin, unix.TCSETS, copy); err != nil {
			fmt.Fprintf(os.Stderr, "disableRawMode: error restoring original terminal flags: %s", err)
		}
	}, nil
}
```

This function gets the attributes of terminal, sets the `c_lflag`/`Lflag` to remove the `ECHO` flag, and then sets the new set of flags as the attributes.
This function returns a function and an error. The returned function can be called to reset the terminal attributes to the original set, thus restoring the attributes.

This function is used in the main function as follows:

```go
func main() {
	disableRawMode, err := rawMode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "main: error setting terminal to raw mode")
		os.Exit(1)
	}
	reader := bufio.NewReader(os.Stdin)

	for {
		s, err := reader.ReadByte()
		...
	}
	disableRawMode()
}
```

**Output:**

```sh
Pressed:  97
Pressed:  10
Pressed:  103
Pressed:  10
Pressed:  114
Pressed:  10
```

## Disabling Canonical Mode

Now its time to disable canonical mode. In this mode, every character we input needs to be passed with a line delimiter (like `\n`) to be sent to the program's read() loop (here, bufio's reader). In non-canonical mode, this delimiter is no longer required, and keypresses register instantly.

To disable canonical mode, we need to disable the ICANON flag along with ECHO:

```go
func rawMode() (func(), error) {
	...

	raw.Lflag &^= unix.ECHO | unix.ICANON

	...
}
```

## Disabling signals

Another `c_lflag`/`Lflag` is `ISIG`, which converts characters like `Ctrl-C`/`Ctrl-Z` into corresponding signals like `INTR`, `QUIT`, `SUSP`, or `DSUSP`. We can disable this in a similar manner:

```go
func rawMode() (func(), error) {
	...

	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.ISIG

	...
}
```

## Disabling `Ctrl-S` and `Ctrl-Q`

By default, these keys are used for software flow control. `Ctrl-S` stops data from being transmitted to the terminal until `Ctrl-Q` is pressed.
We can turn off this feature by disabling the flag `IXON`, which is an input flag (`c_iflag`/`Iflag`).

```go
func rawMode() (func(), error) {
	...

	raw.Iflag &^= unix.IXON
	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.ISIG

	...
}
```

## Disabling `Ctrl-V`

`Ctrl-V` makes the terminal wait for another character to be typed that it sends literally. For example, before we disabled `Ctrl-C`, typing `Ctrl-V` then `Ctrl-C` would have sent the byte `3` literally instead of sending an interrupt signal.
This can be disabled via the `IEXTEN` flag, which is another `Lflag`.

```go
func rawMode() (func(), error) {
	...

	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.ISIG | unix.IEXTEN

	...
}
```

## Fixing `Ctrl-M`

If you run the program now and go through the whole alphabet while holding down Ctrl, you should see that we have every letter except M. Ctrl-M is weird: it’s being read as 10, when we expect it to be read as 13, since it is the 13th letter of the alphabet, and Ctrl-J already produces a 10. What else produces 10? The Enter key does.

It turns out that the terminal is helpfully translating any carriage returns (13, '\r') inputted by the user into newlines (10, '\n'). This can be disabled via the ICRNL flag.

```go
func rawMode() (func(), error) {
	...

	raw.Iflag &^= unix.IXON | unix.ICRNL
	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.ISIG | unix.IEXTEN

	...
}
```

## Disabling Output Processing

To turn off all output processing, we disable the `Oflag` `OPOST` in a similar way we disabled `Lflag`s and `Iflag`s.

This would disable the terminal from converting every `\n` to `\r\n`, so we would only notice the cursor going down in a newline, not to the left of the screen:

```sh
Pressed:  115
             Pressed:  100
                          Pressed:  102
```

To fix this, add `\r` before all print statements that have a `\n` character. For the line `fmt.Println("Pressed:", s)`, we can replace it with the more flexible Printf method as `fmt.Printf("Pressed: %c\r\n", s)`.

## Some more flags

```go
func rawMode() (func(), error) {
	...

	raw.Oflag &^= unix.OPOST
	raw.Iflag &^= unix.IXON | unix.ICRNL | unix.BRKINT | unix.INPCK | unix.ISTRIP
	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.ISIG | unix.IEXTEN
	raw.Cflag |= unix.CS8

	...
}
```

- `BRKINT`: When enabled, a break condition generates a SIGINT signal (similar to pressing `Ctrl-C`)
- `INPCK`: Parity checking, doesn't seem to apply to modern terminal emulators
- `ISTRIP`: Causes 8th bit of each input byte to be stripped. Probably already turned off
- `CS8`: A bit mask (not a flag), sets character size (CS) to 8 bits per byte. Its already set that way on my system

## Read with timeout

In `man 3 termios`, there are four cases of different values of `c_cc` values for non canonical mode. The one we are concerned with is:

- `VMIN` = 0: Minimum number of characters required to be read is 0
- `VTIME` > 0: Timeout until when `read` should wait to read characters is greater than 0

This combination gives a timeout to our reader which, after waiting for `VTIME` deciseconds, will return and break the loop if no characters were pressed.

You can try it out with the following code:

```go
func rawMode() (func(), error) {
	...

	raw.Cflag |= unix.CS8
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 1

	...
}
```

- `VMIN = 1`: Waits for our reader to read at least one byte/character.
- `VTIME = 1`: A value of 0 would mean no timeout. Setting it as 1 lets the reader return as soon as there is a character to be read.
