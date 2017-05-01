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

type Operation interface {
	Perform(e *Editor) Operation // returns the operation's inverse
}

type Op struct {
	Initialized bool
	CursorRow   int
	CursorCol   int
	Multiplier  int
}

func (op *Op) init(e *Editor) {
	if !op.Initialized {
		op.Initialized = true
		op.CursorRow = e.CursorRow
		op.CursorCol = e.CursorCol
		op.Multiplier = e.MultiplierValue()
	}
}

func (op *Op) copy(other *Op) {
	op.Initialized = other.Initialized
	op.CursorRow = other.CursorRow
	op.CursorCol = other.CursorCol
	op.Multiplier = other.Multiplier
}

// Replace a character

type ReplaceCharacterOperation struct {
	Op
	Character rune
}

func (op *ReplaceCharacterOperation) Perform(e *Editor) Operation {
	op.init(e)

	old := e.Buffer.Rows[op.CursorRow].ReplaceChar(op.CursorCol, op.Character)
	e.SetCursor(op.CursorRow, op.CursorCol)
	inverse := &ReplaceCharacterOperation{}
	inverse.copy(&op.Op)
	inverse.Character = old
	return inverse
}

// Insert text
//  (todo)

func (e *Editor) InsertChar(c rune) {
	if c == '\n' {
		e.InsertRow()
		e.CursorRow++
		e.CursorCol = 0
		return
	}
	// if the cursor is past the nmber of rows, add a row
	for e.CursorRow >= len(e.Buffer.Rows) {
		e.Buffer.Rows = append(e.Buffer.Rows, NewRow(""))
	}
	e.Buffer.Rows[e.CursorRow].InsertChar(e.CursorCol, c)
	e.CursorCol += 1
}

func (e *Editor) InsertRow() {
	if e.CursorRow > len(e.Buffer.Rows)-1 {
		e.Buffer.Rows = append(e.Buffer.Rows, NewRow(""))
		e.CursorRow = len(e.Buffer.Rows) - 1
	} else {
		position := e.CursorRow
		newRow := e.Buffer.Rows[position].Split(e.CursorCol)
		e.Message = newRow.Text
		i := position + 1
		e.Buffer.Rows = append(e.Buffer.Rows, NewRow(""))
		copy(e.Buffer.Rows[i+1:], e.Buffer.Rows[i:])
		e.Buffer.Rows[i] = newRow
	}
}

func (e *Editor) DeleteRow() {
	if len(e.Buffer.Rows) == 0 {
		return
	}
	e.PasteBoard = ""
	N := e.MultiplierValue()
	for i := 0; i < N; i++ {
		if i > 0 {
			e.PasteBoard += "\n"
		}
		position := e.CursorRow
		e.PasteBoard += e.Buffer.Rows[position].Text
		e.Buffer.Rows = append(e.Buffer.Rows[0:position], e.Buffer.Rows[position+1:]...)
		if position > len(e.Buffer.Rows)-1 {
			position = len(e.Buffer.Rows) - 1
		}
		if position < 0 {
			position = 0
		}
		e.CursorRow = position
	}
}

func (e *Editor) YankRow() {
	if len(e.Buffer.Rows) == 0 {
		return
	}
	e.PasteBoard = ""
	N := e.MultiplierValue()
	for i := 0; i < N; i++ {
		if i > 0 {
			e.PasteBoard += "\n"
		}
		position := e.CursorRow + i
		if position < len(e.Buffer.Rows) {
			e.PasteBoard += e.Buffer.Rows[position].Text
		}
	}
}

func (e *Editor) DeleteWord() {
	if len(e.Buffer.Rows) == 0 {
		return
	}

	c := e.Buffer.Rows[e.CursorRow].DeleteChar(e.CursorCol)
	for {
		if e.CursorCol > e.Buffer.Rows[e.CursorRow].Length()-1 {
			break
		}
		if c == ' ' {
			break
		}
		c = e.Buffer.Rows[e.CursorRow].DeleteChar(e.CursorCol)
	}
	if e.CursorCol > e.Buffer.Rows[e.CursorRow].Length()-1 {
		e.CursorCol--
	}
	if e.CursorCol < 0 {
		e.CursorCol = 0
	}
}

func (e *Editor) DeleteCharacterUnderCursor() {
	if len(e.Buffer.Rows) == 0 {
		return
	}
	e.Buffer.Rows[e.CursorRow].DeleteChar(e.CursorCol)
	if e.CursorCol > e.Buffer.Rows[e.CursorRow].Length()-1 {
		e.CursorCol--
	}
	if e.CursorCol < 0 {
		e.CursorCol = 0
	}
}

func (e *Editor) MoveCursorToStartOfLine() {
	e.CursorCol = 0
}

func (e *Editor) MoveCursorPastEndOfLine() {
	if len(e.Buffer.Rows) == 0 {
		return
	}
	e.CursorCol = e.Buffer.Rows[e.CursorRow].Length()
}

func (e *Editor) KeepCursorInRow() {
	if len(e.Buffer.Rows) == 0 {
		e.CursorCol = 0
	} else {
		if e.CursorRow >= len(e.Buffer.Rows) {
			e.CursorRow = len(e.Buffer.Rows) - 1
		}
		if e.CursorRow < 0 {
			e.CursorRow = 0
		}
		lastIndexInRow := e.Buffer.Rows[e.CursorRow].Length() - 1
		if e.CursorCol > lastIndexInRow {
			e.CursorCol = lastIndexInRow
		}
		if e.CursorCol < 0 {
			e.CursorCol = 0
		}
	}
}

func (e *Editor) InsertLineAboveCursor() {
	if len(e.Buffer.Rows) == 0 {
		e.InsertChar(' ')
	}
	i := e.CursorRow
	e.Buffer.Rows = append(e.Buffer.Rows, NewRow(""))
	copy(e.Buffer.Rows[i+1:], e.Buffer.Rows[i:])
	e.Buffer.Rows[i] = NewRow("")
	e.CursorRow = i
	e.CursorCol = 0
}

func (e *Editor) InsertLineBelowCursor() {
	if len(e.Buffer.Rows) == 0 {
		e.InsertChar(' ')
	}
	i := e.CursorRow
	e.Buffer.Rows = append(e.Buffer.Rows, NewRow(""))
	copy(e.Buffer.Rows[i+2:], e.Buffer.Rows[i+1:])
	e.Buffer.Rows[i+1] = NewRow("")
	e.CursorRow = i + 1
	e.CursorCol = 0
}

func (e *Editor) ReplaceCharacter() {
}

func (e *Editor) Paste() {
	e.InsertLineBelowCursor()
	for _, c := range e.PasteBoard {
		e.InsertChar(c)
	}
}
