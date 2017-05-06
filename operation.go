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
	Perform(e Editable) Operation // performs the operation and returns its inverse
}

type Op struct {
	Cursor     Point
	Multiplier int
	Undo       bool
}

func (op *Op) init(e Editable) {
	if op.Undo {
		e.SetCursor(op.Cursor)
	} else {
		op.Cursor = e.GetCursor()
		if op.Multiplier == 0 {
			op.Multiplier = e.MultiplierValue()
		}
	}
}

func (op *Op) copy(other *Op) {
	op.Cursor.Row = other.Cursor.Row
	op.Cursor.Col = other.Cursor.Col
	op.Multiplier = other.Multiplier
	op.Undo = other.Undo
}

// Replace a character

type ReplaceCharacter struct {
	Op
	Character rune
}

func (op *ReplaceCharacter) Perform(e Editable) Operation {
	op.init(e)
	old := e.ReplaceCharacterAtCursor(op.Cursor, op.Character)
	inverse := &ReplaceCharacter{}
	inverse.copy(&op.Op)
	inverse.Undo = true
	inverse.Character = old
	return inverse
}

// Delete a row

type DeleteRow struct {
	Op
}

func (op *DeleteRow) Perform(e Editable) Operation {
	e.MoveCursorToStartOfLine()
	op.init(e)
	log.Printf("Deleting %d row(s) at row %d", op.Multiplier, e.GetCursor().Row)
	deletedText := e.DeleteRowsAtCursor(op.Multiplier)
	e.SetPasteBoard(deletedText, PasteNewLine)
	inverse := &Insert{
		Position: InsertAtCursor,
		Text:     deletedText,
	}
	inverse.copy(&op.Op)
	inverse.Undo = true
	return inverse
}

// Delete a word

type DeleteWord struct {
	Op
}

func (op *DeleteWord) Perform(e Editable) Operation {
	op.init(e)
	log.Printf("Deleting %d words(s) at row %d", op.Multiplier, e.GetCursor().Row)
	deletedText := e.DeleteWordsAtCursor(op.Multiplier)
	e.SetPasteBoard(deletedText, InsertAtCursor)
	inverse := &Insert{
		Position: InsertAtCursor,
		Text:     string(deletedText),
	}
	inverse.copy(&op.Op)
	inverse.Undo = true
	return inverse
}

// Delete a character

type DeleteCharacter struct {
	Op
	FinallyDeleteRow bool
}

func (op *DeleteCharacter) Perform(e Editable) Operation {
	op.init(e)
	log.Printf("Deleting %d character(s) at row %d", op.Multiplier, e.GetCursor().Row)

	deletedText := e.DeleteCharactersAtCursor(op.Multiplier, op.Undo, op.FinallyDeleteRow)

	inverse := &Insert{
		Position: InsertAtCursor,
		Text:     deletedText,
	}
	inverse.copy(&op.Op)
	inverse.Undo = true
	return inverse
}

// Paste

type Paste struct {
	Op
}

func (op *Paste) Perform(e Editable) Operation {
	if e.GetPasteMode() == PasteNewLine {
		e.MoveToStartOfLineBelowCursor()
	}

	op.init(e)

	cursor := op.Cursor

	for _, c := range e.GetPasteText() {
		e.InsertChar(c)
	}
	if e.GetPasteMode() == PasteNewLine {
		e.SetCursor(cursor)
		inverse := &DeleteCharacter{}
		inverse.copy(&op.Op)
		inverse.Multiplier = len(e.GetPasteText())
		inverse.Cursor.Col = 0
		inverse.Undo = true
		return inverse
	} else {
		return nil
	}
}

// Insert

type Insert struct {
	Op
	Position int
	Text     string
	Inverse  *DeleteCharacter
}

func (op *Insert) Perform(e Editable) Operation {
	op.init(e)

	if op.Text != "" {
		e.SetCursor(op.Cursor)
	} else {
		op.Cursor = e.GetCursor()
		e.SetInsertOperation(op)
	}

	op.Cursor = e.InsertText(op.Text, op.Position)

	inverse := &DeleteCharacter{}
	inverse.copy(&op.Op)
	inverse.Multiplier = len(op.Text)
	inverse.Undo = true
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

func (op *ReverseCaseCharacter) Perform(e Editable) Operation {
	op.init(e)
	log.Printf("Reversing case of %d character(s) at row %d", op.Multiplier, e.GetCursor().Row)

	e.ReverseCaseCharactersAtCursor(op.Multiplier)
	if op.Undo {
		e.SetCursor(op.Cursor)
	}

	inverse := &ReverseCaseCharacter{}
	inverse.copy(&op.Op)
	inverse.Undo = true
	return inverse
}
