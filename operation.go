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

type Operation interface {
	Perform(e *Editor) Operation // performs the operation and returns its inverse
}

type Op struct {
	CursorRow  int
	CursorCol  int
	Multiplier int
	Undo       bool
}

func (op *Op) init(e *Editor) {
	if op.Undo {
		e.CursorRow = op.CursorRow
		e.CursorCol = op.CursorCol
	} else {
		op.CursorRow = e.CursorRow
		op.CursorCol = e.CursorCol
		op.Multiplier = e.MultiplierValue()
	}
}

func (op *Op) copy(other *Op) {
	op.CursorRow = other.CursorRow
	op.CursorCol = other.CursorCol
	op.Multiplier = other.Multiplier
	op.Undo = other.Undo
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
	inverse.Undo = true
	inverse.Character = old
	return inverse
}

// Delete a row

type DeleteRowOperation struct {
	Op
}

func (op *DeleteRowOperation) Perform(e *Editor) Operation {
	e.MoveCursorToStartOfLine()
	op.init(e)
	log.Printf("Deleting %d row(s) at row %d", op.Multiplier, e.CursorRow)
	deletedText := ""
	for i := 0; i < op.Multiplier; i++ {
		position := e.CursorRow
		if position < len(e.Buffer.Rows) {
			deletedText += e.Buffer.Rows[position].Text
			deletedText += "\n"
			e.Buffer.Rows = append(e.Buffer.Rows[0:position], e.Buffer.Rows[position+1:]...)
			position = clipToRange(position, 0, len(e.Buffer.Rows)-1)
			e.CursorRow = position
		} else {
			break
		}
	}
	e.PasteBoard = deletedText
	inverse := &InsertOperation{
		Position: InsertAtCursor,
		Text:     deletedText,
	}
	inverse.copy(&op.Op)
	inverse.Undo = true
	return inverse
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
	Text     string
}

func (op *InsertOperation) Perform(e *Editor) Operation {
	op.init(e)
	if len(e.Buffer.Rows) == 0 {
		e.AppendBlankRow()
	}
	switch op.Position {
	case InsertAtCursor:
		break
	case InsertAfterCursor:
		e.CursorCol++
		e.CursorCol = clipToRange(e.CursorCol, 0, e.Buffer.Rows[e.CursorRow].Length())
	case InsertAtStartOfLine:
		e.CursorCol = 0
	case InsertAfterEndOfLine:
		e.CursorCol = e.Buffer.Rows[e.CursorRow].Length()
	case InsertAtNewLineBelowCursor:
		e.InsertLineBelowCursor()
	case InsertAtNewLineAboveCursor:
		e.InsertLineAboveCursor()
	}
	if op.Text != "" {
		e.CursorRow = op.CursorRow
		e.CursorCol = op.CursorCol
		r := e.CursorRow
		c := e.CursorCol
		for _, c := range op.Text {
			e.InsertChar(c)
		}
		e.CursorRow = r
		e.CursorCol = c
	} else {
		e.Mode = ModeInsert
		e.Insert = op
	}
	return nil
}

func (op *InsertOperation) Close() {
	log.Printf("Inserted text:\n%s", op.Text)
}
