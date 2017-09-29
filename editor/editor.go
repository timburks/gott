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
	origin        gott.Point           // origin of editing area
	size          gott.Size            // size of editing area
	focusedWindow gott.Window          // window with cursor focus
	rootWindow    gott.Window          // root window for display
	allWindows    []gott.Window        // all windows being managed by the editor
	pasteText     string               // used to cut/copy and paste
	pasteMode     int                  // how to paste the string on the pasteboard
	previous      gott.Operation       // last operation performed, available to repeat
	undo          []gott.Operation     // stack of operations to undo
	insert        gott.InsertOperation // when in insert mode, the current insert operation
}

func NewEditor() *Editor {
	e := &Editor{}
	w := e.CreateWindow()
	w.GetBuffer().SetNameAndReadOnly("*output*", true)
	e.rootWindow = w
	return e
}

func (e *Editor) CreateWindow() gott.Window {
	e.focusedWindow = NewWindow(e)
	e.allWindows = append(e.allWindows, e.focusedWindow)
	return e.focusedWindow
}

func (e *Editor) ListWindows() {
	var s string
	for i, window := range e.allWindows {
		if i > 0 {
			s += "\n"
		}
		s += fmt.Sprintf(" [%d] %s", window.GetNumber(), window.GetName())
	}
	listing := []byte(s)
	e.SelectWindow(0)
	e.focusedWindow.GetBuffer().LoadBytes(listing)
}

func (e *Editor) SelectWindow(number int) error {
	for _, window := range e.allWindows {
		if window.GetNumber() == number {
			e.focusedWindow = window
			e.rootWindow = window
			e.LayoutWindows()
			return nil
		}
	}
	return errors.New(fmt.Sprintf("No window exists for identifier %d", number))
}

func (e *Editor) SelectWindowNext() error {
	e.focusedWindow = e.focusedWindow.GetWindowNext()
	e.LayoutWindows()
	return nil
}

func (e *Editor) SelectWindowPrevious() error {
	e.focusedWindow = e.focusedWindow.GetWindowPrevious()
	e.LayoutWindows()
	return nil
}

func (e *Editor) ReadFile(path string) error {
	// create a new buffer
	window := e.CreateWindow()
	window.GetBuffer().SetFileName(path)
	// read the specified file into the buffer
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	window.GetBuffer().LoadBytes(b)

	e.rootWindow = window
	return nil
}

func (e *Editor) Bytes() []byte {
	return e.focusedWindow.GetBuffer().GetBytes()
}

func (e *Editor) WriteFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	b := e.Bytes()
	if strings.HasSuffix(path, ".go") {
		out, err := e.Gofmt(e.focusedWindow.GetBuffer().GetFileName(), b)
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
	if e.focusedWindow.GetBuffer().GetReadOnly() {
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
	e.focusedWindow.PerformSearch(text)
}

func (e *Editor) MoveCursor(direction int, multiplier int) {
	e.focusedWindow.MoveCursor(direction, multiplier)
}

func (e *Editor) MoveCursorForward() int {
	return e.focusedWindow.MoveCursorForward()
}

func (e *Editor) MoveCursorBackward() int {
	return e.focusedWindow.MoveCursorBackward()
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
	e.focusedWindow.MoveCursorToNextWord(multiplier)
}

func (e *Editor) MoveForwardToFirstNonSpace() {
	e.focusedWindow.MoveForwardToFirstNonSpace()
}

func (e *Editor) MoveCursorBackToFirstNonSpace() int {
	return e.focusedWindow.MoveCursorBackToFirstNonSpace()
}

func (e *Editor) MoveCursorBackBeforeCurrentWord() int {
	return e.focusedWindow.MoveCursorBackBeforeCurrentWord()
}

func (e *Editor) MoveCursorBackToStartOfCurrentWord() {
	e.focusedWindow.MoveCursorBackToStartOfCurrentWord()
}

func (e *Editor) MoveCursorToPreviousWord(multiplier int) {
	e.focusedWindow.MoveCursorToPreviousWord(multiplier)
}

// These editor primitives will make changes in insert mode and associate them with to the current operation.

func (e *Editor) InsertChar(c rune) {
	e.focusedWindow.InsertChar(c)
}

func (e *Editor) InsertRow() {
	e.focusedWindow.InsertRow()
}

func (e *Editor) BackspaceChar() rune {
	return e.focusedWindow.BackspaceChar()
}

func (e *Editor) JoinRow(multiplier int) []gott.Point {
	return e.focusedWindow.JoinRow(multiplier)
}

func (e *Editor) YankRow(multiplier int) {
	e.focusedWindow.YankRow(multiplier)
}

func (e *Editor) KeepCursorInRow() {
	e.focusedWindow.KeepCursorInRow()
}

/*
func (e *Editor) AppendBlankRow() {
	e.focusedWindow.AppendBlankRow()
}

func (e *Editor) InsertLineAboveCursor() {
	e.focusedWindow.InsertLineAboveCursor()
}

func (e *Editor) InsertLineBelowCursor() {
	e.focusedWindow.InsertLineBelowCursor()
}
*/
func (e *Editor) MoveCursorToStartOfLine() {
	e.focusedWindow.MoveCursorToStartOfLine()
}

func (e *Editor) MoveCursorToStartOfLineBelowCursor() {
	e.focusedWindow.MoveCursorToStartOfLineBelowCursor()
}

// editable

func (e *Editor) GetCursor() gott.Point {
	return e.focusedWindow.GetCursor()
}

func (e *Editor) SetCursor(cursor gott.Point) {
	e.focusedWindow.SetCursor(cursor)
}

func (e *Editor) ReplaceCharacterAtCursor(cursor gott.Point, c rune) rune {
	return e.focusedWindow.ReplaceCharacterAtCursor(cursor, c)
}

func (e *Editor) DeleteRowsAtCursor(multiplier int) string {
	return e.focusedWindow.DeleteRowsAtCursor(multiplier)
}

func (e *Editor) SetPasteBoard(text string, mode int) {
	e.pasteText = text
	e.pasteMode = mode
}

func (e *Editor) DeleteWordsAtCursor(multiplier int) string {
	return e.focusedWindow.DeleteWordsAtCursor(multiplier)
}

func (e *Editor) DeleteCharactersAtCursor(multiplier int, undo bool, finallyDeleteRow bool) string {
	return e.focusedWindow.DeleteCharactersAtCursor(multiplier, undo, finallyDeleteRow)
}

func (e *Editor) ChangeWordAtCursor(multiplier int, text string) (string, int) {
	return e.focusedWindow.ChangeWordAtCursor(multiplier, text)
}

func (e *Editor) InsertText(text string, position int) (gott.Point, int) {
	return e.focusedWindow.InsertText(text, position)
}

func (e *Editor) SetInsertOperation(insert gott.InsertOperation) {
	e.insert = insert
}

func (e *Editor) GetInsertOperation() gott.InsertOperation {
	return e.insert
}

func (e *Editor) GetPasteMode() int {
	return e.pasteMode
}

func (e *Editor) GetPasteText() string {
	return e.pasteText
}

func (e *Editor) ReverseCaseCharactersAtCursor(multiplier int) {
	e.focusedWindow.ReverseCaseCharactersAtCursor(multiplier)
}

func (e *Editor) PageUp(multiplier int) {
	e.focusedWindow.PageUp(multiplier)
}

func (e *Editor) PageDown(multiplier int) {
	e.focusedWindow.PageDown(multiplier)
}

func (e *Editor) HalfPageUp(multiplier int) {
	e.focusedWindow.HalfPageUp(multiplier)
}

func (e *Editor) HalfPageDown(multiplier int) {
	e.focusedWindow.HalfPageDown(multiplier)
}

func (e *Editor) SetSize(s gott.Size) {
	e.size = s
}

func (e *Editor) CloseInsert() {
	e.insert.Close()
	e.insert = nil
}

func (e *Editor) MoveToBeginningOfLine() {
	e.focusedWindow.MoveToBeginningOfLine()
}

func (e *Editor) MoveToEndOfLine() {
	e.focusedWindow.MoveToEndOfLine()
}

func (e *Editor) GetActiveWindow() gott.Window {
	return e.focusedWindow
}

func (e *Editor) LayoutWindows() {
	// layout the visible windows
	e.rootWindow.Layout(gott.Rect{
		Origin: e.origin,
		Size:   e.size,
	})
}

func (e *Editor) RenderWindows(d gott.Display) {
	// render the visible windows
	e.rootWindow.Render(d)
	// the focused window should set the cursor
	e.focusedWindow.SetCursorForDisplay(d)
}

func (e *Editor) SplitWindowVertically() {
	w1, w2 := e.focusedWindow.SplitVertically()
	e.allWindows = append(e.allWindows, w1)
	e.allWindows = append(e.allWindows, w2)
	e.focusedWindow = w1
}

func (e *Editor) SplitWindowHorizontally() {
	w1, w2 := e.focusedWindow.SplitHorizontally()
	e.allWindows = append(e.allWindows, w1)
	e.allWindows = append(e.allWindows, w2)
	e.focusedWindow = w1
}

func (e *Editor) CloseActiveWindow() {
	e.focusedWindow = e.focusedWindow.Close()
}
