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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"unicode"

	gott "github.com/timburks/gott/types"
)

// The Editor manages the editing of text in a Buffer.
type Editor struct {
	origin    gott.Point           // origin of editing area
	size      gott.Size            // size of editing area
	window    *Window              // active window being displayed
	windows   []*Window            // all windows being managed by the editor
	pasteText string               // used to cut/copy and paste
	pasteMode int                  // how to paste the string on the pasteboard
	previous  gott.Operation       // last operation performed, available to repeat
	undo      []gott.Operation     // stack of operations to undo
	insert    gott.InsertOperation // when in insert mode, the current insert operation
}

func NewEditor() *Editor {
	e := &Editor{}
	e.CreateWindow()
	e.window.buffer.ReadOnly = true // buffer zero is for command output
	e.window.buffer.Name = "*output*"
	return e
}

func (e *Editor) CreateWindow() *Window {
	e.window = NewWindow()
	e.windows = append(e.windows, e.window)
	return e.window
}

func (e *Editor) ListWindows() {
	var s string
	for i, window := range e.windows {
		if i > 0 {
			s += "\n"
		}
		s += fmt.Sprintf(" [%d] %s", window.number, window.buffer.Name)
	}
	listing := []byte(s)
	e.SelectWindow(0)
	e.window.buffer.LoadBytes(listing)
}

func (e *Editor) SelectWindow(number int) error {
	for _, window := range e.windows {
		if window.number == number {
			e.window = window
			return nil
		}
	}
	return errors.New(fmt.Sprintf("No window exists for identifier %d", number))
}

func (e *Editor) SelectWindowNext() error {
	next := e.window.number + 1
	if next < len(e.windows) {
		e.window = e.windows[next]
	}
	return nil
}

func (e *Editor) SelectWindowPrevious() error {
	prev := e.window.number - 1
	if prev >= 0 {
		e.window = e.windows[prev]
	}
	return nil
}

func (e *Editor) ReadFile(path string) error {
	// create a new buffer
	window := e.CreateWindow()
	window.buffer.SetFileName(path)
	// read the specified file into the buffer
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	window.buffer.LoadBytes(b)
	return nil
}

func (e *Editor) Bytes() []byte {
	return e.window.buffer.Bytes()
}

func (e *Editor) WriteFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	b := e.Bytes()
	if strings.HasSuffix(path, ".go") {
		out, err := e.Gofmt(e.window.buffer.GetFileName(), b)
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
	// if the current buffer is read only, don't perform any operations.
	if e.window.buffer.GetReadOnly() {
		return
	}
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
	if e.window.buffer.GetRowCount() == 0 {
		return
	}
	row := e.window.cursor.Row
	col := e.window.cursor.Col + 1

	for {
		var s string
		if col < e.window.buffer.GetRowLength(row) {
			s = e.window.buffer.TextAfter(row, col)
		} else {
			s = ""
		}
		i := strings.Index(s, text)
		if i != -1 {
			// found it
			e.window.cursor.Row = row
			e.window.cursor.Col = col + i
			return
		} else {
			col = 0
			row = row + 1
			if row == e.window.buffer.GetRowCount() {
				row = 0
			}
		}
		if row == e.window.cursor.Row {
			break
		}
	}
}

func (e *Editor) MoveCursor(direction int, multiplier int) {
	for i := 0; i < multiplier; i++ {
		switch direction {
		case gott.MoveLeft:
			if e.window.cursor.Col > 0 {
				e.window.cursor.Col--
			}
		case gott.MoveRight:
			if e.window.cursor.Row < e.window.buffer.GetRowCount() {
				rowLength := e.window.buffer.GetRowLength(e.window.cursor.Row)
				if e.window.cursor.Col < rowLength-1 {
					e.window.cursor.Col++
				}
			}
		case gott.MoveUp:
			if e.window.cursor.Row > 0 {
				e.window.cursor.Row--
			}
		case gott.MoveDown:
			if e.window.cursor.Row < e.window.buffer.GetRowCount()-1 {
				e.window.cursor.Row++
			}
		}
		// don't go past the end of the current line
		if e.window.cursor.Row < e.window.buffer.GetRowCount() {
			rowLength := e.window.buffer.GetRowLength(e.window.cursor.Row)
			if e.window.cursor.Col > rowLength-1 {
				e.window.cursor.Col = rowLength - 1
				if e.window.cursor.Col < 0 {
					e.window.cursor.Col = 0
				}
			}
		}
	}
}

func (e *Editor) MoveCursorForward() int {
	if e.window.cursor.Row < e.window.buffer.GetRowCount() {
		rowLength := e.window.buffer.GetRowLength(e.window.cursor.Row)
		if e.window.cursor.Col < rowLength-1 {
			e.window.cursor.Col++
			return gott.AtNextCharacter
		} else {
			e.window.cursor.Col = 0
			if e.window.cursor.Row+1 < e.window.buffer.GetRowCount() {
				e.window.cursor.Row++
				return gott.AtNextLine
			} else {
				return gott.AtEndOfFile
			}
		}
	} else {
		return gott.AtEndOfFile
	}
}

func (e *Editor) MoveCursorBackward() int {
	if e.window.cursor.Row < e.window.buffer.GetRowCount() {
		if e.window.cursor.Col > 0 {
			e.window.cursor.Col--
			return gott.AtNextCharacter
		} else {
			if e.window.cursor.Row > 0 {
				e.window.cursor.Row--
				rowLength := e.window.buffer.GetRowLength(e.window.cursor.Row)
				e.window.cursor.Col = rowLength - 1
				if e.window.cursor.Col < 0 {
					e.window.cursor.Col = 0
				}
				return gott.AtNextLine
			} else {
				return gott.AtEndOfFile
			}
		}
	} else {
		return gott.AtEndOfFile
	}
}

func isSpace(c rune) bool {
	return c == ' ' || c == rune(0)
}

func isAlphaNumeric(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_'
}

func isNonAlphaNumeric(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != ' ' && c != rune(0)
}

func (e *Editor) MoveCursorToNextWord(multiplier int) {
	for i := 0; i < multiplier; i++ {
		e.moveCursorToNextWord()
	}
}

func (e *Editor) moveCursorToNextWord() {
	c := e.window.buffer.GetCharacterAtCursor(e.window.cursor)
	if isSpace(c) { // if we're on a space, move to first non-space
		for isSpace(c) {
			if e.MoveCursorForward() != gott.AtNextCharacter {
				e.MoveForwardToFirstNonSpace()
				return
			}
			c = e.window.buffer.GetCharacterAtCursor(e.window.cursor)
		}
		return
	}
	if isAlphaNumeric(c) {
		// move past all letters/digits
		for isAlphaNumeric(c) {
			if e.MoveCursorForward() != gott.AtNextCharacter {
				e.MoveForwardToFirstNonSpace()
				return // we reached a new line or EOF
			}
			c = e.window.buffer.GetCharacterAtCursor(e.window.cursor)
		}
		// move past any spaces
		for isSpace(c) {
			if e.MoveCursorForward() != gott.AtNextCharacter {
				return // we reached a new line or EOF
			}
			c = e.window.buffer.GetCharacterAtCursor(e.window.cursor)
		}
	} else { // non-alphanumeric
		// move past all nonletters/digits
		for isNonAlphaNumeric(c) {
			if e.MoveCursorForward() != gott.AtNextCharacter {
				e.MoveForwardToFirstNonSpace()
				return // we reached a new line or EOF
			}
			c = e.window.buffer.GetCharacterAtCursor(e.window.cursor)
		}
		// move past any spaces
		for isSpace(c) {
			if e.MoveCursorForward() != gott.AtNextCharacter {
				return // we reached a new line or EOF
			}
			c = e.window.buffer.GetCharacterAtCursor(e.window.cursor)
		}
	}
}

func (e *Editor) MoveForwardToFirstNonSpace() {
	c := e.window.buffer.GetCharacterAtCursor(e.window.cursor)
	if c == ' ' { // if we're on a space, move to first non-space
		for c == ' ' {
			if e.MoveCursorForward() != gott.AtNextCharacter {
				return
			}
			c = e.window.buffer.GetCharacterAtCursor(e.window.cursor)
		}
		return
	}
}

func (e *Editor) MoveCursorBackToFirstNonSpace() int {
	// move back to first non-space (end of word)
	c := e.window.buffer.GetCharacterAtCursor(e.window.cursor)
	for isSpace(c) {
		p := e.MoveCursorBackward()
		if p != gott.AtNextCharacter {
			return p
		}
		c = e.window.buffer.GetCharacterAtCursor(e.window.cursor)
	}
	return gott.AtNextCharacter
}

func (e *Editor) MoveCursorBackBeforeCurrentWord() int {
	c := e.window.buffer.GetCharacterAtCursor(e.window.cursor)
	if isAlphaNumeric(c) {
		for isAlphaNumeric(c) {
			p := e.MoveCursorBackward()
			if p != gott.AtNextCharacter {
				return p
			}
			c = e.window.buffer.GetCharacterAtCursor(e.window.cursor)
		}
	} else if isNonAlphaNumeric(c) {
		for isNonAlphaNumeric(c) {
			p := e.MoveCursorBackward()
			if p != gott.AtNextCharacter {
				return p
			}
			c = e.window.buffer.GetCharacterAtCursor(e.window.cursor)
		}
	}
	return gott.AtNextCharacter
}

func (e *Editor) MoveCursorBackToStartOfCurrentWord() {
	c := e.window.buffer.GetCharacterAtCursor(e.window.cursor)
	if isSpace(c) {
		return
	}
	p := e.MoveCursorBackBeforeCurrentWord()
	if p != gott.AtEndOfFile {
		e.MoveCursorForward()
	}
}

func (e *Editor) MoveCursorToPreviousWord(multiplier int) {
	for i := 0; i < multiplier; i++ {
		e.moveCursorToPreviousWord()
	}
}

func (e *Editor) moveCursorToPreviousWord() {
	// get current character
	c := e.window.buffer.GetCharacterAtCursor(e.window.cursor)
	if isSpace(c) { // we started at a space
		e.MoveCursorBackToFirstNonSpace()
		e.MoveCursorBackToStartOfCurrentWord()
	} else {
		original := e.GetCursor()
		e.MoveCursorBackToStartOfCurrentWord()
		final := e.GetCursor()
		if original == final { // cursor didn't move
			e.MoveCursorBackBeforeCurrentWord()
			c = e.window.buffer.GetCharacterAtCursor(e.window.cursor)
			if c == rune(0) {
				return
			}
			e.MoveCursorBackToFirstNonSpace()
			e.MoveCursorBackToStartOfCurrentWord()
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
		e.window.cursor.Row++
		e.window.cursor.Col = 0
		return
	}
	// if the cursor is past the nmber of rows, add a row
	for e.window.cursor.Row >= e.window.buffer.GetRowCount() {
		e.AppendBlankRow()
	}
	e.window.buffer.InsertCharacter(e.window.cursor.Row, e.window.cursor.Col, c)
	e.window.cursor.Col += 1
}

func (e *Editor) InsertRow() {
	e.window.buffer.Highlighted = false
	if e.window.cursor.Row >= e.window.buffer.GetRowCount() {
		// we should never get here
		e.AppendBlankRow()
	} else {
		newRow := e.window.buffer.rows[e.window.cursor.Row].Split(e.window.cursor.Col)
		i := e.window.cursor.Row + 1
		// add a dummy row at the end of the Rows slice
		e.AppendBlankRow()
		// move rows to make room for the one we are adding
		copy(e.window.buffer.rows[i+1:], e.window.buffer.rows[i:])
		// add the new row
		e.window.buffer.rows[i] = newRow
	}
}

func (e *Editor) BackspaceChar() rune {
	if e.window.buffer.GetRowCount() == 0 {
		return rune(0)
	}
	if e.insert.Length() == 0 {
		return rune(0)
	}
	e.window.buffer.Highlighted = false
	e.insert.DeleteCharacter()
	if e.window.cursor.Col > 0 {
		c := e.window.buffer.rows[e.window.cursor.Row].DeleteChar(e.window.cursor.Col - 1)
		e.window.cursor.Col--
		return c
	} else if e.window.cursor.Row > 0 {
		// remove the current row and join it with the previous one
		oldRowText := e.window.buffer.rows[e.window.cursor.Row].Text
		var newCursor gott.Point
		newCursor.Col = len(e.window.buffer.rows[e.window.cursor.Row-1].Text)
		e.window.buffer.rows[e.window.cursor.Row-1].Text = append(e.window.buffer.rows[e.window.cursor.Row-1].Text, oldRowText...)
		e.window.buffer.rows = append(e.window.buffer.rows[0:e.window.cursor.Row], e.window.buffer.rows[e.window.cursor.Row+1:]...)
		e.window.cursor.Row--
		e.window.cursor.Col = newCursor.Col
		return rune('\n')
	} else {
		return rune(0)
	}
}

func (e *Editor) JoinRow(multiplier int) []gott.Point {
	if e.window.buffer.GetRowCount() == 0 {
		return nil
	}
	e.window.buffer.Highlighted = false
	// remove the next row and join it with this one
	insertions := make([]gott.Point, 0)
	for i := 0; i < multiplier; i++ {
		oldRowText := e.window.buffer.rows[e.window.cursor.Row+1].Text
		var newCursor gott.Point
		newCursor.Col = len(e.window.buffer.rows[e.window.cursor.Row].Text)
		e.window.buffer.rows[e.window.cursor.Row].Text = append(e.window.buffer.rows[e.window.cursor.Row].Text, oldRowText...)
		e.window.buffer.rows = append(e.window.buffer.rows[0:e.window.cursor.Row+1], e.window.buffer.rows[e.window.cursor.Row+2:]...)
		e.window.cursor.Col = newCursor.Col
		insertions = append(insertions, e.window.cursor)
	}
	return insertions
}

func (e *Editor) YankRow(multiplier int) {
	if e.window.buffer.GetRowCount() == 0 {
		return
	}
	pasteText := ""
	for i := 0; i < multiplier; i++ {
		position := e.window.cursor.Row + i
		if position < e.window.buffer.GetRowCount() {
			pasteText += string(e.window.buffer.rows[position].Text) + "\n"
		}
	}

	e.SetPasteBoard(pasteText, gott.PasteNewLine)
}

func (e *Editor) KeepCursorInRow() {
	if e.window.buffer.GetRowCount() == 0 {
		e.window.cursor.Col = 0
	} else {
		if e.window.cursor.Row >= e.window.buffer.GetRowCount() {
			e.window.cursor.Row = e.window.buffer.GetRowCount() - 1
		}
		if e.window.cursor.Row < 0 {
			e.window.cursor.Row = 0
		}
		lastIndexInRow := e.window.buffer.rows[e.window.cursor.Row].Length() - 1
		if e.window.cursor.Col > lastIndexInRow {
			e.window.cursor.Col = lastIndexInRow
		}
		if e.window.cursor.Col < 0 {
			e.window.cursor.Col = 0
		}
	}
}

func (e *Editor) AppendBlankRow() {
	e.window.buffer.rows = append(e.window.buffer.rows, NewRow(""))
}

func (e *Editor) InsertLineAboveCursor() {
	e.window.buffer.Highlighted = false
	e.AppendBlankRow()
	copy(e.window.buffer.rows[e.window.cursor.Row+1:], e.window.buffer.rows[e.window.cursor.Row:])
	e.window.buffer.rows[e.window.cursor.Row] = NewRow("")
	e.window.cursor.Col = 0
}

func (e *Editor) InsertLineBelowCursor() {
	e.window.buffer.Highlighted = false
	e.AppendBlankRow()
	copy(e.window.buffer.rows[e.window.cursor.Row+2:], e.window.buffer.rows[e.window.cursor.Row+1:])
	e.window.buffer.rows[e.window.cursor.Row+1] = NewRow("")
	e.window.cursor.Row += 1
	e.window.cursor.Col = 0
}

func (e *Editor) MoveCursorToStartOfLine() {
	e.window.cursor.Col = 0
}

func (e *Editor) MoveCursorToStartOfLineBelowCursor() {
	e.window.cursor.Col = 0
	e.window.cursor.Row += 1
}

// editable

func (e *Editor) GetCursor() gott.Point {
	return e.window.cursor
}

func (e *Editor) SetCursor(cursor gott.Point) {
	e.window.cursor = cursor
}

func (e *Editor) ReplaceCharacterAtCursor(cursor gott.Point, c rune) rune {
	e.window.buffer.Highlighted = false
	return e.window.buffer.rows[cursor.Row].ReplaceChar(cursor.Col, c)
}

func (e *Editor) DeleteRowsAtCursor(multiplier int) string {
	e.window.buffer.Highlighted = false
	deletedText := ""
	for i := 0; i < multiplier; i++ {
		row := e.window.cursor.Row
		if row < e.window.buffer.GetRowCount() {
			if i > 0 {
				deletedText += "\n"
			}
			deletedText += string(e.window.buffer.rows[row].Text)
			e.window.buffer.rows = append(e.window.buffer.rows[0:row], e.window.buffer.rows[row+1:]...)
		} else {
			break
		}
	}
	e.window.cursor.Row = clipToRange(e.window.cursor.Row, 0, e.window.buffer.GetRowCount()-1)
	return deletedText
}

func (e *Editor) SetPasteBoard(text string, mode int) {
	e.pasteText = text
	e.pasteMode = mode
}

func (e *Editor) DeleteWordsAtCursor(multiplier int) string {
	e.window.buffer.Highlighted = false
	deletedText := ""
	for i := 0; i < multiplier; i++ {
		if e.window.buffer.GetRowCount() == 0 {
			break
		}
		// if the row is empty, delete the row...
		row := e.window.cursor.Row
		col := e.window.cursor.Col
		b := e.window.buffer
		if col >= b.rows[row].Length() {
			position := e.window.cursor.Row
			e.window.buffer.rows = append(e.window.buffer.rows[0:position], e.window.buffer.rows[position+1:]...)
			deletedText += "\n"
			e.KeepCursorInRow()
		} else {
			// else do this...
			c := e.window.buffer.rows[e.window.cursor.Row].DeleteChar(e.window.cursor.Col)
			deletedText += string(c)
			for {
				if e.window.cursor.Col > e.window.buffer.rows[e.window.cursor.Row].Length()-1 {
					break
				}
				if c == ' ' {
					break
				}
				c = e.window.buffer.rows[e.window.cursor.Row].DeleteChar(e.window.cursor.Col)
				deletedText += string(c)
			}
			if e.window.cursor.Col > e.window.buffer.rows[e.window.cursor.Row].Length()-1 {
				e.window.cursor.Col--
			}
			if e.window.cursor.Col < 0 {
				e.window.cursor.Col = 0
			}
		}
	}
	return deletedText
}

func (e *Editor) DeleteCharactersAtCursor(multiplier int, undo bool, finallyDeleteRow bool) string {
	e.window.buffer.Highlighted = false
	deletedText := e.window.buffer.DeleteCharacters(e.window.cursor.Row, e.window.cursor.Col, multiplier, undo)
	if e.window.cursor.Col > e.window.buffer.rows[e.window.cursor.Row].Length()-1 {
		e.window.cursor.Col--
	}
	if e.window.cursor.Col < 0 {
		e.window.cursor.Col = 0
	}
	if finallyDeleteRow && e.window.buffer.GetRowCount() > 0 {
		e.window.buffer.DeleteRow(e.window.cursor.Row)
	}
	return deletedText
}

func (e *Editor) ChangeWordAtCursor(multiplier int, text string) (string, int) {
	e.window.buffer.Highlighted = false
	// delete the next N words and enter insert mode.
	deletedText := e.DeleteWordsAtCursor(multiplier)

	var mode int
	if text != "" { // repeat
		r := e.window.cursor.Row
		c := e.window.cursor.Col
		for _, c := range text {
			e.InsertChar(c)
		}
		e.window.cursor.Row = r
		e.window.cursor.Col = c
		mode = gott.ModeEdit
	} else {
		mode = gott.ModeInsert
	}

	return deletedText, mode
}

func (e *Editor) InsertText(text string, position int) (gott.Point, int) {
	e.window.buffer.Highlighted = false
	if e.window.buffer.GetRowCount() == 0 {
		e.AppendBlankRow()
	}
	switch position {
	case gott.InsertAtCursor:
		break
	case gott.InsertAfterCursor:
		e.window.cursor.Col++
		e.window.cursor.Col = clipToRange(e.window.cursor.Col, 0, e.window.buffer.rows[e.window.cursor.Row].Length())
	case gott.InsertAtStartOfLine:
		e.window.cursor.Col = 0
	case gott.InsertAfterEndOfLine:
		e.window.cursor.Col = e.window.buffer.rows[e.window.cursor.Row].Length()
	case gott.InsertAtNewLineBelowCursor:
		e.InsertLineBelowCursor()
	case gott.InsertAtNewLineAboveCursor:
		e.InsertLineAboveCursor()
	}
	var mode int
	if text != "" {
		r := e.window.cursor.Row
		c := e.window.cursor.Col
		for _, c := range text {
			e.InsertChar(c)
		}
		e.window.cursor.Row = r
		e.window.cursor.Col = c
		mode = gott.ModeEdit
	} else {
		mode = gott.ModeInsert
	}
	return e.window.cursor, mode
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
	if e.window.buffer.GetRowCount() == 0 {
		return
	}
	e.window.buffer.Highlighted = false
	row := e.window.buffer.rows[e.window.cursor.Row]
	for i := 0; i < multiplier; i++ {
		c := row.Text[e.window.cursor.Col]
		if unicode.IsUpper(c) {
			row.ReplaceChar(e.window.cursor.Col, unicode.ToLower(c))
		}
		if unicode.IsLower(c) {
			row.ReplaceChar(e.window.cursor.Col, unicode.ToUpper(c))
		}
		if e.window.cursor.Col < row.Length()-1 {
			e.window.cursor.Col++
		}
	}
}

func (e *Editor) PageUp(multiplier int) {
	// move to the top of the screen
	e.window.cursor.Row = e.window.offset.Rows
	for m := 0; m < multiplier; m++ {
		// move up by a page
		e.MoveCursor(gott.MoveUp, e.size.Rows)
	}
}

func (e *Editor) PageDown(multiplier int) {
	// move to the bottom of the screen
	e.window.cursor.Row = e.window.offset.Rows + e.size.Rows - 1
	for m := 0; m < multiplier; m++ {
		// move down by a page
		e.MoveCursor(gott.MoveDown, e.size.Rows)
	}
}

func (e *Editor) HalfPageUp(multiplier int) {
	// move to the top of the screen
	e.window.cursor.Row = e.window.offset.Rows
	for m := 0; m < multiplier; m++ {
		// move up by a half page
		e.MoveCursor(gott.MoveUp, e.size.Rows/2)
	}
}

func (e *Editor) HalfPageDown(multiplier int) {
	// move to the bottom of the screen
	e.window.cursor.Row = e.window.offset.Rows + e.size.Rows - 1
	for m := 0; m < multiplier; m++ {
		// move down by a half page
		e.MoveCursor(gott.MoveDown, e.size.Rows/2)
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
	e.window.cursor.Col = 0
}

func (e *Editor) MoveToEndOfLine() {
	e.window.cursor.Col = 0
	if e.window.cursor.Row < e.window.buffer.GetRowCount() {
		e.window.cursor.Col = e.window.buffer.GetRowLength(e.window.cursor.Row) - 1
		if e.window.cursor.Col < 0 {
			e.window.cursor.Col = 0
		}
	}
}

func (e *Editor) GetActiveWindow() gott.Window {
	return e.window
}

func (e *Editor) RenderEditWindows(d gott.Display) {
	// layout the visible buffers
	e.window.origin = gott.Point{Row: 0, Col: 0}
	e.window.size = gott.Size{Rows: e.size.Rows, Cols: e.size.Cols}
	// render the visible buffers
	e.window.Render(d)
	// the active buffer should set the cursor
	e.window.SetCursor(d)
}
