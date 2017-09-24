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
	origin         gott.Point           // origin of editing area
	size           gott.Size            // size of editing area
	focusedWindow  *Window              // window with cursor focus
	visibleWindows []*Window            // windows that are currently on screen
	allWindows     []*Window            // all windows being managed by the editor
	pasteText      string               // used to cut/copy and paste
	pasteMode      int                  // how to paste the string on the pasteboard
	previous       gott.Operation       // last operation performed, available to repeat
	undo           []gott.Operation     // stack of operations to undo
	insert         gott.InsertOperation // when in insert mode, the current insert operation
}

func NewEditor() *Editor {
	e := &Editor{}
	e.CreateWindow()
	e.focusedWindow.buffer.ReadOnly = true // buffer zero is for command output
	e.focusedWindow.buffer.Name = "*output*"
	return e
}

func (e *Editor) CreateWindow() *Window {
	e.focusedWindow = NewWindow()
	e.allWindows = append(e.allWindows, e.focusedWindow)
	return e.focusedWindow
}

func (e *Editor) ListWindows() {
	var s string
	for i, window := range e.allWindows {
		if i > 0 {
			s += "\n"
		}
		s += fmt.Sprintf(" [%d] %s", window.number, window.buffer.Name)
	}
	listing := []byte(s)
	e.SelectWindow(0)
	e.focusedWindow.buffer.LoadBytes(listing)
}

func (e *Editor) SelectWindow(number int) error {
	for _, window := range e.allWindows {
		if window.number == number {
			e.focusedWindow = window
			return nil
		}
	}
	return errors.New(fmt.Sprintf("No window exists for identifier %d", number))
}

func (e *Editor) SelectWindowNext() error {
	next := e.focusedWindow.number + 1
	if next < len(e.allWindows) {
		e.focusedWindow = e.allWindows[next]
	}
	return nil
}

func (e *Editor) SelectWindowPrevious() error {
	prev := e.focusedWindow.number - 1
	if prev >= 0 {
		e.focusedWindow = e.allWindows[prev]
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
	return e.focusedWindow.buffer.Bytes()
}

func (e *Editor) WriteFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	b := e.Bytes()
	if strings.HasSuffix(path, ".go") {
		out, err := e.Gofmt(e.focusedWindow.buffer.GetFileName(), b)
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
	if e.focusedWindow.buffer.GetReadOnly() {
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
	if e.focusedWindow.buffer.GetRowCount() == 0 {
		return
	}
	row := e.focusedWindow.cursor.Row
	col := e.focusedWindow.cursor.Col + 1

	for {
		var s string
		if col < e.focusedWindow.buffer.GetRowLength(row) {
			s = e.focusedWindow.buffer.TextAfter(row, col)
		} else {
			s = ""
		}
		i := strings.Index(s, text)
		if i != -1 {
			// found it
			e.focusedWindow.cursor.Row = row
			e.focusedWindow.cursor.Col = col + i
			return
		} else {
			col = 0
			row = row + 1
			if row == e.focusedWindow.buffer.GetRowCount() {
				row = 0
			}
		}
		if row == e.focusedWindow.cursor.Row {
			break
		}
	}
}

func (e *Editor) MoveCursor(direction int, multiplier int) {
	for i := 0; i < multiplier; i++ {
		switch direction {
		case gott.MoveLeft:
			if e.focusedWindow.cursor.Col > 0 {
				e.focusedWindow.cursor.Col--
			}
		case gott.MoveRight:
			if e.focusedWindow.cursor.Row < e.focusedWindow.buffer.GetRowCount() {
				rowLength := e.focusedWindow.buffer.GetRowLength(e.focusedWindow.cursor.Row)
				if e.focusedWindow.cursor.Col < rowLength-1 {
					e.focusedWindow.cursor.Col++
				}
			}
		case gott.MoveUp:
			if e.focusedWindow.cursor.Row > 0 {
				e.focusedWindow.cursor.Row--
			}
		case gott.MoveDown:
			if e.focusedWindow.cursor.Row < e.focusedWindow.buffer.GetRowCount()-1 {
				e.focusedWindow.cursor.Row++
			}
		}
		// don't go past the end of the current line
		if e.focusedWindow.cursor.Row < e.focusedWindow.buffer.GetRowCount() {
			rowLength := e.focusedWindow.buffer.GetRowLength(e.focusedWindow.cursor.Row)
			if e.focusedWindow.cursor.Col > rowLength-1 {
				e.focusedWindow.cursor.Col = rowLength - 1
				if e.focusedWindow.cursor.Col < 0 {
					e.focusedWindow.cursor.Col = 0
				}
			}
		}
	}
}

func (e *Editor) MoveCursorForward() int {
	if e.focusedWindow.cursor.Row < e.focusedWindow.buffer.GetRowCount() {
		rowLength := e.focusedWindow.buffer.GetRowLength(e.focusedWindow.cursor.Row)
		if e.focusedWindow.cursor.Col < rowLength-1 {
			e.focusedWindow.cursor.Col++
			return gott.AtNextCharacter
		} else {
			e.focusedWindow.cursor.Col = 0
			if e.focusedWindow.cursor.Row+1 < e.focusedWindow.buffer.GetRowCount() {
				e.focusedWindow.cursor.Row++
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
	if e.focusedWindow.cursor.Row < e.focusedWindow.buffer.GetRowCount() {
		if e.focusedWindow.cursor.Col > 0 {
			e.focusedWindow.cursor.Col--
			return gott.AtNextCharacter
		} else {
			if e.focusedWindow.cursor.Row > 0 {
				e.focusedWindow.cursor.Row--
				rowLength := e.focusedWindow.buffer.GetRowLength(e.focusedWindow.cursor.Row)
				e.focusedWindow.cursor.Col = rowLength - 1
				if e.focusedWindow.cursor.Col < 0 {
					e.focusedWindow.cursor.Col = 0
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
	c := e.focusedWindow.buffer.GetCharacterAtCursor(e.focusedWindow.cursor)
	if isSpace(c) { // if we're on a space, move to first non-space
		for isSpace(c) {
			if e.MoveCursorForward() != gott.AtNextCharacter {
				e.MoveForwardToFirstNonSpace()
				return
			}
			c = e.focusedWindow.buffer.GetCharacterAtCursor(e.focusedWindow.cursor)
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
			c = e.focusedWindow.buffer.GetCharacterAtCursor(e.focusedWindow.cursor)
		}
		// move past any spaces
		for isSpace(c) {
			if e.MoveCursorForward() != gott.AtNextCharacter {
				return // we reached a new line or EOF
			}
			c = e.focusedWindow.buffer.GetCharacterAtCursor(e.focusedWindow.cursor)
		}
	} else { // non-alphanumeric
		// move past all nonletters/digits
		for isNonAlphaNumeric(c) {
			if e.MoveCursorForward() != gott.AtNextCharacter {
				e.MoveForwardToFirstNonSpace()
				return // we reached a new line or EOF
			}
			c = e.focusedWindow.buffer.GetCharacterAtCursor(e.focusedWindow.cursor)
		}
		// move past any spaces
		for isSpace(c) {
			if e.MoveCursorForward() != gott.AtNextCharacter {
				return // we reached a new line or EOF
			}
			c = e.focusedWindow.buffer.GetCharacterAtCursor(e.focusedWindow.cursor)
		}
	}
}

func (e *Editor) MoveForwardToFirstNonSpace() {
	c := e.focusedWindow.buffer.GetCharacterAtCursor(e.focusedWindow.cursor)
	if c == ' ' { // if we're on a space, move to first non-space
		for c == ' ' {
			if e.MoveCursorForward() != gott.AtNextCharacter {
				return
			}
			c = e.focusedWindow.buffer.GetCharacterAtCursor(e.focusedWindow.cursor)
		}
		return
	}
}

func (e *Editor) MoveCursorBackToFirstNonSpace() int {
	// move back to first non-space (end of word)
	c := e.focusedWindow.buffer.GetCharacterAtCursor(e.focusedWindow.cursor)
	for isSpace(c) {
		p := e.MoveCursorBackward()
		if p != gott.AtNextCharacter {
			return p
		}
		c = e.focusedWindow.buffer.GetCharacterAtCursor(e.focusedWindow.cursor)
	}
	return gott.AtNextCharacter
}

func (e *Editor) MoveCursorBackBeforeCurrentWord() int {
	c := e.focusedWindow.buffer.GetCharacterAtCursor(e.focusedWindow.cursor)
	if isAlphaNumeric(c) {
		for isAlphaNumeric(c) {
			p := e.MoveCursorBackward()
			if p != gott.AtNextCharacter {
				return p
			}
			c = e.focusedWindow.buffer.GetCharacterAtCursor(e.focusedWindow.cursor)
		}
	} else if isNonAlphaNumeric(c) {
		for isNonAlphaNumeric(c) {
			p := e.MoveCursorBackward()
			if p != gott.AtNextCharacter {
				return p
			}
			c = e.focusedWindow.buffer.GetCharacterAtCursor(e.focusedWindow.cursor)
		}
	}
	return gott.AtNextCharacter
}

func (e *Editor) MoveCursorBackToStartOfCurrentWord() {
	c := e.focusedWindow.buffer.GetCharacterAtCursor(e.focusedWindow.cursor)
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
	c := e.focusedWindow.buffer.GetCharacterAtCursor(e.focusedWindow.cursor)
	if isSpace(c) { // we started at a space
		e.MoveCursorBackToFirstNonSpace()
		e.MoveCursorBackToStartOfCurrentWord()
	} else {
		original := e.GetCursor()
		e.MoveCursorBackToStartOfCurrentWord()
		final := e.GetCursor()
		if original == final { // cursor didn't move
			e.MoveCursorBackBeforeCurrentWord()
			c = e.focusedWindow.buffer.GetCharacterAtCursor(e.focusedWindow.cursor)
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
		e.focusedWindow.cursor.Row++
		e.focusedWindow.cursor.Col = 0
		return
	}
	// if the cursor is past the nmber of rows, add a row
	for e.focusedWindow.cursor.Row >= e.focusedWindow.buffer.GetRowCount() {
		e.AppendBlankRow()
	}
	e.focusedWindow.buffer.InsertCharacter(e.focusedWindow.cursor.Row, e.focusedWindow.cursor.Col, c)
	e.focusedWindow.cursor.Col += 1
}

func (e *Editor) InsertRow() {
	e.focusedWindow.buffer.Highlighted = false
	if e.focusedWindow.cursor.Row >= e.focusedWindow.buffer.GetRowCount() {
		// we should never get here
		e.AppendBlankRow()
	} else {
		newRow := e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row].Split(e.focusedWindow.cursor.Col)
		i := e.focusedWindow.cursor.Row + 1
		// add a dummy row at the end of the Rows slice
		e.AppendBlankRow()
		// move rows to make room for the one we are adding
		copy(e.focusedWindow.buffer.rows[i+1:], e.focusedWindow.buffer.rows[i:])
		// add the new row
		e.focusedWindow.buffer.rows[i] = newRow
	}
}

func (e *Editor) BackspaceChar() rune {
	if e.focusedWindow.buffer.GetRowCount() == 0 {
		return rune(0)
	}
	if e.insert.Length() == 0 {
		return rune(0)
	}
	e.focusedWindow.buffer.Highlighted = false
	e.insert.DeleteCharacter()
	if e.focusedWindow.cursor.Col > 0 {
		c := e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row].DeleteChar(e.focusedWindow.cursor.Col - 1)
		e.focusedWindow.cursor.Col--
		return c
	} else if e.focusedWindow.cursor.Row > 0 {
		// remove the current row and join it with the previous one
		oldRowText := e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row].Text
		var newCursor gott.Point
		newCursor.Col = len(e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row-1].Text)
		e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row-1].Text = append(e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row-1].Text, oldRowText...)
		e.focusedWindow.buffer.rows = append(e.focusedWindow.buffer.rows[0:e.focusedWindow.cursor.Row], e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row+1:]...)
		e.focusedWindow.cursor.Row--
		e.focusedWindow.cursor.Col = newCursor.Col
		return rune('\n')
	} else {
		return rune(0)
	}
}

func (e *Editor) JoinRow(multiplier int) []gott.Point {
	if e.focusedWindow.buffer.GetRowCount() == 0 {
		return nil
	}
	e.focusedWindow.buffer.Highlighted = false
	// remove the next row and join it with this one
	insertions := make([]gott.Point, 0)
	for i := 0; i < multiplier; i++ {
		oldRowText := e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row+1].Text
		var newCursor gott.Point
		newCursor.Col = len(e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row].Text)
		e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row].Text = append(e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row].Text, oldRowText...)
		e.focusedWindow.buffer.rows = append(e.focusedWindow.buffer.rows[0:e.focusedWindow.cursor.Row+1], e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row+2:]...)
		e.focusedWindow.cursor.Col = newCursor.Col
		insertions = append(insertions, e.focusedWindow.cursor)
	}
	return insertions
}

func (e *Editor) YankRow(multiplier int) {
	if e.focusedWindow.buffer.GetRowCount() == 0 {
		return
	}
	pasteText := ""
	for i := 0; i < multiplier; i++ {
		position := e.focusedWindow.cursor.Row + i
		if position < e.focusedWindow.buffer.GetRowCount() {
			pasteText += string(e.focusedWindow.buffer.rows[position].Text) + "\n"
		}
	}

	e.SetPasteBoard(pasteText, gott.PasteNewLine)
}

func (e *Editor) KeepCursorInRow() {
	if e.focusedWindow.buffer.GetRowCount() == 0 {
		e.focusedWindow.cursor.Col = 0
	} else {
		if e.focusedWindow.cursor.Row >= e.focusedWindow.buffer.GetRowCount() {
			e.focusedWindow.cursor.Row = e.focusedWindow.buffer.GetRowCount() - 1
		}
		if e.focusedWindow.cursor.Row < 0 {
			e.focusedWindow.cursor.Row = 0
		}
		lastIndexInRow := e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row].Length() - 1
		if e.focusedWindow.cursor.Col > lastIndexInRow {
			e.focusedWindow.cursor.Col = lastIndexInRow
		}
		if e.focusedWindow.cursor.Col < 0 {
			e.focusedWindow.cursor.Col = 0
		}
	}
}

func (e *Editor) AppendBlankRow() {
	e.focusedWindow.buffer.rows = append(e.focusedWindow.buffer.rows, NewRow(""))
}

func (e *Editor) InsertLineAboveCursor() {
	e.focusedWindow.buffer.Highlighted = false
	e.AppendBlankRow()
	copy(e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row+1:], e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row:])
	e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row] = NewRow("")
	e.focusedWindow.cursor.Col = 0
}

func (e *Editor) InsertLineBelowCursor() {
	e.focusedWindow.buffer.Highlighted = false
	e.AppendBlankRow()
	copy(e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row+2:], e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row+1:])
	e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row+1] = NewRow("")
	e.focusedWindow.cursor.Row += 1
	e.focusedWindow.cursor.Col = 0
}

func (e *Editor) MoveCursorToStartOfLine() {
	e.focusedWindow.cursor.Col = 0
}

func (e *Editor) MoveCursorToStartOfLineBelowCursor() {
	e.focusedWindow.cursor.Col = 0
	e.focusedWindow.cursor.Row += 1
}

// editable

func (e *Editor) GetCursor() gott.Point {
	return e.focusedWindow.cursor
}

func (e *Editor) SetCursor(cursor gott.Point) {
	e.focusedWindow.cursor = cursor
}

func (e *Editor) ReplaceCharacterAtCursor(cursor gott.Point, c rune) rune {
	e.focusedWindow.buffer.Highlighted = false
	return e.focusedWindow.buffer.rows[cursor.Row].ReplaceChar(cursor.Col, c)
}

func (e *Editor) DeleteRowsAtCursor(multiplier int) string {
	e.focusedWindow.buffer.Highlighted = false
	deletedText := ""
	for i := 0; i < multiplier; i++ {
		row := e.focusedWindow.cursor.Row
		if row < e.focusedWindow.buffer.GetRowCount() {
			if i > 0 {
				deletedText += "\n"
			}
			deletedText += string(e.focusedWindow.buffer.rows[row].Text)
			e.focusedWindow.buffer.rows = append(e.focusedWindow.buffer.rows[0:row], e.focusedWindow.buffer.rows[row+1:]...)
		} else {
			break
		}
	}
	e.focusedWindow.cursor.Row = clipToRange(e.focusedWindow.cursor.Row, 0, e.focusedWindow.buffer.GetRowCount()-1)
	return deletedText
}

func (e *Editor) SetPasteBoard(text string, mode int) {
	e.pasteText = text
	e.pasteMode = mode
}

func (e *Editor) DeleteWordsAtCursor(multiplier int) string {
	e.focusedWindow.buffer.Highlighted = false
	deletedText := ""
	for i := 0; i < multiplier; i++ {
		if e.focusedWindow.buffer.GetRowCount() == 0 {
			break
		}
		// if the row is empty, delete the row...
		row := e.focusedWindow.cursor.Row
		col := e.focusedWindow.cursor.Col
		b := e.focusedWindow.buffer
		if col >= b.rows[row].Length() {
			position := e.focusedWindow.cursor.Row
			e.focusedWindow.buffer.rows = append(e.focusedWindow.buffer.rows[0:position], e.focusedWindow.buffer.rows[position+1:]...)
			deletedText += "\n"
			e.KeepCursorInRow()
		} else {
			// else do this...
			c := e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row].DeleteChar(e.focusedWindow.cursor.Col)
			deletedText += string(c)
			for {
				if e.focusedWindow.cursor.Col > e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row].Length()-1 {
					break
				}
				if c == ' ' {
					break
				}
				c = e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row].DeleteChar(e.focusedWindow.cursor.Col)
				deletedText += string(c)
			}
			if e.focusedWindow.cursor.Col > e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row].Length()-1 {
				e.focusedWindow.cursor.Col--
			}
			if e.focusedWindow.cursor.Col < 0 {
				e.focusedWindow.cursor.Col = 0
			}
		}
	}
	return deletedText
}

func (e *Editor) DeleteCharactersAtCursor(multiplier int, undo bool, finallyDeleteRow bool) string {
	e.focusedWindow.buffer.Highlighted = false
	deletedText := e.focusedWindow.buffer.DeleteCharacters(e.focusedWindow.cursor.Row, e.focusedWindow.cursor.Col, multiplier, undo)
	if e.focusedWindow.cursor.Col > e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row].Length()-1 {
		e.focusedWindow.cursor.Col--
	}
	if e.focusedWindow.cursor.Col < 0 {
		e.focusedWindow.cursor.Col = 0
	}
	if finallyDeleteRow && e.focusedWindow.buffer.GetRowCount() > 0 {
		e.focusedWindow.buffer.DeleteRow(e.focusedWindow.cursor.Row)
	}
	return deletedText
}

func (e *Editor) ChangeWordAtCursor(multiplier int, text string) (string, int) {
	e.focusedWindow.buffer.Highlighted = false
	// delete the next N words and enter insert mode.
	deletedText := e.DeleteWordsAtCursor(multiplier)

	var mode int
	if text != "" { // repeat
		r := e.focusedWindow.cursor.Row
		c := e.focusedWindow.cursor.Col
		for _, c := range text {
			e.InsertChar(c)
		}
		e.focusedWindow.cursor.Row = r
		e.focusedWindow.cursor.Col = c
		mode = gott.ModeEdit
	} else {
		mode = gott.ModeInsert
	}

	return deletedText, mode
}

func (e *Editor) InsertText(text string, position int) (gott.Point, int) {
	e.focusedWindow.buffer.Highlighted = false
	if e.focusedWindow.buffer.GetRowCount() == 0 {
		e.AppendBlankRow()
	}
	switch position {
	case gott.InsertAtCursor:
		break
	case gott.InsertAfterCursor:
		e.focusedWindow.cursor.Col++
		e.focusedWindow.cursor.Col = clipToRange(e.focusedWindow.cursor.Col, 0, e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row].Length())
	case gott.InsertAtStartOfLine:
		e.focusedWindow.cursor.Col = 0
	case gott.InsertAfterEndOfLine:
		e.focusedWindow.cursor.Col = e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row].Length()
	case gott.InsertAtNewLineBelowCursor:
		e.InsertLineBelowCursor()
	case gott.InsertAtNewLineAboveCursor:
		e.InsertLineAboveCursor()
	}
	var mode int
	if text != "" {
		r := e.focusedWindow.cursor.Row
		c := e.focusedWindow.cursor.Col
		for _, c := range text {
			e.InsertChar(c)
		}
		e.focusedWindow.cursor.Row = r
		e.focusedWindow.cursor.Col = c
		mode = gott.ModeEdit
	} else {
		mode = gott.ModeInsert
	}
	return e.focusedWindow.cursor, mode
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
	if e.focusedWindow.buffer.GetRowCount() == 0 {
		return
	}
	e.focusedWindow.buffer.Highlighted = false
	row := e.focusedWindow.buffer.rows[e.focusedWindow.cursor.Row]
	for i := 0; i < multiplier; i++ {
		c := row.Text[e.focusedWindow.cursor.Col]
		if unicode.IsUpper(c) {
			row.ReplaceChar(e.focusedWindow.cursor.Col, unicode.ToLower(c))
		}
		if unicode.IsLower(c) {
			row.ReplaceChar(e.focusedWindow.cursor.Col, unicode.ToUpper(c))
		}
		if e.focusedWindow.cursor.Col < row.Length()-1 {
			e.focusedWindow.cursor.Col++
		}
	}
}

func (e *Editor) PageUp(multiplier int) {
	// move to the top of the screen
	e.focusedWindow.cursor.Row = e.focusedWindow.offset.Rows
	for m := 0; m < multiplier; m++ {
		// move up by a page
		e.MoveCursor(gott.MoveUp, e.size.Rows)
	}
}

func (e *Editor) PageDown(multiplier int) {
	// move to the bottom of the screen
	e.focusedWindow.cursor.Row = e.focusedWindow.offset.Rows + e.size.Rows - 1
	for m := 0; m < multiplier; m++ {
		// move down by a page
		e.MoveCursor(gott.MoveDown, e.size.Rows)
	}
}

func (e *Editor) HalfPageUp(multiplier int) {
	// move to the top of the screen
	e.focusedWindow.cursor.Row = e.focusedWindow.offset.Rows
	for m := 0; m < multiplier; m++ {
		// move up by a half page
		e.MoveCursor(gott.MoveUp, e.size.Rows/2)
	}
}

func (e *Editor) HalfPageDown(multiplier int) {
	// move to the bottom of the screen
	e.focusedWindow.cursor.Row = e.focusedWindow.offset.Rows + e.size.Rows - 1
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
	e.focusedWindow.cursor.Col = 0
}

func (e *Editor) MoveToEndOfLine() {
	e.focusedWindow.cursor.Col = 0
	if e.focusedWindow.cursor.Row < e.focusedWindow.buffer.GetRowCount() {
		e.focusedWindow.cursor.Col = e.focusedWindow.buffer.GetRowLength(e.focusedWindow.cursor.Row) - 1
		if e.focusedWindow.cursor.Col < 0 {
			e.focusedWindow.cursor.Col = 0
		}
	}
}

func (e *Editor) GetActiveWindow() gott.Window {
	return e.focusedWindow
}

func (e *Editor) LayoutWindows() {
	// layout the visible windows
	e.focusedWindow.origin = gott.Point{Row: 0, Col: 0}
	e.focusedWindow.size = gott.Size{Rows: e.size.Rows, Cols: e.size.Cols}
	// go back to a single window
	e.visibleWindows = make([]*Window, 0, 0)
	e.visibleWindows = append(e.visibleWindows, e.focusedWindow)
}

func (e *Editor) RenderWindows(d gott.Display) {

	// render the visible windows
	for _, window := range e.visibleWindows {
		window.Render(d)
	}

	// the active buffer should set the cursor
	e.focusedWindow.SetCursor(d)
}

func (e *Editor) SplitWindow() {
	// create a second window
	secondWindow := e.focusedWindow.Copy()
	e.allWindows = append(e.allWindows, secondWindow)
	e.visibleWindows = append(e.visibleWindows, secondWindow)

	// adjust window sizes
	height := e.focusedWindow.size.Rows
	newUpperHeight := height / 2
	newLowerHeight := height - newUpperHeight
	e.focusedWindow.size.Rows = newUpperHeight
	secondWindow.size.Rows = newLowerHeight
	secondWindow.origin.Row += newUpperHeight
}
