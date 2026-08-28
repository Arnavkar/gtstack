package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

const pickerHint = "↑↓/jk  enter checkout  q abort"

func pickBranch(rows []pickRow) (pickRow, error) {
	fd := int(os.Stdin.Fd())
	old, err := term.MakeRaw(fd)
	if err != nil {
		return pickRow{}, err
	}
	defer term.Restore(fd, old)

	sel := 0
	for i, r := range rows {
		if strings.Contains(r.text, "(current)") {
			sel = i
			break
		}
	}

	nLines := len(rows) + 1
	draw := func() {
		var b strings.Builder
		b.WriteString("\r\x1b[?25l")
		for i, r := range rows {
			b.WriteString("\x1b[2K")
			if i == sel {
				if useColor() {
					b.WriteString("\x1b[7m")
					b.WriteString(r.text)
					b.WriteString("\x1b[0m")
				} else {
					b.WriteString("> ")
					b.WriteString(r.text)
				}
			} else if useColor() {
				b.WriteString(r.text)
			} else {
				b.WriteString("  ")
				b.WriteString(r.text)
			}
			b.WriteString("\r\n")
		}
		b.WriteString("\x1b[2K")
		if useColor() {
			b.WriteString("\x1b[2m")
			b.WriteString(pickerHint)
			b.WriteString("\x1b[0m")
		} else {
			b.WriteString(pickerHint)
		}
		b.WriteString("\r\n")
		fmt.Fprint(os.Stderr, b.String())
	}

	draw()
	defer fmt.Fprintf(os.Stderr, "\x1b[%dA\x1b[0J\x1b[?25h", nLines)

	buf := make([]byte, 8)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			return pickRow{}, fmt.Errorf("checkout cancelled")
		}
		switch {
		case buf[0] == 3, buf[0] == 'q', buf[0] == 'Q':
			return pickRow{}, fmt.Errorf("checkout cancelled")
		case buf[0] == '\r', buf[0] == '\n':
			return rows[sel], nil
		case buf[0] == 'j', buf[0] == 14:
			sel++
		case buf[0] == 'k', buf[0] == 16:
			sel--
		case n >= 3 && buf[0] == 27 && buf[1] == '[':
			switch buf[2] {
			case 'A':
				sel--
			case 'B':
				sel++
			}
		}
		if sel < 0 {
			sel = 0
		}
		if sel >= len(rows) {
			sel = len(rows) - 1
		}
		fmt.Fprintf(os.Stderr, "\x1b[%dA", nLines)
		draw()
	}
}
