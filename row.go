//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package main

import "strings"

// A row of text in the editor
type Row struct {
	Text []rune
}

// We replace any tabs with spaces
func NewRow(text string) Row {
	r := Row{}
	r.Text = []rune(strings.Replace(text, "\t", "        ", -1))
	return r
}

func (r *Row) DisplayText() string {
	return string(r.Text)
}

func (r *Row) Length() int {
	return len(r.Text)
}

func (r *Row) InsertChar(position int, c rune) {
	line := make([]rune, 0)
	if position <= len(r.Text) {
		line = append(line, r.Text[0:position]...)
	} else {
		line = append(line, r.Text...)
	}
	line = append(line, c)
	if position < len(r.Text) {
		line = append(line, r.Text[position:]...)
	}
	r.Text = line
}

// replace character at position and return the replaced character
func (r *Row) ReplaceChar(position int, c rune) rune {
	if (position < 0) || (position >= len(r.Text)) {
		return rune(0)
	}
	result := rune(r.Text[position])
	r.Text[position] = c
	return result
}

// delete character at position and return the deleted character
func (r *Row) DeleteChar(position int) rune {
	if len(r.Text) == 0 {
		return 0
	}
	if position > len(r.Text)-1 {
		position = len(r.Text) - 1
	}
	c := rune(r.Text[position])
	r.Text = append(r.Text[0:position], r.Text[position+1:]...)
	return c
}

// split row at position, return a new row containing the remaining text.
func (r *Row) Split(position int) Row {
	if position < len(r.Text) {
		after := r.Text[position:]
		r.Text = r.Text[0:position]
		return NewRow(string(after))
	} else {
		return NewRow("")
	}
}
