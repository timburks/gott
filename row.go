package main

import "strings"

// A row of text in the editor
type Row struct {
	Text string
}

// We replace any tabs with spaces
func NewRow(text string) Row {
	r := Row{}
	r.Text = strings.Replace(text, "\t", "        ", -1)
	return r
}

func (r *Row) DisplayText() string {
	return r.Text
}

func (r *Row) Length() int {
	return len(r.DisplayText())
}

func (r *Row) InsertChar(position int, c rune) {
	line := ""
	if position <= len(r.Text) {
		line += r.Text[0:position]
	} else {
		line += r.Text
	}
	line += string(c)
	if position < len(r.Text) {
		line += r.Text[position:]
	}
	r.Text = line
}

func (r *Row) ReplaceChar(position int, c rune) {
	if (position < 0) || (position >= len(r.Text)) {
		return
	}
	r.Text = r.Text[0:position] + string(c) + r.Text[position+1:]
}

// delete character at position and return the deleted character
func (r *Row) DeleteChar(position int) rune {
	if r.Length() == 0 {
		return 0
	}
	if position > r.Length()-1 {
		position = r.Length() - 1
	}
	ch := rune(r.Text[position])
	r.Text = r.Text[0:position] + r.Text[position+1:]
	return ch
}

// split row at position, return a new row containing the remaining text.
func (r *Row) Split(position int) Row {
	before := r.Text[0:position]
	after := r.Text[position:]
	r.Text = before
	return NewRow(after)
}
