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
	"unicode"

	gott "github.com/timburks/gott/types"
)

var lastBufferNumber = -1

// A Buffer represents a file being edited

type Buffer struct {
	number      int
	Name        string
	rows        []*Row
	fileName    string
	mode        string
	Highlighted bool
	ReadOnly    bool
}

func NewBuffer() *Buffer {
	lastBufferNumber++
	b := &Buffer{}
	b.number = lastBufferNumber
	b.rows = make([]*Row, 0)
	b.Highlighted = false
	return b
}

func (b *Buffer) GetIndex() int {
	return b.number
}

func (b *Buffer) GetName() string {
	return b.Name
}

func (b *Buffer) GetFileName() string {
	return b.fileName
}

func (b *Buffer) GetReadOnly() bool {
	return b.ReadOnly
}

func (b *Buffer) SetFileName(name string) {
	b.fileName = name
	if strings.HasSuffix(name, ".go") {
		b.mode = "go"
	} else {
		b.mode = "txt"
	}
	b.Name = name
}

func (b *Buffer) LoadBytes(bytes []byte) {
	s := string(bytes)
	lines := strings.Split(s, "\n")
	b.rows = make([]*Row, 0)
	for _, line := range lines {
		b.rows = append(b.rows, NewRow(line))
	}
	b.Highlighted = false
}

func (b *Buffer) Bytes() []byte {
	var s string
	for i, row := range b.rows {
		if i > 0 {
			s += "\n"
		}
		s += string(row.Text)
	}
	return []byte(s)
}

func (b *Buffer) GetRowCount() int {
	return len(b.rows)
}

func (b *Buffer) GetRowLength(i int) int {
	if i < len(b.rows) {
		return b.rows[i].Length()
	} else {
		return 0
	}
}

func (b *Buffer) GetCharacterAtCursor(cursor gott.Point) rune {
	if cursor.Row < len(b.rows) {
		row := b.rows[cursor.Row]
		if cursor.Col < row.Length() && cursor.Col >= 0 {
			return row.Text[cursor.Col]
		}
	}
	return rune(0)
}

func (b *Buffer) TextAfter(row, col int) string {
	if row < len(b.rows) {
		return b.rows[row].TextAfter(col)
	} else {
		return ""
	}
}

func (b *Buffer) InsertCharacter(row, col int, c rune) {
	b.Highlighted = false
	if row < len(b.rows) {
		b.rows[row].InsertChar(col, c)
	}
}

func (b *Buffer) DeleteRow(row int) {
	b.Highlighted = false
	if row < len(b.rows) {
		b.rows = append(b.rows[0:row], b.rows[row+1:]...)
	}
}

func (b *Buffer) DeleteCharacters(row int, col int, count int, joinLines bool) string {
	b.Highlighted = false
	deletedText := ""
	if b.GetRowCount() == 0 {
		return deletedText
	}
	for i := 0; i < count; i++ {
		if col < b.rows[row].Length() {
			c := b.rows[row].DeleteChar(col)
			deletedText += string(c)
		} else if joinLines && row < b.GetRowCount()-1 {
			// join next row to current row
			nextRow := b.rows[row+1]
			b.rows[row].Join(nextRow)
			// remove next row
			b.DeleteRow(row + 1)
			deletedText += "\n"
		}
	}
	return deletedText
}

func checkalphanum(line string, start, end int) bool {
	if start > 0 {
		c := rune(line[start-1])
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			return true
		}
	}
	if end < len(line) {
		c := rune(line[end])
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			return true
		}
	}
	return false
}

// draw text in an area defined by origin and size with a specified offset into the buffer
func (b *Buffer) Render(origin gott.Point, size gott.Size, offset gott.Size, display gott.Display) {

	if !b.Highlighted {
		switch b.mode {
		case "go":
			h := NewGoHighlighter()
			h.Highlight(b)
		}
		b.Highlighted = true
	}

	for i := origin.Row; i < origin.Row+size.Rows; i++ {
		var line string
		var colors []gott.Color
		if (i + offset.Rows) < len(b.rows) {

			line = b.rows[i+offset.Rows].DisplayText()
			colors = b.rows[i+offset.Rows].Colors

			if offset.Cols < len(line) {
				line = line[offset.Cols:]
				colors = colors[offset.Cols:]
			} else {
				line = ""
			}
		} else {
			line = "~"
			colors = make([]gott.Color, 1, 1)
			colors[0] = gott.ColorWhite
		}
		// truncate line to fit screen
		if len(line) > size.Cols {
			line = line[0:size.Cols]
			colors = colors[0:size.Cols]
		}

		for j, c := range line {
			var color gott.Color = gott.ColorWhite
			if j < len(colors) {
				color = colors[j]
			}
			display.SetCell(j+origin.Col, i, rune(c), color)
		}
	}
}
