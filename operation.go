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
	e.PasteNewLine = true

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

	deletedText := ""

	for i := 0; i < op.Multiplier; i++ {
		if len(e.Buffer.Rows) == 0 {
			break
		}

		// if the row is empty, delete the row...
		if e.Buffer.Rows[e.CursorRow].Length() == 0 {
			position := e.CursorRow
			e.Buffer.Rows = append(e.Buffer.Rows[0:position], e.Buffer.Rows[position+1:]...)
			deletedText += "\n"
		} else {
			// else do this...
			c := e.Buffer.Rows[e.CursorRow].DeleteChar(e.CursorCol)
			deletedText += string(c)
			for {
				if e.CursorCol > e.Buffer.Rows[e.CursorRow].Length()-1 {
					break
				}
				if c == ' ' {
					break
				}
				c = e.Buffer.Rows[e.CursorRow].DeleteChar(e.CursorCol)
				deletedText += string(c)
			}
			if e.CursorCol > e.Buffer.Rows[e.CursorRow].Length()-1 {
				e.CursorCol--
			}
			if e.CursorCol < 0 {
				e.CursorCol = 0
			}
		}

	}
	inverse := &InsertOperation{
		Position: InsertAtCursor,
		Text:     string(deletedText),
	}
	inverse.copy(&op.Op)
	inverse.Undo = true
	return inverse
}

// Delete a character

type DeleteCharacterOperation struct {
	Op
	FinallyDeleteRow bool
}

func (op *DeleteCharacterOperation) Perform(e *Editor) Operation {
	op.init(e)
	log.Printf("Deleting %d character(s) at row %d", op.Multiplier, e.CursorRow)

	if len(e.Buffer.Rows) == 0 {
		return nil
	}
	deletedText := ""
	for i := 0; i < op.Multiplier; i++ {
		if e.Buffer.Rows[e.CursorRow].Length() > 0 {
			c := e.Buffer.Rows[e.CursorRow].DeleteChar(e.CursorCol)
			deletedText += string(c)
		} else if op.Undo && e.CursorRow < len(e.Buffer.Rows)-1 {
			// delete current row
			e.Buffer.Rows = append(e.Buffer.Rows[0:e.CursorRow], e.Buffer.Rows[e.CursorRow+1:]...)
			deletedText += "\n"
		}
	}
	if e.CursorCol > e.Buffer.Rows[e.CursorRow].Length()-1 {
		e.CursorCol--
	}
	if e.CursorCol < 0 {
		e.CursorCol = 0
	}
	if op.FinallyDeleteRow && len(e.Buffer.Rows) > 0 {
		e.Buffer.Rows = append(e.Buffer.Rows[0:e.CursorRow], e.Buffer.Rows[e.CursorRow+1:]...)
	}

	inverse := &InsertOperation{
		Position: InsertAtCursor,
		Text:     deletedText,
	}
	inverse.copy(&op.Op)
	inverse.Undo = true
	return inverse
}

// Paste

type PasteOperation struct {
	Op
}

func (op *PasteOperation) Perform(e *Editor) Operation {
	if e.PasteNewLine {
		e.MoveToStartOfLineBelowCursor()
	}

	op.init(e)

	row := op.CursorRow
	col := op.CursorCol

	for _, c := range e.PasteBoard {
		e.InsertChar(c)
	}
	if e.PasteNewLine {

		e.CursorRow = row
		e.CursorCol = col

		inverse := &DeleteCharacterOperation{}
		inverse.copy(&op.Op)
		inverse.Multiplier = len(e.PasteBoard)
		inverse.CursorCol = 0
		inverse.Undo = true
		return inverse
	} else {
		return nil
	}
}

// Insert

type InsertOperation struct {
	Op
	Position int
	Text     string
	Inverse  *DeleteCharacterOperation
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
		op.CursorRow = e.CursorRow
		op.CursorCol = e.CursorCol

		e.Mode = ModeInsert
		e.Insert = op
	}

	inverse := &DeleteCharacterOperation{}
	inverse.copy(&op.Op)
	inverse.Multiplier = 0
	inverse.Undo = true
	if op.Position == InsertAtNewLineBelowCursor ||
		op.Position == InsertAtNewLineAboveCursor {
		inverse.FinallyDeleteRow = true
	}
	op.Inverse = inverse
	return inverse
}

func (op *InsertOperation) Close() {
	op.Inverse.Multiplier = len(op.Text)
}
