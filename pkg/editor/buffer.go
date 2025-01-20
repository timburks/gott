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

	gott "github.com/timburks/gott/pkg/types"
)

// A Buffer represents a file being edited.
// Buffers are displayed in windows but also may be manipulated offscreen.
type Buffer struct {
	Name         string
	ReadOnly     bool
	rows         []*Row
	fileName     string
	languageMode string
	Highlighted  bool
}

func NewBuffer() *Buffer {
	b := &Buffer{}
	b.rows = make([]*Row, 0)
	b.Highlighted = false
	return b
}

func (b *Buffer) SetNameAndReadOnly(name string, readOnly bool) {
	b.Name = name
	b.ReadOnly = readOnly
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
	previous := b.GetBytes()
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

func (b *Buffer) GetBytes() []byte {
	var s string
	for i, row := range b.rows {
		if i > 0 {
			s += "\n"
		}
		s += string(row.GetText())
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
			return row.GetText()[cursor.Col]
		}
	}
	return rune(0)
}

func (b *Buffer) TextFromPosition(row, col int) string {
	if row < len(b.rows) {
		return b.rows[row].TextFromColumn(col)
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

func (b *Buffer) FirstPositionInRowAfterCol(row int, col int, text string) int {
	if row < b.GetRowCount() {
		return b.rows[row].FirstPositionAfterCol(col, text)
	} else {
		return -1
	}
}

func (b *Buffer) LastPositionInRowBeforeCol(row int, col int, text string) int {
	if row < b.GetRowCount() {
		return b.rows[row].LastPositionBeforeCol(col, text)
	} else {
		return -1
	}
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
