package main

import (
	"fmt"
	"strings"

	"github.com/nsf/termbox-go"
)

// A Buffer represents a file
type Buffer struct {
	Rows       []Row
	FileName   string
	X, Y, W, H int
	YOffset    int
}

func NewBuffer() *Buffer {
	b := &Buffer{}
	b.Rows = make([]Row, 0)
	return b
}

func (b *Buffer) ReadBytes(bytes []byte) {
	s := string(bytes)
	lines := strings.Split(s, "\n")
	b.Rows = make([]Row, 0)
	for _, line := range lines {
		b.Rows = append(b.Rows, NewRow(line))
	}
}

func (b *Buffer) Bytes() []byte {
	var s string
	for _, row := range b.Rows {
		s += row.Text + "\n"
	}
	return []byte(s)
}

func (b *Buffer) Render() {
	for y := b.Y; y < b.H; y++ {
		var line string
		if (y + b.X) < len(b.Rows) {
			line = b.Rows[y+b.YOffset].DisplayText()
			if b.X < len(line) {
				line = line[b.X:]
			} else {
				line = ""
			}
		} else {
			line = "~"
			if y == b.H/3 {
				welcome := fmt.Sprintf("the gott editor -- version %s", VERSION)
				padding := (b.W - len(welcome)) / 2
				for i := 1; i <= padding; i++ {
					line = line + " "
				}
				line += welcome
			}
		}
		// truncate line to fit screen
		if len(line) > b.W {
			line = line[0:b.W]
		}
		for x, c := range line {
			termbox.SetCell(x, y, rune(c), termbox.ColorWhite, termbox.ColorBlack)
		}
	}
}
