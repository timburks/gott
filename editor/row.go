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
	text   []rune
	colors []gott.Color
}

// We replace any tabs with spaces
func NewRow(text string) *Row {
	r := &Row{}
	r.setText([]rune(strings.Replace(text, "\t", "        ", -1)))
	return r
}

func (r *Row) GetText() []rune {
	return r.text
}

func (r *Row) GetColors() []gott.Color {
	return r.colors
}

func (r *Row) SetText(text []rune) {
	r.setText(text)
}

func (r *Row) setText(text []rune) {
	r.text = text
	r.colors = make([]gott.Color, len(r.text), len(r.text))
	for j, _ := range r.colors {
		r.colors[j] = 0xff
	}
}

func (r *Row) DisplayText() string {
	return string(r.text)
}

func (r *Row) Length() int {
	return len(r.text)
}

func (r *Row) InsertChar(col int, c rune) {
	line := make([]rune, 0)
	if col <= len(r.text) {
		line = append(line, r.text[0:col]...)
	} else {
		line = append(line, r.text...)
	}
	line = append(line, c)
	if col < len(r.text) {
		line = append(line, r.text[col:]...)
	}
	r.setText(line)
}

// replace character at col and return the replaced character
func (r *Row) ReplaceChar(col int, c rune) rune {
	if (col < 0) || (col >= len(r.text)) {
		return rune(0)
	}
	result := rune(r.text[col])
	r.text[col] = c
	return result
}

// delete character at col and return the deleted character
func (r *Row) DeleteChar(col int) rune {
	if len(r.text) == 0 {
		return 0
	}
	if col > len(r.text)-1 {
		col = len(r.text) - 1
	}
	c := rune(r.text[col])
	r.setText(append(r.text[0:col], r.text[col+1:]...))
	return c
}

// splits row at col, return a new row containing the remaining text.
func (r *Row) Split(col int) *Row {
	if col < len(r.text) {
		after := r.text[col:]
		r.setText(r.text[0:col])
		return NewRow(string(after))
	} else {
		return NewRow("")
	}
}

// joins rows by appending the passed-in row to the current row
func (r *Row) Join(other *Row) {
	r.setText(append(r.text, other.text...))
}

// returns the text after a specified column
func (r *Row) TextAfter(col int) string {
	if col < len(r.text) {
		return string(r.text[col:])
	} else {
		return ""
	}
}
