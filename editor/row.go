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
package editor

import (
	"strings"

	gott "github.com/timburks/gott/types"
)

// A row of text in the editor
type Row struct {
	Text   []rune
	Colors []gott.Color
}

// We replace any tabs with spaces
func NewRow(text string) *Row {
	r := &Row{}
	r.setText([]rune(strings.Replace(text, "\t", "        ", -1)))
	return r
}

func (r *Row) setText(text []rune) {
	r.Text = text
	r.Colors = make([]gott.Color, len(r.Text), len(r.Text))
	for j, _ := range r.Colors {
		r.Colors[j] = 0xff
	}
}

func (r *Row) DisplayText() string {
	return string(r.Text)
}

func (r *Row) Length() int {
	return len(r.Text)
}

func (r *Row) InsertChar(col int, c rune) {
	line := make([]rune, 0)
	if col <= len(r.Text) {
		line = append(line, r.Text[0:col]...)
	} else {
		line = append(line, r.Text...)
	}
	line = append(line, c)
	if col < len(r.Text) {
		line = append(line, r.Text[col:]...)
	}
	r.setText(line)
}

// replace character at col and return the replaced character
func (r *Row) ReplaceChar(col int, c rune) rune {
	if (col < 0) || (col >= len(r.Text)) {
		return rune(0)
	}
	result := rune(r.Text[col])
	r.Text[col] = c
	return result
}

// delete character at col and return the deleted character
func (r *Row) DeleteChar(col int) rune {
	if len(r.Text) == 0 {
		return 0
	}
	if col > len(r.Text)-1 {
		col = len(r.Text) - 1
	}
	c := rune(r.Text[col])
	r.setText(append(r.Text[0:col], r.Text[col+1:]...))
	return c
}

// splits row at col, return a new row containing the remaining text.
func (r *Row) Split(col int) *Row {
	if col < len(r.Text) {
		after := r.Text[col:]
		r.setText(r.Text[0:col])
		return NewRow(string(after))
	} else {
		return NewRow("")
	}
}

// joins rows by appending the passed-in row to the current row
func (r *Row) Join(other *Row) {
	r.setText(append(r.Text, other.Text...))
}

// returns the text after a specified column
func (r *Row) TextAfter(col int) string {
	if col < len(r.Text) {
		return string(r.Text[col:])
	} else {
		return ""
	}
}
