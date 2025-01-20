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
	"fmt"
	"unicode"

	gott "github.com/timburks/gott/pkg/types"
)

// This is the number of the last window created. Use it to uniquely number windows.
var lastWindowNumber = -1

// A Window instance manages a rectangular area onscreen.
// A window can be a view of a buffer or a container for two other windows.
// When a window contains text, it also has an associated cursor position.
// Many editing operations are implemented in Window to allow cursor management.
type Window struct {
	editor     gott.Editor
	number     int
	origin     gott.Point
	size       gott.Size
	cursor     gott.Point // cursor position
	offset     gott.Size  // display offset
	buffer     *Buffer    // a window either contains a buffer or child windows, but not both
	parent     *Window    // parent of window
	child1     *Window    // left/top child
	child2     *Window    // right/bottom child
	horizontal bool       // true if split is horizontal
}

func NewWindow(e gott.Editor) *Window {
	lastWindowNumber++
	w := &Window{}
	w.editor = e
	w.number = lastWindowNumber
	w.buffer = NewBuffer()
	return w
}

func (w *Window) Copy() *Window {
	newWindow := &Window{}
	*newWindow = *w
	return newWindow
}

func (w *Window) GetNumber() int {
	return w.number
}

func (w *Window) GetName() string {
	if w.buffer != nil {
		return w.buffer.Name
	} else {
		return "**"
	}
}

func (w *Window) GetBuffer() gott.Buffer {
	return w.buffer
}

func (w *Window) GetIndex() int {
	return w.number
}

func (w *Window) GetParent() gott.Window {
	return w.parent
}

func (w *Window) SetParent(p gott.Window) {
	w.parent = p.(*Window)
}

func (w *Window) Layout(r gott.Rect) {
	w.origin = r.Origin
	w.size = r.Size

	if w.buffer != nil {
		return
	}
	// adjust window sizes
	var r1, r2 gott.Rect
	if !w.horizontal {
		r1 = r
		r2 = r
		r1.Size.Rows = r.Size.Rows / 2
		r2.Size.Rows = r.Size.Rows - r1.Size.Rows
		r2.Origin.Row += r1.Size.Rows
	} else {
		borderWidth := 1
		r1 = r
		r2 = r
		r1.Size.Cols = r.Size.Cols / 2
		r2.Size.Cols = r.Size.Cols - r1.Size.Cols - borderWidth
		r2.Origin.Col += r1.Size.Cols + borderWidth
	}
	w.child1.Layout(r1)
	w.child2.Layout(r2)
}

func (w *Window) SplitVertically() (gott.Window, gott.Window) {

	// make two copies of this window
	w1 := w.Copy()
	w2 := w.Copy()

	// update the window numbers
	w.number = -1
	lastWindowNumber++
	w2.number = lastWindowNumber

	w1.parent = w
	w2.parent = w

	// nil out this buffer and set this window's contents
	w.buffer = nil
	w.child1 = w1
	w.child2 = w2
	w.horizontal = false

	w.Layout(gott.Rect{Origin: w.origin, Size: w.size})

	// return the new windows
	return w1, w2
}

func (w *Window) SplitHorizontally() (gott.Window, gott.Window) {
	// make two copies of this window
	w1 := w.Copy()
	w2 := w.Copy()

	// update the window numbers
	w.number = -1
	lastWindowNumber++
	w2.number = lastWindowNumber

	w1.parent = w
	w2.parent = w

	// nil out this buffer and set this window's contents
	w.buffer = nil
	w.child1 = w1
	w.child2 = w2

	w.horizontal = true
	w.Layout(gott.Rect{Origin: w.origin, Size: w.size})

	// return the new windows
	return w1, w2
}

func (w *Window) Close() gott.Window {
	parent := w.parent
	if parent == nil {
		return w
	}
	var replacement *Window
	if w == parent.child1 {
		replacement = parent.child2
	}
	if w == parent.child2 {
		replacement = parent.child1
	}
	if replacement != nil {
		parent.number = replacement.number
		parent.cursor = replacement.cursor
		parent.offset = replacement.offset
		parent.buffer = replacement.buffer
		parent.child1 = replacement.child1
		parent.child2 = replacement.child2
		parent.horizontal = replacement.horizontal
		if replacement.child1 != nil {
			parent.child1.parent = parent
		}
		if replacement.child2 != nil {
			parent.child2.parent = parent
		}
	}
	parent.Layout(gott.Rect{Origin: parent.origin, Size: parent.size})
	return parent.GetChildNext()
}

func (w *Window) FindWindow(number int) gott.Window {
	if w.number == number {
		return w
	}
	if w.child1 != nil {
		child := w.child1.FindWindow(number)
		if child != nil {
			return child
		}
	}
	if w.child2 != nil {
		child := w.child2.FindWindow(number)
		if child != nil {
			return child
		}
	}
	return nil
}

func (w *Window) GetWindowNext() gott.Window {
	parent := w.parent
	if parent == nil {
		return w.GetChildNext()
	}
	if w == parent.child1 {
		return parent.child2.GetChildNext()
	} else if w == parent.child2 {
		return parent.GetWindowNext()
	} else {
		return w.GetChildNext() // we should never get here
	}
}

func (w *Window) GetChildNext() gott.Window {
	if w.buffer != nil {
		return w
	}
	return w.child1.GetChildNext()
}

func (w *Window) GetWindowPrevious() gott.Window {
	parent := w.parent
	if parent == nil {
		return w.GetChildPrevious()
	}
	if w == parent.child2 {
		return parent.child1.GetChildPrevious()
	} else if w == parent.child1 {
		return parent.GetWindowPrevious()
	} else {
		return w.GetChildPrevious() // we should never get here
	}
}

func (w *Window) GetChildPrevious() gott.Window {
	if w.buffer != nil {
		return w
	}
	return w.child2.GetChildPrevious()
}

// draw text in an area defined by origin and size with a specified offset into the buffer
func (w *Window) Render(display gott.Display) {
	if w.buffer != nil {
		w.RenderBuffer(display)
	} else {
		if w.child1 != nil {
			w.child1.Render(display)
		}
		if w.child2 != nil {
			w.child2.Render(display)
			if w.horizontal {
				// Draw a vertical dividing bar
				col := w.child2.origin.Col - 1
				for row := w.origin.Row; row < w.origin.Row+w.size.Rows; row++ {
					display.SetCell(col, row, rune('.'), gott.ColorWhite)
				}
			}
		}
	}
}

func (w *Window) RenderBuffer(display gott.Display) {
	w.adjustDisplayOffsetForScrolling()

	b := w.buffer
	if !b.Highlighted {
		switch b.languageMode {
		case "go":
			h := NewGoHighlighter()
			h.Highlight(b)
		}
		b.Highlighted = true
	}

	for i := 0; i < w.size.Rows-1; i++ {
		var line string
		var colors []gott.Color
		if (i + w.offset.Rows) < len(b.rows) {
			line = b.rows[i+w.offset.Rows].GetString()
			colors = b.rows[i+w.offset.Rows].GetColors()
			if w.offset.Cols < len(line) {
				line = line[w.offset.Cols:]
				colors = colors[w.offset.Cols:]
			} else {
				line = ""
			}
		} else {
			line = "~"
			colors = make([]gott.Color, 1)
			colors[0] = gott.ColorWhite
		}
		// truncate line to fit screen
		if len(line) > w.size.Cols {
			line = line[0:w.size.Cols]
			colors = colors[0:w.size.Cols]
		}
		for j, c := range line {
			var color gott.Color = gott.ColorWhite
			if j < len(colors) {
				color = colors[j]
			}
			display.SetCell(j+w.origin.Col, i+w.origin.Row, rune(c), color)
		}
	}

	// Draw the info bar as a single line at the bottom of the buffer window.
	infoText := w.computeInfoBarText(w.size.Cols)
	infoRow := w.origin.Row + w.size.Rows - 1
	for x, ch := range infoText {
		display.SetCell(x+w.origin.Col, infoRow, rune(ch), gott.ColorWhite)
	}
}

// Compute the text to display on the info bar.
func (w *Window) computeInfoBarText(length int) string {
	b := w.buffer
	finalText := fmt.Sprintf(" %d/%d ", w.cursor.Row+1, b.GetRowCount())
	text := fmt.Sprintf("%d> %s ", w.GetIndex(), b.GetName())
	if b.GetReadOnly() {
		text = text + "(read-only) "
	}
	for len(text) <= length-len(finalText)-1 {
		text = text + "."
	}
	text += finalText
	return text
}

// Recompute the display offset to keep the cursor onscreen.
func (w *Window) adjustDisplayOffsetForScrolling() {
	if w.cursor.Row < w.offset.Rows {
		// scroll up
		w.offset.Rows = w.cursor.Row
	}
	// reserve the last row for the info bar
	textRows := w.size.Rows - 1
	if w.cursor.Row-w.offset.Rows >= textRows {
		// scroll down
		w.offset.Rows = w.cursor.Row - textRows + 1
	}
	if w.cursor.Col < w.offset.Cols {
		// scroll left
		w.offset.Cols = w.cursor.Col
	}
	if w.cursor.Col-w.offset.Cols >= w.size.Cols {
		// scroll right
		w.offset.Cols = w.cursor.Col - w.size.Cols + 1
	}
}

func (w *Window) GetCursor() gott.Point {
	return w.cursor
}

func (w *Window) SetCursor(cursor gott.Point) {
	w.cursor = cursor
}

func (w *Window) SetCursorForDisplay(d gott.Display) {
	d.SetCursor(gott.Point{
		Col: w.cursor.Col - w.offset.Cols + w.origin.Col,
		Row: w.cursor.Row - w.offset.Rows + w.origin.Row,
	})
}

func (w *Window) PerformSearchForward(text string) {
	if w.buffer.GetRowCount() == 0 {
		return
	}
	row := w.cursor.Row
	col := w.cursor.Col
	for {
		position := w.buffer.FirstPositionInRowAfterCol(row, col, text)
		if position != -1 {
			// found it
			w.cursor.Row = row
			w.cursor.Col = position
			return
		} else {
			col = -1
			row = row + 1
			if row == w.buffer.GetRowCount() {
				row = 0
			}
		}
		if row == w.cursor.Row {
			break
		}
	}
}

func (w *Window) PerformSearchBackward(text string) {
	if w.buffer.GetRowCount() == 0 {
		return
	}
	row := w.cursor.Row
	col := w.cursor.Col
	for {
		position := w.buffer.LastPositionInRowBeforeCol(row, col, text)
		if position != -1 {
			// found it
			w.cursor.Row = row
			w.cursor.Col = position
			return
		} else {
			row = row - 1
			if row < 0 {
				row = w.buffer.GetRowCount() - 1
			}
			col = w.buffer.GetRowLength(row)
			if col < 0 {
				col = 0
			}
		}
		if row == w.cursor.Row {
			break
		}
	}

}

func (w *Window) MoveCursor(direction int, multiplier int) {
	for i := 0; i < multiplier; i++ {
		switch direction {
		case gott.MoveLeft:
			if w.cursor.Col > 0 {
				w.cursor.Col--
			}
		case gott.MoveRight:
			if w.cursor.Row < w.buffer.GetRowCount() {
				rowLength := w.buffer.GetRowLength(w.cursor.Row)
				if w.cursor.Col < rowLength-1 {
					w.cursor.Col++
				}
			}
		case gott.MoveUp:
			if w.cursor.Row > 0 {
				w.cursor.Row--
			}
		case gott.MoveDown:
			if w.cursor.Row < w.buffer.GetRowCount()-1 {
				w.cursor.Row++
			}
		}
		// don't go past the end of the current line
		if w.cursor.Row < w.buffer.GetRowCount() {
			rowLength := w.buffer.GetRowLength(w.cursor.Row)
			if w.cursor.Col > rowLength-1 {
				w.cursor.Col = rowLength - 1
				if w.cursor.Col < 0 {
					w.cursor.Col = 0
				}
			}
		}
	}
}

func (w *Window) MoveCursorForward() int {
	if w.cursor.Row < w.buffer.GetRowCount() {
		rowLength := w.buffer.GetRowLength(w.cursor.Row)
		if w.cursor.Col < rowLength-1 {
			w.cursor.Col++
			return gott.AtNextCharacter
		} else {
			w.cursor.Col = 0
			if w.cursor.Row+1 < w.buffer.GetRowCount() {
				w.cursor.Row++
				return gott.AtNextLine
			} else {
				return gott.AtEndOfFile
			}
		}
	} else {
		return gott.AtEndOfFile
	}
}

func (w *Window) MoveCursorBackward() int {
	if w.cursor.Row < w.buffer.GetRowCount() {
		if w.cursor.Col > 0 {
			w.cursor.Col--
			return gott.AtNextCharacter
		} else {
			if w.cursor.Row > 0 {
				w.cursor.Row--
				rowLength := w.buffer.GetRowLength(w.cursor.Row)
				w.cursor.Col = rowLength - 1
				if w.cursor.Col < 0 {
					w.cursor.Col = 0
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

func (w *Window) MoveToBeginningOfLine() {
	w.cursor.Col = 0
}

func (w *Window) MoveToEndOfLine() {
	w.cursor.Col = 0
	if w.cursor.Row < w.buffer.GetRowCount() {
		w.cursor.Col = w.buffer.GetRowLength(w.cursor.Row) - 1
		if w.cursor.Col < 0 {
			w.cursor.Col = 0
		}
	}
}

func (w *Window) MoveCursorToNextWord(multiplier int) {
	for i := 0; i < multiplier; i++ {
		w.moveCursorToNextWord()
	}
}

func (w *Window) moveCursorToNextWord() {
	c := w.buffer.GetCharacterAtCursor(w.cursor)
	if isSpace(c) { // if we're on a space, move to first non-space
		for isSpace(c) {
			if w.MoveCursorForward() != gott.AtNextCharacter {
				w.MoveForwardToFirstNonSpace()
				return
			}
			c = w.buffer.GetCharacterAtCursor(w.cursor)
		}
		return
	}
	if isAlphaNumeric(c) {
		// move past all letters/digits
		for isAlphaNumeric(c) {
			if w.MoveCursorForward() != gott.AtNextCharacter {
				w.MoveForwardToFirstNonSpace()
				return // we reached a new line or EOF
			}
			c = w.buffer.GetCharacterAtCursor(w.cursor)
		}
		// move past any spaces
		for isSpace(c) {
			if w.MoveCursorForward() != gott.AtNextCharacter {
				return // we reached a new line or EOF
			}
			c = w.buffer.GetCharacterAtCursor(w.cursor)
		}
	} else { // non-alphanumeric
		// move past all nonletters/digits
		for isNonAlphaNumeric(c) {
			if w.MoveCursorForward() != gott.AtNextCharacter {
				w.MoveForwardToFirstNonSpace()
				return // we reached a new line or EOF
			}
			c = w.buffer.GetCharacterAtCursor(w.cursor)
		}
		// move past any spaces
		for isSpace(c) {
			if w.MoveCursorForward() != gott.AtNextCharacter {
				return // we reached a new line or EOF
			}
			c = w.buffer.GetCharacterAtCursor(w.cursor)
		}
	}
}

func (w *Window) MoveForwardToFirstNonSpace() {
	c := w.buffer.GetCharacterAtCursor(w.cursor)
	if c == ' ' { // if we're on a space, move to first non-space
		for c == ' ' {
			if w.MoveCursorForward() != gott.AtNextCharacter {
				return
			}
			c = w.buffer.GetCharacterAtCursor(w.cursor)
		}
		return
	}
}

func (w *Window) MoveCursorBackToFirstNonSpace() int {
	// move back to first non-space (end of word)
	c := w.buffer.GetCharacterAtCursor(w.cursor)
	for isSpace(c) {
		p := w.MoveCursorBackward()
		if p != gott.AtNextCharacter {
			return p
		}
		c = w.buffer.GetCharacterAtCursor(w.cursor)
	}
	return gott.AtNextCharacter
}

func (w *Window) MoveCursorBackBeforeCurrentWord() int {
	c := w.buffer.GetCharacterAtCursor(w.cursor)
	if isAlphaNumeric(c) {
		for isAlphaNumeric(c) {
			p := w.MoveCursorBackward()
			if p != gott.AtNextCharacter {
				return p
			}
			c = w.buffer.GetCharacterAtCursor(w.cursor)
		}
	} else if isNonAlphaNumeric(c) {
		for isNonAlphaNumeric(c) {
			p := w.MoveCursorBackward()
			if p != gott.AtNextCharacter {
				return p
			}
			c = w.buffer.GetCharacterAtCursor(w.cursor)
		}
	}
	return gott.AtNextCharacter
}

func (w *Window) MoveCursorBackToStartOfCurrentWord() {
	c := w.buffer.GetCharacterAtCursor(w.cursor)
	if isSpace(c) {
		return
	}
	p := w.MoveCursorBackBeforeCurrentWord()
	if p != gott.AtEndOfFile {
		w.MoveCursorForward()
	}
}

func (w *Window) MoveCursorToPreviousWord(multiplier int) {
	for i := 0; i < multiplier; i++ {
		w.moveCursorToPreviousWord()
	}
}

func (w *Window) moveCursorToPreviousWord() {
	// get current character
	c := w.buffer.GetCharacterAtCursor(w.cursor)
	if isSpace(c) { // we started at a space
		w.MoveCursorBackToFirstNonSpace()
		w.MoveCursorBackToStartOfCurrentWord()
	} else {
		original := w.GetCursor()
		w.MoveCursorBackToStartOfCurrentWord()
		final := w.GetCursor()
		if original == final { // cursor didn't move
			w.MoveCursorBackBeforeCurrentWord()
			c = w.buffer.GetCharacterAtCursor(w.cursor)
			if c == rune(0) {
				return
			}
			w.MoveCursorBackToFirstNonSpace()
			w.MoveCursorBackToStartOfCurrentWord()
		}
	}
}

func (w *Window) PageUp(multiplier int) {
	// move to the top of the screen
	w.cursor.Row = w.offset.Rows
	for m := 0; m < multiplier; m++ {
		// move up by a page
		w.MoveCursor(gott.MoveUp, w.size.Rows)
	}
}

func (w *Window) PageDown(multiplier int) {
	// move to the bottom of the screen
	w.cursor.Row = min(
		w.offset.Rows+w.size.Rows-1,
		w.buffer.GetRowCount()-1)
	for m := 0; m < multiplier; m++ {
		// move down by a page
		w.MoveCursor(gott.MoveDown, w.size.Rows)
	}
}

func (w *Window) HalfPageUp(multiplier int) {
	// move to the top of the screen
	w.cursor.Row = w.offset.Rows
	for m := 0; m < multiplier; m++ {
		// move up by a half page
		w.MoveCursor(gott.MoveUp, w.size.Rows/2)
	}
}

func (w *Window) HalfPageDown(multiplier int) {
	// move to the bottom of the screen
	w.cursor.Row = min(
		w.offset.Rows+w.size.Rows-1,
		w.buffer.GetRowCount()-1)
	for m := 0; m < multiplier; m++ {
		// move down by a half page
		w.MoveCursor(gott.MoveDown, w.size.Rows/2)
	}
}

func (w *Window) ReverseCaseCharactersAtCursor(multiplier int) {
	if w.buffer.GetRowCount() == 0 {
		return
	}
	w.buffer.Highlighted = false
	row := w.buffer.rows[w.cursor.Row]
	for i := 0; i < multiplier; i++ {
		c := row.GetText()[w.cursor.Col]
		if unicode.IsUpper(c) {
			row.ReplaceChar(w.cursor.Col, unicode.ToLower(c))
		}
		if unicode.IsLower(c) {
			row.ReplaceChar(w.cursor.Col, unicode.ToUpper(c))
		}
		if w.cursor.Col < row.Length()-1 {
			w.cursor.Col++
		}
	}
}

func (w *Window) InsertChar(c rune) {
	insert := w.editor.GetInsertOperation()
	if insert != nil {
		insert.AddCharacter(c)
	}
	if c == '\n' {
		w.InsertRow()
		w.cursor.Row++
		w.cursor.Col = 0
		return
	}
	// if the cursor is past the nmber of rows, add a row
	for w.cursor.Row >= w.buffer.GetRowCount() {
		w.AppendBlankRow()
	}
	w.buffer.InsertCharacter(w.cursor.Row, w.cursor.Col, c)
	w.cursor.Col += 1
}

func (w *Window) InsertRow() {
	w.buffer.Highlighted = false
	if w.cursor.Row >= w.buffer.GetRowCount() {
		// we should never get here
		w.AppendBlankRow()
	} else {
		newRow := w.buffer.rows[w.cursor.Row].Split(w.cursor.Col)
		i := w.cursor.Row + 1
		// add a dummy row at the end of the Rows slice
		w.AppendBlankRow()
		// move rows to make room for the one we are adding
		copy(w.buffer.rows[i+1:], w.buffer.rows[i:])
		// add the new row
		w.buffer.rows[i] = newRow
	}
}

func (w *Window) BackspaceChar() rune {
	insert := w.editor.GetInsertOperation()

	if w.buffer.GetRowCount() == 0 {
		return rune(0)
	}
	if insert.Length() == 0 {
		return rune(0)
	}
	w.buffer.Highlighted = false
	insert.DeleteCharacter()
	if w.cursor.Col > 0 {
		c := w.buffer.rows[w.cursor.Row].DeleteChar(w.cursor.Col - 1)
		w.cursor.Col--
		return c
	} else if w.cursor.Row > 0 {
		// remove the current row and join it with the previous one
		oldRowText := w.buffer.rows[w.cursor.Row].GetText()
		var newCursor gott.Point
		newCursor.Col = len(w.buffer.rows[w.cursor.Row-1].GetText())
		w.buffer.rows[w.cursor.Row-1].SetText(append(w.buffer.rows[w.cursor.Row-1].GetText(), oldRowText...))
		w.buffer.DeleteRow(w.cursor.Row)
		w.cursor.Row--
		w.cursor.Col = newCursor.Col
		return rune('\n')
	} else {
		return rune(0)
	}
}

func (w *Window) JoinRow(multiplier int) []gott.Point {
	if w.buffer.GetRowCount() == 0 {
		return nil
	}
	w.buffer.Highlighted = false
	// remove the next row and join it with this one
	insertions := make([]gott.Point, 0)
	for i := 0; i < multiplier; i++ {
		oldRowText := w.buffer.rows[w.cursor.Row+1].GetText()
		var newCursor gott.Point
		newCursor.Col = len(w.buffer.rows[w.cursor.Row].GetText())
		w.buffer.rows[w.cursor.Row].SetText(append(w.buffer.rows[w.cursor.Row].GetText(), oldRowText...))
		w.buffer.rows = append(w.buffer.rows[0:w.cursor.Row+1], w.buffer.rows[w.cursor.Row+2:]...)
		//w.buffer.DeleteRow(w.cursor.Row+1)
		w.cursor.Col = newCursor.Col
		insertions = append(insertions, w.cursor)
	}
	return insertions
}

func (w *Window) YankRow(multiplier int) {
	if w.buffer.GetRowCount() == 0 {
		return
	}
	pasteText := ""
	for i := 0; i < multiplier; i++ {
		position := w.cursor.Row + i
		if position < w.buffer.GetRowCount() {
			pasteText += string(w.buffer.rows[position].GetText()) + "\n"
		}
	}

	w.editor.SetPasteBoard(pasteText, gott.PasteNewLine)
}

func (w *Window) KeepCursorInRow() {
	if w.buffer.GetRowCount() == 0 {
		w.cursor.Col = 0
	} else {
		if w.cursor.Row >= w.buffer.GetRowCount() {
			w.cursor.Row = w.buffer.GetRowCount() - 1
		}
		if w.cursor.Row < 0 {
			w.cursor.Row = 0
		}
		lastIndexInRow := w.buffer.rows[w.cursor.Row].Length() - 1
		if w.cursor.Col > lastIndexInRow {
			w.cursor.Col = lastIndexInRow
		}
		if w.cursor.Col < 0 {
			w.cursor.Col = 0
		}
	}
}

func (w *Window) AppendBlankRow() {
	w.buffer.rows = append(w.buffer.rows, NewRow(""))
}

func (w *Window) InsertLineAboveCursor() {
	w.buffer.Highlighted = false
	w.AppendBlankRow()
	copy(w.buffer.rows[w.cursor.Row+1:], w.buffer.rows[w.cursor.Row:])
	w.buffer.rows[w.cursor.Row] = NewRow("")
	w.cursor.Col = 0
}

func (w *Window) InsertLineBelowCursor() {
	w.buffer.Highlighted = false
	w.AppendBlankRow()
	copy(w.buffer.rows[w.cursor.Row+2:], w.buffer.rows[w.cursor.Row+1:])
	w.buffer.rows[w.cursor.Row+1] = NewRow("")
	w.cursor.Row += 1
	w.cursor.Col = 0
}

func (w *Window) MoveCursorToStartOfLine() {
	w.cursor.Col = 0
}

func (w *Window) MoveCursorToStartOfLineBelowCursor() {
	w.cursor.Col = 0
	w.cursor.Row += 1
}

func (w *Window) ReplaceCharacterAtCursor(cursor gott.Point, c rune) rune {
	w.buffer.Highlighted = false
	return w.buffer.rows[cursor.Row].ReplaceChar(cursor.Col, c)
}

func (w *Window) DeleteRowsAtCursor(multiplier int) string {
	w.buffer.Highlighted = false
	deletedText := ""
	for i := 0; i < multiplier; i++ {
		row := w.cursor.Row
		if row < w.buffer.GetRowCount() {
			deletedText += string(w.buffer.rows[row].GetText())
			if row < w.buffer.GetRowCount()-1 {
				deletedText += "\n"
			}
			w.buffer.rows = append(w.buffer.rows[0:row], w.buffer.rows[row+1:]...)
		} else {
			break
		}
	}
	w.cursor.Row = clipToRange(w.cursor.Row, 0, w.buffer.GetRowCount()-1)
	return deletedText
}

func kindOfWord(c rune) int {
	if c == ' ' {
		return gott.WordSpace
	} else if isAlphaNumeric(c) {
		return gott.WordAlphaNumeric
	} else {
		return gott.WordPunctuation
	}
}

func (w *Window) DeleteWordsAtCursor(multiplier int) string {
	w.buffer.Highlighted = false
	deletedText := ""
	for i := 0; i < multiplier; i++ {
		if w.buffer.GetRowCount() == 0 {
			break
		}
		row := w.cursor.Row
		col := w.cursor.Col
		b := w.buffer
		if col >= b.rows[row].Length() {
			// if the row is empty, delete the row
			position := w.cursor.Row
			w.buffer.DeleteRow(position)
			deletedText += "\n"
			w.KeepCursorInRow()
		} else {
			// otherwise delete the next word
			c := w.buffer.rows[w.cursor.Row].DeleteChar(w.cursor.Col)
			deletedText += string(c)
			for {
				if w.cursor.Col > w.buffer.rows[w.cursor.Row].Length()-1 {
					break
				}
				if c == ' ' {
					break
				}
				c = w.buffer.rows[w.cursor.Row].DeleteChar(w.cursor.Col)
				deletedText += string(c)
			}
			if w.cursor.Col > w.buffer.rows[w.cursor.Row].Length()-1 {
				w.cursor.Col--
			}
			if w.cursor.Col < 0 {
				w.cursor.Col = 0
			}
		}
	}
	return deletedText
}

func (w *Window) DeleteCharactersAtCursor(multiplier int, undo bool, finallyDeleteRow bool) string {
	w.buffer.Highlighted = false
	deletedText := w.buffer.DeleteCharacters(w.cursor.Row, w.cursor.Col, multiplier, undo)
	if w.cursor.Col > w.buffer.rows[w.cursor.Row].Length()-1 {
		w.cursor.Col--
	}
	if w.cursor.Col < 0 {
		w.cursor.Col = 0
	}
	if finallyDeleteRow && w.buffer.GetRowCount() > 0 {
		w.buffer.DeleteRow(w.cursor.Row)
	}
	return deletedText
}

func (w *Window) ChangeWordAtCursor(multiplier int, text string) (string, int) {
	w.buffer.Highlighted = false
	// delete the next N words and enter insert mode.
	deletedText := w.DeleteWordsAtCursor(multiplier)

	var mode int
	if text != "" { // repeat
		r := w.cursor.Row
		c := w.cursor.Col
		for _, c := range text {
			w.InsertChar(c)
		}
		w.cursor.Row = r
		w.cursor.Col = c
		mode = gott.ModeEdit
	} else {
		mode = gott.ModeInsert
	}

	return deletedText, mode
}

func (w *Window) InsertText(text string, position int) (gott.Point, int) {
	w.buffer.Highlighted = false
	if w.buffer.GetRowCount() == 0 {
		w.AppendBlankRow()
	}
	switch position {
	case gott.InsertAtCursor:
		break
	case gott.InsertAfterCursor:
		w.cursor.Col++
		w.cursor.Col = clipToRange(w.cursor.Col, 0, w.buffer.rows[w.cursor.Row].Length())
	case gott.InsertAtStartOfLine:
		w.cursor.Col = 0
	case gott.InsertAfterEndOfLine:
		w.cursor.Col = w.buffer.rows[w.cursor.Row].Length()
	case gott.InsertAtNewLineBelowCursor:
		w.InsertLineBelowCursor()
	case gott.InsertAtNewLineAboveCursor:
		w.InsertLineAboveCursor()
	}
	var mode int
	if text != "" {
		r := w.cursor.Row
		c := w.cursor.Col
		for _, c := range text {
			w.InsertChar(c)
		}
		w.cursor.Row = r
		w.cursor.Col = c
		mode = gott.ModeEdit
	} else {
		mode = gott.ModeInsert
	}
	return w.cursor, mode
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
