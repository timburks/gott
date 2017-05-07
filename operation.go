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
	Perform(e Editable, multiplier int) Operation // performs the operation and returns its inverse
}

type Op struct {
	Cursor     Point
	Multiplier int
	Undo       bool
}

func (op *Op) init(e Editable, multiplier int) {
	if op.Undo {
		e.SetCursor(op.Cursor)
	} else {
		op.Cursor = e.GetCursor()
		if op.Multiplier == 0 {
			op.Multiplier = multiplier
		}
	}
}

func (op *Op) copyForUndo(other *Op) {
	op.Cursor = other.Cursor
	op.Multiplier = other.Multiplier
	op.Undo = true
}

// Replace a character

type ReplaceCharacter struct {
	Op
	Character rune
}

func (op *ReplaceCharacter) Perform(e Editable, multiplier int) Operation {
	op.init(e, multiplier)
	old := e.ReplaceCharacterAtCursor(op.Cursor, op.Character)
	inverse := &ReplaceCharacter{}
	inverse.copyForUndo(&op.Op)
	inverse.Character = old
	return inverse
}

// Delete a row

type DeleteRow struct {
	Op
}

func (op *DeleteRow) Perform(e Editable, multiplier int) Operation {
	e.MoveCursorToStartOfLine()
	op.init(e, multiplier)
	log.Printf("Deleting %d row(s) at row %d", op.Multiplier, e.GetCursor().Row)
	deletedText := e.DeleteRowsAtCursor(op.Multiplier)
	e.SetPasteBoard(deletedText, PasteNewLine)
	inverse := &Insert{
		Position: InsertAtCursor,
		Text:     deletedText,
	}
	inverse.copyForUndo(&op.Op)
	return inverse
}

// Delete a word

type DeleteWord struct {
	Op
}

func (op *DeleteWord) Perform(e Editable, multiplier int) Operation {
	op.init(e, multiplier)
	log.Printf("Deleting %d words(s) at row %d", op.Multiplier, e.GetCursor().Row)
	deletedText := e.DeleteWordsAtCursor(op.Multiplier)
	e.SetPasteBoard(deletedText, InsertAtCursor)
	inverse := &Insert{
		Position: InsertAtCursor,
		Text:     string(deletedText),
	}
	inverse.copyForUndo(&op.Op)
	return inverse
}

// Delete a character

type DeleteCharacter struct {
	Op
	FinallyDeleteRow bool
}

func (op *DeleteCharacter) Perform(e Editable, multiplier int) Operation {
	op.init(e, multiplier)
	log.Printf("Deleting %d character(s) at %d,%d undo=%t", op.Multiplier, e.GetCursor().Row, e.GetCursor().Col, op.Undo)

	deletedText := e.DeleteCharactersAtCursor(op.Multiplier, op.Undo, op.FinallyDeleteRow)
	log.Printf("Deleted: [%s]", deletedText)
	inverse := &Insert{
		Position: InsertAtCursor,
		Text:     deletedText,
	}
	inverse.copyForUndo(&op.Op)
	return inverse
}

// Paste

type Paste struct {
	Op
}

func (op *Paste) Perform(e Editable, multiplier int) Operation {
	if e.GetPasteMode() == PasteNewLine {
		e.MoveCursorToStartOfLineBelowCursor()
	}

	op.init(e, multiplier)

	cursor := op.Cursor

	for _, c := range e.GetPasteText() {
		e.InsertChar(c)
	}
	if e.GetPasteMode() == PasteNewLine {
		e.SetCursor(cursor)
		inverse := &DeleteCharacter{}
		inverse.copyForUndo(&op.Op)
		inverse.Multiplier = len(e.GetPasteText())
		inverse.Cursor.Col = 0
		return inverse
	} else {
		return nil
	}
}

// Insert

type Insert struct {
	Op
	Position  int
	Text      string
	Inverse   *DeleteCharacter
	Commander *Commander
}

func (op *Insert) Perform(e Editable, multiplier int) Operation {
	op.init(e, multiplier)

	if op.Text != "" {
		e.SetCursor(op.Cursor)
	} else {
		op.Cursor = e.GetCursor()
		e.SetInsertOperation(op)
	}

	var newMode int
	op.Cursor, newMode = e.InsertText(op.Text, op.Position)
	if op.Commander != nil {
		op.Commander.SetMode(newMode)
	}

	inverse := &DeleteCharacter{}
	inverse.copyForUndo(&op.Op)
	inverse.Multiplier = len(op.Text)
	if op.Position == InsertAtNewLineBelowCursor ||
		op.Position == InsertAtNewLineAboveCursor {
		inverse.FinallyDeleteRow = true
	}
	op.Inverse = inverse
	return inverse
}

func (op *Insert) Close() {
	op.Inverse.Multiplier = len(op.Text)
}

// Reverse the case of a character

type ReverseCaseCharacter struct {
	Op
}

func (op *ReverseCaseCharacter) Perform(e Editable, multiplier int) Operation {
	op.init(e, multiplier)
	log.Printf("Reversing case of %d character(s) at row %d", op.Multiplier, e.GetCursor().Row)

	e.ReverseCaseCharactersAtCursor(op.Multiplier)
	if op.Undo {
		e.SetCursor(op.Cursor)
	}

	inverse := &ReverseCaseCharacter{}
	inverse.copyForUndo(&op.Op)
	return inverse
}
