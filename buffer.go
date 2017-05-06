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

import (
	"fmt"
	"strings"

	"github.com/nsf/termbox-go"
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
	for _, row := range b.rows {
		s += string(row.Text) + "\n"
	}
	return []byte(s)
}

func (b *Buffer) RowCount() int {
	return len(b.rows)
}

func (b *Buffer) RowLength(i int) int {
	return b.rows[i].Length()
}

func (b *Buffer) TextAfterPosition(row, col int) string {
	return string(b.rows[row].Text[col:])
}

func (b *Buffer) InsertCharacter(row, col int, c rune) {
	b.rows[row].InsertChar(col, c)
}

// draw text in an area defined by origin and size with a specified offset into the buffer
func (b *Buffer) Render(origin Point, size Size, offset Size) {
	for i := origin.Row; i < size.Rows; i++ {
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
			if i == size.Rows/3 {
				welcome := fmt.Sprintf("the gott editor -- version %s", VERSION)
				padding := (size.Cols - len(welcome)) / 2
				for j := 1; j <= padding; j++ {
					line = line + " "
				}
				line += welcome
			}
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
