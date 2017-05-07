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

	"github.com/nsf/termbox-go"

	gott "github.com/timburks/gott/types"
)

// A Buffer represents a file being edited

type Buffer struct {
	rows     []Row
	FileName string
}

func NewBuffer() *Buffer {
	b := &Buffer{}
	b.rows = make([]Row, 0)
	return b
}

func (b *Buffer) GetFileName() string {
	return b.FileName
}

func (b *Buffer) ReadBytes(bytes []byte) {
	s := string(bytes)
	lines := strings.Split(s, "\n")
	b.rows = make([]Row, 0)
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

func (b *Buffer) TextAfter(row, col int) string {
	if row < len(b.rows) {
		return b.rows[row].TextAfter(col)
	} else {
		return ""
	}
}

func (b *Buffer) InsertCharacter(row, col int, c rune) {
	if row < len(b.rows) {
		b.rows[row].InsertChar(col, c)
	}
}

func (b *Buffer) DeleteRow(row int) {
	if row < len(b.rows) {
		b.rows = append(b.rows[0:row], b.rows[row+1:]...)
	}
}

func (b *Buffer) DeleteCharacters(row int, col int, count int, joinLines bool) string {
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

// draw text in an area defined by origin and size with a specified offset into the buffer
func (b *Buffer) Render(origin gott.Point, size gott.Size, offset gott.Size) {
	for i := origin.Row; i < origin.Row+size.Rows; i++ {
		var line string
		if (i + offset.Rows) < len(b.rows) {
			line = b.rows[i+offset.Rows].DisplayText()
			if offset.Cols < len(line) {
				line = line[offset.Cols:]
			} else {
				line = ""
			}
		} else {
			line = "~"
		}
		// truncate line to fit screen
		if len(line) > size.Cols {
			line = line[0:size.Cols]
		}
		for j, c := range line {
			termbox.SetCell(j+origin.Col, i, rune(c), termbox.ColorWhite, termbox.ColorBlack)
		}
	}
}
