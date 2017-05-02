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
	"log"
)

func clipToRange(i, min, max int) int {
	if i > max {
		i = max
	}
	if i < min {
		i = min
	}
	return i
}

type Operation interface {
	Perform(e *Editor) Operation // performs the operation and returns its inverse
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

// Delete a row

type DeleteRowOperation struct {
	Op
}

func (op *DeleteRowOperation) Perform(e *Editor) Operation {
	op.init(e)
	log.Printf("Deleting %d row(s) at row %d", op.Multiplier, e.CursorRow)

	e.PasteBoard = ""
	for i := 0; i < op.Multiplier; i++ {
		if i > 0 {
			e.PasteBoard += "\n"
		}
		position := e.CursorRow
		if position < len(e.Buffer.Rows) {
			e.PasteBoard += e.Buffer.Rows[position].Text
			e.Buffer.Rows = append(e.Buffer.Rows[0:position], e.Buffer.Rows[position+1:]...)
			position = clipToRange(position, 0, len(e.Buffer.Rows)-1)
			e.CursorRow = position
		} else {
			break
		}
	}
	return nil
}

// Delete a word

type DeleteWordOperation struct {
	Op
}

func (op *DeleteWordOperation) Perform(e *Editor) Operation {
	op.init(e)
	log.Printf("Deleting %d words(s) at row %d", op.Multiplier, e.CursorRow)

	if len(e.Buffer.Rows) == 0 {
		return nil
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
	return nil
}

// Delete a character

type DeleteCharacterOperation struct {
	Op
}

func (op *DeleteCharacterOperation) Perform(e *Editor) Operation {
	op.init(e)
	log.Printf("Deleting %d character(s) at row %d", op.Multiplier, e.CursorRow)

	if len(e.Buffer.Rows) == 0 {
		return nil
	}
	e.Buffer.Rows[e.CursorRow].DeleteChar(e.CursorCol)
	if e.CursorCol > e.Buffer.Rows[e.CursorRow].Length()-1 {
		e.CursorCol--
	}
	if e.CursorCol < 0 {
		e.CursorCol = 0
	}
	return nil
}

// Paste

type PasteOperation struct {
	Op
}

func (op *PasteOperation) Perform(e *Editor) Operation {
	op.init(e)
	e.InsertLineBelowCursor()
	for _, c := range e.PasteBoard {
		e.InsertChar(c)
	}
	return nil
}

// Insert

type InsertOperation struct {
	Op
	Position int
}

func (op *InsertOperation) Perform(e *Editor) Operation {
	op.init(e)
	switch op.Position {
	case InsertAtCursor:
		break
	case InsertAfterCursor:
		e.CursorCol++
	case InsertAtStartOfLine:
		e.MoveCursorToStartOfLine()
	case InsertAfterEndOfLine:
		e.MoveCursorPastEndOfLine()
	case InsertAtNewLineBelowCursor:
		e.InsertLineBelowCursor()
	case InsertAtNewLineAboveCursor:
		e.InsertLineAboveCursor()
	}
	e.Mode = ModeInsert
	return nil
}
