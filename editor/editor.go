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
	"io/ioutil"
	"os"
	"strings"
	"unicode"

	gott "github.com/timburks/gott/types"
)

// The Editor manages the editing of text in a Buffer.
type Editor struct {
	Cursor    gott.Point           // cursor position
	Offset    gott.Size            // display offset
	Buffer    *Buffer              // active buffer being edited
	size      gott.Size            // size of editing area
	pasteText string               // used to cut/copy and paste
	pasteMode int                  // how to paste the string on the pasteboard
	previous  gott.Operation       // last operation performed, available to repeat
	undo      []gott.Operation     // stack of operations to undo
	insert    gott.InsertOperation // when in insert mode, the current insert operation
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
	if strings.HasSuffix(path, ".go") {
		out, err := Gofmt(e.Buffer.FileName, b)
		if err == nil {
			f.Write(out)
		} else {
			f.Write(b)
		}
	} else {
		f.Write(b)
	}
	return nil
}

func (e *Editor) Perform(op gott.Operation, multiplier int) {
	// perform the operation
	inverse := op.Perform(e, multiplier)
	// save the operation for repeats
	e.previous = op
	// save the inverse of the operation for undo
	if inverse != nil {
		e.undo = append(e.undo, inverse)
	}
}

func (e *Editor) Repeat() {
	if e.previous != nil {
		inverse := e.previous.Perform(e, 0)
		if inverse != nil {
			e.undo = append(e.undo, inverse)
		}
	}
}

func (e *Editor) PerformUndo() {
	if len(e.undo) > 0 {
		last := len(e.undo) - 1
		undo := e.undo[last]
		e.undo = e.undo[0:last]
		undo.Perform(e, 0)
	}
}

func (e *Editor) PerformSearch(text string) {
	if e.Buffer.GetRowCount() == 0 {
		return
	}
	row := e.Cursor.Row
	col := e.Cursor.Col + 1

	for {
		var s string
		if col < e.Buffer.GetRowLength(row) {
			s = e.Buffer.TextAfter(row, col)
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
			if row == e.Buffer.GetRowCount() {
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
	if e.Cursor.Row-e.Offset.Rows >= e.size.Rows {
		e.Offset.Rows = e.Cursor.Row - e.size.Rows + 1
	}
	if e.Cursor.Col < e.Offset.Cols {
		e.Offset.Cols = e.Cursor.Col
	}
	if e.Cursor.Col-e.Offset.Cols >= e.size.Cols {
		e.Offset.Cols = e.Cursor.Col - e.size.Cols + 1
	}
}

func (e *Editor) MoveCursor(direction int) {
	switch direction {
	case gott.MoveLeft:
		if e.Cursor.Col > 0 {
			e.Cursor.Col--
		}
	case gott.MoveRight:
		if e.Cursor.Row < e.Buffer.GetRowCount() {
			rowLength := e.Buffer.GetRowLength(e.Cursor.Row)
			if e.Cursor.Col < rowLength-1 {
				e.Cursor.Col++
			}
		}
	case gott.MoveUp:
		if e.Cursor.Row > 0 {
			e.Cursor.Row--
		}
	case gott.MoveDown:
		if e.Cursor.Row < e.Buffer.GetRowCount()-1 {
			e.Cursor.Row++
		}
	}
	// don't go past the end of the current line
	if e.Cursor.Row < e.Buffer.GetRowCount() {
		rowLength := e.Buffer.GetRowLength(e.Cursor.Row)
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
	if e.insert != nil {
		e.insert.AddCharacter(c)
	}
	if c == '\n' {
		e.InsertRow()
		e.Cursor.Row++
		e.Cursor.Col = 0
		return
	}
	// if the cursor is past the nmber of rows, add a row
	for e.Cursor.Row >= e.Buffer.GetRowCount() {
		e.AppendBlankRow()
	}
	e.Buffer.InsertCharacter(e.Cursor.Row, e.Cursor.Col, c)
	e.Cursor.Col += 1
}

func (e *Editor) InsertRow() {
	if e.Cursor.Row >= e.Buffer.GetRowCount() {
		// we should never get here
		e.AppendBlankRow()
	} else {
		newRow := e.Buffer.rows[e.Cursor.Row].Split(e.Cursor.Col)
		i := e.Cursor.Row + 1
		// add a dummy row at the end of the Rows slice
		e.AppendBlankRow()
		// move rows to make room for the one we are adding
		copy(e.Buffer.rows[i+1:], e.Buffer.rows[i:])
		// add the new row
		e.Buffer.rows[i] = newRow
	}
}

func (e *Editor) BackspaceChar() rune {
	if e.Buffer.GetRowCount() == 0 {
		return rune(0)
	}
	if e.insert.Length() == 0 {
		return rune(0)
	}
	e.insert.DeleteCharacter()
	if e.Cursor.Col > 0 {
		c := e.Buffer.rows[e.Cursor.Row].DeleteChar(e.Cursor.Col - 1)
		e.Cursor.Col--
		return c
	} else if e.Cursor.Row > 0 {
		// remove the current row and join it with the previous one
		oldRowText := e.Buffer.rows[e.Cursor.Row].Text
		var newCursor gott.Point
		newCursor.Col = len(e.Buffer.rows[e.Cursor.Row-1].Text)
		e.Buffer.rows[e.Cursor.Row-1].Text = append(e.Buffer.rows[e.Cursor.Row-1].Text, oldRowText...)
		e.Buffer.rows = append(e.Buffer.rows[0:e.Cursor.Row], e.Buffer.rows[e.Cursor.Row+1:]...)
		e.Cursor.Row--
		e.Cursor.Col = newCursor.Col
		return rune('\n')
	} else {
		return rune(0)
	}
}

func (e *Editor) YankRow(multiplier int) {
	if e.Buffer.GetRowCount() == 0 {
		return
	}
	pasteText := ""
	for i := 0; i < multiplier; i++ {
		position := e.Cursor.Row + i
		if position < e.Buffer.GetRowCount() {
			pasteText += string(e.Buffer.rows[position].Text) + "\n"
		}
	}

	e.SetPasteBoard(pasteText, gott.PasteNewLine)
}

func (e *Editor) KeepCursorInRow() {
	if e.Buffer.GetRowCount() == 0 {
		e.Cursor.Col = 0
	} else {
		if e.Cursor.Row >= e.Buffer.GetRowCount() {
			e.Cursor.Row = e.Buffer.GetRowCount() - 1
		}
		if e.Cursor.Row < 0 {
			e.Cursor.Row = 0
		}
		lastIndexInRow := e.Buffer.rows[e.Cursor.Row].Length() - 1
		if e.Cursor.Col > lastIndexInRow {
			e.Cursor.Col = lastIndexInRow
		}
		if e.Cursor.Col < 0 {
			e.Cursor.Col = 0
		}
	}
}

func (e *Editor) AppendBlankRow() {
	e.Buffer.rows = append(e.Buffer.rows, NewRow(""))
}

func (e *Editor) InsertLineAboveCursor() {
	e.AppendBlankRow()
	copy(e.Buffer.rows[e.Cursor.Row+1:], e.Buffer.rows[e.Cursor.Row:])
	e.Buffer.rows[e.Cursor.Row] = NewRow("")
	e.Cursor.Col = 0
}

func (e *Editor) InsertLineBelowCursor() {
	e.AppendBlankRow()
	copy(e.Buffer.rows[e.Cursor.Row+2:], e.Buffer.rows[e.Cursor.Row+1:])
	e.Buffer.rows[e.Cursor.Row+1] = NewRow("")
	e.Cursor.Row += 1
	e.Cursor.Col = 0
}

func (e *Editor) MoveCursorToStartOfLine() {
	e.Cursor.Col = 0
}

func (e *Editor) MoveCursorToStartOfLineBelowCursor() {
	e.Cursor.Col = 0
	e.Cursor.Row += 1
}

// editable

func (e *Editor) GetCursor() gott.Point {
	return e.Cursor
}

func (e *Editor) SetCursor(cursor gott.Point) {
	e.Cursor = cursor
}

func (e *Editor) ReplaceCharacterAtCursor(cursor gott.Point, c rune) rune {
	return e.Buffer.rows[cursor.Row].ReplaceChar(cursor.Col, c)
}

func (e *Editor) DeleteRowsAtCursor(multiplier int) string {
	deletedText := ""
	for i := 0; i < multiplier; i++ {
		row := e.Cursor.Row
		if row < e.Buffer.GetRowCount() {
			if i > 0 {
				deletedText += "\n"
			}
			deletedText += string(e.Buffer.rows[row].Text)
			e.Buffer.rows = append(e.Buffer.rows[0:row], e.Buffer.rows[row+1:]...)
		} else {
			break
		}
	}
	e.Cursor.Row = clipToRange(e.Cursor.Row, 0, e.Buffer.GetRowCount()-1)
	return deletedText
}

func (e *Editor) SetPasteBoard(text string, mode int) {
	e.pasteText = text
	e.pasteMode = mode
}

func (e *Editor) DeleteWordsAtCursor(multiplier int) string {
	deletedText := ""
	for i := 0; i < multiplier; i++ {
		if e.Buffer.GetRowCount() == 0 {
			break
		}
		// if the row is empty, delete the row...
		row := e.Cursor.Row
		col := e.Cursor.Col
		b := e.Buffer
		if col >= b.rows[row].Length() {
			position := e.Cursor.Row
			e.Buffer.rows = append(e.Buffer.rows[0:position], e.Buffer.rows[position+1:]...)
			deletedText += "\n"
			e.KeepCursorInRow()
		} else {
			// else do this...
			c := e.Buffer.rows[e.Cursor.Row].DeleteChar(e.Cursor.Col)
			deletedText += string(c)
			for {
				if e.Cursor.Col > e.Buffer.rows[e.Cursor.Row].Length()-1 {
					break
				}
				if c == ' ' {
					break
				}
				c = e.Buffer.rows[e.Cursor.Row].DeleteChar(e.Cursor.Col)
				deletedText += string(c)
			}
			if e.Cursor.Col > e.Buffer.rows[e.Cursor.Row].Length()-1 {
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
	deletedText := e.Buffer.DeleteCharacters(e.Cursor.Row, e.Cursor.Col, multiplier, undo)
	if e.Cursor.Col > e.Buffer.rows[e.Cursor.Row].Length()-1 {
		e.Cursor.Col--
	}
	if e.Cursor.Col < 0 {
		e.Cursor.Col = 0
	}
	if finallyDeleteRow && e.Buffer.GetRowCount() > 0 {
		e.Buffer.DeleteRow(e.Cursor.Row)
	}
	return deletedText
}

func (e *Editor) InsertText(text string, position int) (gott.Point, int) {
	if e.Buffer.GetRowCount() == 0 {
		e.AppendBlankRow()
	}
	switch position {
	case gott.InsertAtCursor:
		break
	case gott.InsertAfterCursor:
		e.Cursor.Col++
		e.Cursor.Col = clipToRange(e.Cursor.Col, 0, e.Buffer.rows[e.Cursor.Row].Length())
	case gott.InsertAtStartOfLine:
		e.Cursor.Col = 0
	case gott.InsertAfterEndOfLine:
		e.Cursor.Col = e.Buffer.rows[e.Cursor.Row].Length()
	case gott.InsertAtNewLineBelowCursor:
		e.InsertLineBelowCursor()
	case gott.InsertAtNewLineAboveCursor:
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
		mode = gott.ModeEdit
	} else {
		mode = gott.ModeInsert
	}
	return e.Cursor, mode
}

func (e *Editor) SetInsertOperation(insert gott.InsertOperation) {
	e.insert = insert
}

func (e *Editor) GetPasteMode() int {
	return e.pasteMode
}

func (e *Editor) GetPasteText() string {
	return e.pasteText
}

func (e *Editor) ReverseCaseCharactersAtCursor(multiplier int) {
	if e.Buffer.GetRowCount() == 0 {
		return
	}
	row := &e.Buffer.rows[e.Cursor.Row]
	for i := 0; i < multiplier; i++ {
		c := row.Text[e.Cursor.Col]
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

func (e *Editor) PageUp() {
	// move to the top of the screen
	e.Cursor.Row = e.Offset.Rows
	// move up by a page
	for i := 0; i < e.size.Rows; i++ {
		e.MoveCursor(gott.MoveUp)
	}
}

func (e *Editor) PageDown() {
	// move to the bottom of the screen
	e.Cursor.Row = e.Offset.Rows + e.size.Rows - 1
	// move down by a page
	for i := 0; i < e.size.Rows; i++ {
		e.MoveCursor(gott.MoveDown)
	}
}

func (e *Editor) SetSize(s gott.Size) {
	e.size = s
}

func (e *Editor) CloseInsert() {
	e.insert.Close()
	e.insert = nil
}

func (e *Editor) MoveToBeginningOfLine() {
	e.Cursor.Col = 0
}

func (e *Editor) MoveToEndOfLine() {
	e.Cursor.Col = 0
	if e.Cursor.Row < e.Buffer.GetRowCount() {
		e.Cursor.Col = e.Buffer.GetRowLength(e.Cursor.Row) - 1
		if e.Cursor.Col < 0 {
			e.Cursor.Col = 0
		}
	}
}

func (e *Editor) GetBuffer() gott.Buffer {
	return e.Buffer
}

func (e *Editor) GetOffset() gott.Size {
	return e.Offset
}
