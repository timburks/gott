package main

import (
	"strings"
)

// A Buffer represents a file
type Buffer struct {
	Rows     []Row
	FileName string
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
