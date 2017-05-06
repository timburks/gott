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
	"io/ioutil"
	"os"
	"strings"
	"unicode"
)

// Editor modes
const (
	ModeEdit    = 0
	ModeInsert  = 1
	ModeCommand = 2
	ModeSearch  = 3
	ModeQuit    = 9999
)

// Move directions
const (
	MoveUp    = 0
	MoveDown  = 1
	MoveRight = 2
	MoveLeft  = 3
)

// Insert positions
const (
	InsertAtCursor             = 0
	InsertAfterCursor          = 1
	InsertAtStartOfLine        = 2
	InsertAfterEndOfLine       = 3
	InsertAtNewLineBelowCursor = 4
	InsertAtNewLineAboveCursor = 5
)

// Paste modes
const (
	PasteAtCursor = 0
	PasteNewLine  = 1
)

// The Editor manages the editing of text in a Buffer.
type Editor struct {
	EditSize   Size        // size of editing area
	Cursor     Point       // cursor position
	Offset     Size        // display offset
	pasteBoard string      // used to cut/copy and paste
	pasteMode  int         // how to paste the string on the pasteboard
	Buffer     *Buffer     // active buffer being edited
	Previous   Operation   // last operation performed, available to repeat
	Undo       []Operation // stack of operations to undo
	Insert     *Insert     // when in insert mode, the current insert operation
}

func NewEditor() *Editor {
	e := &Editor{}
	e.Buffer = NewBuffer()
	return e
}

func (e *Editor) ReadFile(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	e.Buffer.ReadBytes(b)
	e.Buffer.FileName = path
	return nil
}

func (e *Editor) Bytes() []byte {
	return e.Buffer.Bytes()
}

func (e *Editor) WriteFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	b := e.Bytes()
	out, err := gofmt(e.Buffer.FileName, b)
	if err == nil {
		f.Write(out)
	} else {
		f.Write(b)
	}
	return nil
}

func (e *Editor) Perform(op Operation, multiplier int) {
	// perform the operation
	inverse := op.Perform(e, multiplier)
	// save the operation for repeats
	e.Previous = op
	// save the inverse of the operation for undo
	if inverse != nil {
		e.Undo = append(e.Undo, inverse)
	}
}

func (e *Editor) Repeat() {
	if e.Previous != nil {
		inverse := e.Previous.Perform(e, 0)
		if inverse != nil {
			e.Undo = append(e.Undo, inverse)
		}
	}
}

func (e *Editor) PerformUndo() {
	if len(e.Undo) > 0 {
		last := len(e.Undo) - 1
		undo := e.Undo[last]
		e.Undo = e.Undo[0:last]
		undo.Perform(e, 0)
	}
}

func (e *Editor) PerformSearch(text string) {
	if len(e.Buffer.Rows) == 0 {
		return
	}
	row := e.Cursor.Row
	col := e.Cursor.Col + 1

	for {
		var s string
		if col < e.Buffer.Rows[row].Length() {
			s = e.Buffer.Rows[row].Text[col:]
		} else {
			s = ""
		}
		i := strings.Index(s, text)
		if i != -1 {
			// found it
			e.Cursor.Row = row
			e.Cursor.Col = col + i
			return
		} else {
			col = 0
			row = row + 1
			if row == len(e.Buffer.Rows) {
				row = 0
			}
		}
		if row == e.Cursor.Row {
			break
		}
	}
}

func (e *Editor) Scroll() {
	if e.Cursor.Row < e.Offset.Rows {
		e.Offset.Rows = e.Cursor.Row
	}
	if e.Cursor.Row-e.Offset.Rows >= e.EditSize.Rows {
		e.Offset.Rows = e.Cursor.Row - e.EditSize.Rows + 1
	}
	if e.Cursor.Col < e.Offset.Cols {
		e.Offset.Cols = e.Cursor.Col
	}
	if e.Cursor.Col-e.Offset.Cols >= e.EditSize.Cols {
		e.Offset.Cols = e.Cursor.Col - e.EditSize.Cols + 1
	}
}

func (e *Editor) MoveCursor(direction int) {
	switch direction {
	case MoveLeft:
		if e.Cursor.Col > 0 {
			e.Cursor.Col--
		}
	case MoveRight:
		if e.Cursor.Row < len(e.Buffer.Rows) {
			rowLength := e.Buffer.Rows[e.Cursor.Row].Length()
			if e.Cursor.Col < rowLength-1 {
				e.Cursor.Col++
			}
		}
	case MoveUp:
		if e.Cursor.Row > 0 {
			e.Cursor.Row--
		}
	case MoveDown:
		if e.Cursor.Row < len(e.Buffer.Rows)-1 {
			e.Cursor.Row++
		}
	}
	// don't go past the end of the current line
	if e.Cursor.Row < len(e.Buffer.Rows) {
		rowLength := e.Buffer.Rows[e.Cursor.Row].Length()
		if e.Cursor.Col > rowLength-1 {
			e.Cursor.Col = rowLength - 1
			if e.Cursor.Col < 0 {
				e.Cursor.Col = 0
			}
		}
	}
}

// These editor primitives will make changes in insert mode and associate them with to the current operation.

func (e *Editor) InsertChar(c rune) {
	if e.Insert != nil {
		e.Insert.Text += string(c)
	}
	if c == '\n' {
		e.InsertRow()
		e.Cursor.Row++
		e.Cursor.Col = 0
		return
	}
	// if the cursor is past the nmber of rows, add a row
	for e.Cursor.Row >= len(e.Buffer.Rows) {
		e.AppendBlankRow()
	}
	e.Buffer.Rows[e.Cursor.Row].InsertChar(e.Cursor.Col, c)
	e.Cursor.Col += 1
}

func (e *Editor) InsertRow() {
	if e.Cursor.Row >= len(e.Buffer.Rows) {
		// we should never get here
		e.AppendBlankRow()
	} else {
		newRow := e.Buffer.Rows[e.Cursor.Row].Split(e.Cursor.Col)
		i := e.Cursor.Row + 1
		// add a dummy row at the end of the Rows slice
		e.AppendBlankRow()
		// move rows to make room for the one we are adding
		copy(e.Buffer.Rows[i+1:], e.Buffer.Rows[i:])
		// add the new row
		e.Buffer.Rows[i] = newRow
	}
}

func (e *Editor) BackspaceChar() rune {
	if len(e.Buffer.Rows) == 0 {
		return rune(0)
	}
	if len(e.Insert.Text) == 0 {
		return rune(0)
	}
	e.Insert.Text = e.Insert.Text[0 : len(e.Insert.Text)-1]
	if e.Cursor.Col > 0 {
		c := e.Buffer.Rows[e.Cursor.Row].DeleteChar(e.Cursor.Col - 1)
		e.Cursor.Col--
		return c
	} else if e.Cursor.Row > 0 {
		// remove the current row and join it with the previous one
		oldRowText := e.Buffer.Rows[e.Cursor.Row].Text
		var newCursor Point
		newCursor.Col = len(e.Buffer.Rows[e.Cursor.Row-1].Text)
		e.Buffer.Rows[e.Cursor.Row-1].Text += oldRowText
		e.Buffer.Rows = append(e.Buffer.Rows[0:e.Cursor.Row], e.Buffer.Rows[e.Cursor.Row+1:]...)
		e.Cursor.Row--
		e.Cursor.Col = newCursor.Col
		return rune('\n')
	} else {
		return rune(0)
	}
}

func (e *Editor) YankRow(multiplier int) {
	if len(e.Buffer.Rows) == 0 {
		return
	}
	pasteText := ""
	for i := 0; i < multiplier; i++ {
		position := e.Cursor.Row + i
		if position < len(e.Buffer.Rows) {
			pasteText += e.Buffer.Rows[position].Text + "\n"
		}
	}

	e.SetPasteBoard(pasteText, PasteNewLine)

}

func (e *Editor) KeepCursorInRow() {
	if len(e.Buffer.Rows) == 0 {
		e.Cursor.Col = 0
	} else {
		if e.Cursor.Row >= len(e.Buffer.Rows) {
			e.Cursor.Row = len(e.Buffer.Rows) - 1
		}
		if e.Cursor.Row < 0 {
			e.Cursor.Row = 0
		}
		lastIndexInRow := e.Buffer.Rows[e.Cursor.Row].Length() - 1
		if e.Cursor.Col > lastIndexInRow {
			e.Cursor.Col = lastIndexInRow
		}
		if e.Cursor.Col < 0 {
			e.Cursor.Col = 0
		}
	}
}

func (e *Editor) AppendBlankRow() {
	e.Buffer.Rows = append(e.Buffer.Rows, NewRow(""))
}

func (e *Editor) InsertLineAboveCursor() {
	e.AppendBlankRow()
	copy(e.Buffer.Rows[e.Cursor.Row+1:], e.Buffer.Rows[e.Cursor.Row:])
	e.Buffer.Rows[e.Cursor.Row] = NewRow("")
	e.Cursor.Col = 0
}

func (e *Editor) InsertLineBelowCursor() {
	e.AppendBlankRow()
	copy(e.Buffer.Rows[e.Cursor.Row+2:], e.Buffer.Rows[e.Cursor.Row+1:])
	e.Buffer.Rows[e.Cursor.Row+1] = NewRow("")
	e.Cursor.Row += 1
	e.Cursor.Col = 0
}

func (e *Editor) MoveCursorToStartOfLine() {
	e.Cursor.Col = 0
}

func (e *Editor) MoveToStartOfLineBelowCursor() {
	e.Cursor.Col = 0
	e.Cursor.Row += 1
}

// editable

func (e *Editor) GetCursor() Point {
	return e.Cursor
}

func (e *Editor) SetCursor(cursor Point) {
	e.Cursor = cursor
}

func (e *Editor) ReplaceCharacterAtCursor(cursor Point, c rune) rune {
	return e.Buffer.Rows[cursor.Row].ReplaceChar(cursor.Col, c)
}

func (e *Editor) DeleteRowsAtCursor(multiplier int) string {
	deletedText := ""
	for i := 0; i < multiplier; i++ {
		position := e.Cursor.Row
		if position < len(e.Buffer.Rows) {
			deletedText += e.Buffer.Rows[position].Text
			deletedText += "\n"
			e.Buffer.Rows = append(e.Buffer.Rows[0:position], e.Buffer.Rows[position+1:]...)
			position = clipToRange(position, 0, len(e.Buffer.Rows)-1)
			e.Cursor.Row = position
		} else {
			break
		}
	}
	return deletedText
}

func (e *Editor) SetPasteBoard(text string, mode int) {
	e.pasteBoard = text
	e.pasteMode = mode
}

func (e *Editor) DeleteWordsAtCursor(multiplier int) string {
	deletedText := ""
	for i := 0; i < multiplier; i++ {
		if len(e.Buffer.Rows) == 0 {
			break
		}
		// if the row is empty, delete the row...
		if e.Buffer.Rows[e.Cursor.Row].Length() == 0 {
			position := e.Cursor.Row
			e.Buffer.Rows = append(e.Buffer.Rows[0:position], e.Buffer.Rows[position+1:]...)
			deletedText += "\n"
		} else {
			// else do this...
			c := e.Buffer.Rows[e.Cursor.Row].DeleteChar(e.Cursor.Col)
			deletedText += string(c)
			for {
				if e.Cursor.Col > e.Buffer.Rows[e.Cursor.Row].Length()-1 {
					break
				}
				if c == ' ' {
					break
				}
				c = e.Buffer.Rows[e.Cursor.Row].DeleteChar(e.Cursor.Col)
				deletedText += string(c)
			}
			if e.Cursor.Col > e.Buffer.Rows[e.Cursor.Row].Length()-1 {
				e.Cursor.Col--
			}
			if e.Cursor.Col < 0 {
				e.Cursor.Col = 0
			}
		}
	}
	return deletedText
}

func (e *Editor) DeleteCharactersAtCursor(multiplier int, undo bool, finallyDeleteRow bool) string {
	deletedText := ""
	if len(e.Buffer.Rows) == 0 {
		return deletedText
	}
	for i := 0; i < multiplier; i++ {
		if e.Buffer.Rows[e.Cursor.Row].Length() > 0 {
			c := e.Buffer.Rows[e.Cursor.Row].DeleteChar(e.Cursor.Col)
			deletedText += string(c)
		} else if undo && e.Cursor.Row < len(e.Buffer.Rows)-1 {
			// delete current row
			e.Buffer.Rows = append(e.Buffer.Rows[0:e.Cursor.Row], e.Buffer.Rows[e.Cursor.Row+1:]...)
			deletedText += "\n"
		}
	}
	if e.Cursor.Col > e.Buffer.Rows[e.Cursor.Row].Length()-1 {
		e.Cursor.Col--
	}
	if e.Cursor.Col < 0 {
		e.Cursor.Col = 0
	}
	if finallyDeleteRow && len(e.Buffer.Rows) > 0 {
		e.Buffer.Rows = append(e.Buffer.Rows[0:e.Cursor.Row], e.Buffer.Rows[e.Cursor.Row+1:]...)
	}
	return deletedText
}

func (e *Editor) InsertText(text string, position int) (Point, int) {
	if len(e.Buffer.Rows) == 0 {
		e.AppendBlankRow()
	}
	switch position {
	case InsertAtCursor:
		break
	case InsertAfterCursor:
		e.Cursor.Col++
		e.Cursor.Col = clipToRange(e.Cursor.Col, 0, e.Buffer.Rows[e.Cursor.Row].Length())
	case InsertAtStartOfLine:
		e.Cursor.Col = 0
	case InsertAfterEndOfLine:
		e.Cursor.Col = e.Buffer.Rows[e.Cursor.Row].Length()
	case InsertAtNewLineBelowCursor:
		e.InsertLineBelowCursor()
	case InsertAtNewLineAboveCursor:
		e.InsertLineAboveCursor()
	}
	var mode int
	if text != "" {
		r := e.Cursor.Row
		c := e.Cursor.Col
		for _, c := range text {
			e.InsertChar(c)
		}
		e.Cursor.Row = r
		e.Cursor.Col = c
		mode = ModeEdit
	} else {
		mode = ModeInsert
	}
	return e.Cursor, mode
}

func (e *Editor) SetInsertOperation(insert *Insert) {
	e.Insert = insert
}

func (e *Editor) GetPasteMode() int {
	return e.pasteMode
}

func (e *Editor) GetPasteText() string {
	return e.pasteBoard
}

func (e *Editor) ReverseCaseCharactersAtCursor(multiplier int) {
	if len(e.Buffer.Rows) == 0 {
		return
	}
	row := &e.Buffer.Rows[e.Cursor.Row]
	for i := 0; i < multiplier; i++ {
		c := rune(row.Text[e.Cursor.Col])
		if unicode.IsUpper(c) {
			row.ReplaceChar(e.Cursor.Col, unicode.ToLower(c))
		}
		if unicode.IsLower(c) {
			row.ReplaceChar(e.Cursor.Col, unicode.ToUpper(c))
		}
		if e.Cursor.Col < row.Length()-1 {
			e.Cursor.Col++
		}
	}
}
