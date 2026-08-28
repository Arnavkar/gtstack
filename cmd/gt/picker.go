package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

const pickerPrompt = "Checkout a branch (autocomplete or arrow keys)"

func filterRows(rows []pickRow, q string) []pickRow {
	if q == "" {
		return rows
	}
	ql := strings.ToLower(q)
	var out []pickRow
	for _, r := range rows {
		name := r.branch
		if r.openGithub {
			name = r.text
		}
		if strings.Contains(strings.ToLower(name), ql) {
			out = append(out, r)
		}
	}
	return out
}

func pickBranch(rows []pickRow) (pickRow, error) {
	fd := int(os.Stdin.Fd())
	old, err := term.MakeRaw(fd)
	if err != nil {
		return pickRow{}, err
	}
	defer term.Restore(fd, old)

	query := ""
	visible := rows
	sel := 0
	for i, r := range rows {
		if r.current {
			sel = i
			break
		}
	}

	lastHeight := 0
	clear := func() {
		if lastHeight == 0 {
			return
		}
		fmt.Fprintf(os.Stderr, "\x1b[%dA\x1b[0J", lastHeight)
	}
	defer func() {
		clear()
		fmt.Fprint(os.Stderr, "\x1b[?25h")
	}()

	draw := func() {
		clear()
		var b strings.Builder
		b.WriteString("\r\x1b[?25l\x1b[2K")
		if useColor() {
			b.WriteString("\x1b[36m?\x1b[0m ")
			b.WriteString(pickerPrompt)
			b.WriteString(" \x1b[2m›\x1b[0m ")
		} else {
			b.WriteString("? ")
			b.WriteString(pickerPrompt)
			b.WriteString(" › ")
		}
		b.WriteString(query)
		b.WriteString("\r\n")
		for i, r := range visible {
			b.WriteString("\x1b[2K")
			if i == sel {
				if useColor() {
					b.WriteString("\x1b[32m❯\x1b[0m ")
				} else {
					b.WriteString("> ")
				}
			} else {
				b.WriteString("  ")
			}
			b.WriteString(r.text)
			b.WriteString("\r\n")
		}
		fmt.Fprint(os.Stderr, b.String())
		lastHeight = 1 + len(visible)
	}

	draw()

	buf := make([]byte, 8)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			return pickRow{}, fmt.Errorf("checkout cancelled")
		}
		refit := false
		move := 0
		switch {
		case buf[0] == 3, buf[0] == 27 && n == 1:
			return pickRow{}, fmt.Errorf("checkout cancelled")
		case buf[0] == '\r', buf[0] == '\n':
			if len(visible) == 0 {
				continue
			}
			return visible[sel], nil
		case buf[0] == 127, buf[0] == 8:
			query = trimLastRune(query)
			refit = true
		case buf[0] == 14:
			move = 1
		case buf[0] == 16:
			move = -1
		case n >= 3 && buf[0] == 27 && buf[1] == '[':
			switch buf[2] {
			case 'A':
				move = -1
			case 'B':
				move = 1
			}
		case n == 1 && buf[0] >= 32 && buf[0] < 127:
			query += string(buf[0])
			refit = true
		default:
			continue
		}
		if refit {
			visible = filterRows(rows, query)
			if sel >= len(visible) {
				sel = max(0, len(visible)-1)
			}
		}
		if move != 0 && len(visible) > 0 {
			sel += move
			if sel < 0 {
				sel = 0
			}
			if sel >= len(visible) {
				sel = len(visible) - 1
			}
		}
		draw()
	}
}

func trimLastRune(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return string(r[:len(r)-1])
}
