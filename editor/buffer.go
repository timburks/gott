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
	"fmt"
	"strings"
	"unicode"

	gott "github.com/timburks/gott/types"
)

var lastBufferNumber = -1

// A Buffer represents a file being edited
type Buffer struct {
	origin       gott.Point
	size         gott.Size
	number       int
	Name         string
	rows         []*Row
	fileName     string
	languageMode string
	Highlighted  bool
	ReadOnly     bool
	cursor       gott.Point // cursor position
	offset       gott.Size  // display offset
}

// A Window is a view of a buffer.
type Window struct {
	buffer *Buffer
	origin gott.Point
	size   gott.Size
	number int
	cursor gott.Point // cursor position
	offset gott.Size  // display offset
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
		b.languageMode = "go"
	} else {
		b.languageMode = "txt"
	}
	b.Name = name
}

func (b *Buffer) LoadBytes(bytes []byte) []byte {
	previous := b.Bytes()
	s := string(bytes)
	lines := strings.Split(s, "\n")
	b.rows = make([]*Row, 0)
	for _, line := range lines {
		b.rows = append(b.rows, NewRow(line))
	}
	b.Highlighted = false
	return previous
}

func (b *Buffer) AppendBytes(bytes []byte) {
	s := string(bytes)
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		b.rows = append(b.rows, NewRow(line))
	}
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
func (b *Buffer) Render(display gott.Display) {

	b.adjustDisplayOffsetForScrolling()

	if !b.Highlighted {
		switch b.languageMode {
		case "go":
			h := NewGoHighlighter()
			h.Highlight(b)
		}
		b.Highlighted = true
	}

	for i := b.origin.Row; i < b.origin.Row+b.size.Rows-1; i++ {
		var line string
		var colors []gott.Color
		if (i + b.offset.Rows) < len(b.rows) {
			line = b.rows[i+b.offset.Rows].DisplayText()
			colors = b.rows[i+b.offset.Rows].Colors
			if b.offset.Cols < len(line) {
				line = line[b.offset.Cols:]
				colors = colors[b.offset.Cols:]
			} else {
				line = ""
			}
		} else {
			line = "~"
			colors = make([]gott.Color, 1, 1)
			colors[0] = gott.ColorWhite
		}
		// truncate line to fit screen
		if len(line) > b.size.Cols {
			line = line[0:b.size.Cols]
			colors = colors[0:b.size.Cols]
		}
		for j, c := range line {
			var color gott.Color = gott.ColorWhite
			if j < len(colors) {
				color = colors[j]
			}
			display.SetCell(j+b.origin.Col, i, rune(c), color)
		}
	}

	// Draw the info bar as a single line at the bottom of the buffer window.
	infoText := b.getInfoBarText(b.size.Cols)
	infoRow := b.origin.Row + b.size.Rows - 1
	for x, ch := range infoText {
		display.SetCellReversed(x, infoRow, rune(ch), gott.ColorBlack)
	}
}

func (b *Buffer) getInfoBarText(length int) string {
	finalText := fmt.Sprintf(" %d/%d ", b.cursor.Row, b.GetRowCount())
	text := fmt.Sprintf(" [%d] %s", b.GetIndex(), b.GetName())
	if b.GetReadOnly() {
		text = text + "(read-only)"
	}
	for len(text) < length-len(finalText)-1 {
		text = text + " "
	}
	text += finalText
	return text
}

// Recompute the display offset to keep the cursor onscreen.
func (b *Buffer) adjustDisplayOffsetForScrolling() {
	if b.cursor.Row < b.offset.Rows {
		b.offset.Rows = b.cursor.Row
	}
	textRows := b.size.Rows - 1 // save the last row for the info bar
	if b.cursor.Row-b.offset.Rows >= textRows {
		b.offset.Rows = b.cursor.Row - textRows + 1
	}
	if b.cursor.Col < b.offset.Cols {
		b.offset.Cols = b.cursor.Col
	}
	if b.cursor.Col-b.offset.Cols >= b.size.Cols {
		b.offset.Cols = b.cursor.Col - b.size.Cols + 1
	}
}

func (b *Buffer) SetCursor(d gott.Display) {
	d.SetCursor(gott.Point{
		Col: b.cursor.Col - b.offset.Cols,
		Row: b.cursor.Row - b.offset.Rows,
	})
}
