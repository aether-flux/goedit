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
		s, _ := reader.ReadString('\n')
		fmt.Println("Pressed: ", s)
	}
}
```

**Output:**

```sh
a
Pressed:  a

d
Pressed:  d

qr
Pressed:  qr

^Csignal: interrupt
```

- `Ctrl-C` or `Ctrl-D` will stop/terminate the program.
- `bufio` creates the NewReader using the Stdin from module `os`.

## Exit when pressing `q`

To implement this feature, all we have to do is check if the captured string (s) is 'q', and if yes, break the loop.

```go
  for {
		s, _ := reader.ReadString('\n')
		s = strings.TrimSpace(s)
		if s == "q" {
			break
		}
		fmt.Println("Pressed: ", s)
	}
```

**Output:**

```sh
b
Pressed:  b
da
Pressed:  da
q
(...exited...)
```
