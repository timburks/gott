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
	"log"
	"os"
	"sort"
	"strings"
	"unicode"

	gott "github.com/timburks/gott/pkg/types"
)

// The Editor manages text editing in associated buffers and windows.
// There is typically only one editor in a gott instance.
type Editor struct {
	origin          gott.Point           // origin of editing area
	size            gott.Size            // size of editing area
	focusedWindow   gott.Window          // window with cursor focus
	rootWindow      gott.Window          // root window for display
	documentWindows map[int]gott.Window  // all windows that contain documents; some may be offscreen
	pasteText       string               // used to cut/copy and paste
	pasteMode       int                  // how to paste the string on the pasteboard
	previous        gott.Operation       // last operation performed, available to repeat
	undo            []gott.Operation     // stack of operations to undo
	insert          gott.InsertOperation // when in insert mode, the current insert operation
}

func NewEditor() *Editor {
	e := &Editor{}
	e.documentWindows = make(map[int]gott.Window)
	w := e.CreateWindow()
	w.GetBuffer().SetNameAndReadOnly("*output*", true)
	e.rootWindow = w
	e.documentWindows[w.GetNumber()] = w
	return e
}

func (e *Editor) CreateWindow() gott.Window {
	e.focusedWindow = NewWindow(e)
	e.documentWindows[e.focusedWindow.GetNumber()] = e.focusedWindow
	return e.focusedWindow
}

func (e *Editor) ListWindows() {
	e.SelectWindow(0)
	var s string

	indices := make([]int, 0)
	for _, w := range e.documentWindows {
		indices = append(indices, w.GetNumber())
	}
	sort.Ints(indices)
	log.Printf("indices %+v", indices)
	for _, i := range indices {
		window := e.documentWindows[i]
		if window != nil {
			if s != "" {
				s += "\n"
			}
			s += fmt.Sprintf(" [%d] %s", i, window.GetName())
		} else {
			if s != "" {
				s += "\n"
			}
			s += fmt.Sprintf(" [%d] nil", i)
		}
	}
	listing := []byte(s)
	e.focusedWindow.GetBuffer().LoadBytes(listing)
}

func (e *Editor) SelectWindow(number int) error {
	// first look for an onscreen window
	w := e.rootWindow.FindWindow(number)
	if w != nil {
		// if we find an onscreen window, give it focus
		e.focusedWindow = w.(*Window)
		return nil
	}
	// next look for an offscreen window
	w = e.documentWindows[number]
	if w != nil {
		// replace the focused window with this one.
		removedWindow := e.focusedWindow
		parent := e.focusedWindow.GetParent().(*Window)
		if parent != nil {
			if parent.child1 == e.focusedWindow {
				parent.child1 = w.(*Window)
				w.(*Window).parent = parent
				e.focusedWindow = w.(*Window)
				e.LayoutWindows()
				e.PurgeIfOffscreenDuplicate(removedWindow.(*Window))
				return nil
			}
			if parent.child2 == e.focusedWindow {
				parent.child2 = w.(*Window)
				w.(*Window).parent = parent
				e.focusedWindow = w.(*Window)
				e.LayoutWindows()
				e.PurgeIfOffscreenDuplicate(removedWindow.(*Window))
				return nil
			}
		} else if e.rootWindow == e.focusedWindow {
			e.rootWindow = w.(*Window)
			e.focusedWindow = e.rootWindow
			e.focusedWindow.(*Window).parent = nil
			e.LayoutWindows()
			e.PurgeIfOffscreenDuplicate(removedWindow.(*Window))
			return nil
		} else {
			log.Printf("internal error in SelectWindow()")
		}
	}
	// if we get here, the window doesn't exist
	return fmt.Errorf("no window exists for identifier %d", number)
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
	b, err := os.ReadFile(path)
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

func (e *Editor) LoadBytes(b []byte) {
	e.GetActiveWindow().GetBuffer().LoadBytes(b)
}

func (e *Editor) AppendBytes(b []byte) {
	e.GetActiveWindow().GetBuffer().AppendBytes(b)
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

func (e *Editor) GetFileName() string {
	return e.GetActiveWindow().GetBuffer().GetFileName()
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

func (e *Editor) PerformSearchForward(text string) {
	e.focusedWindow.PerformSearchForward(text)
}

func (e *Editor) PerformSearchBackward(text string) {
	e.focusedWindow.PerformSearchBackward(text)
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

func (e *Editor) MoveCursorToLine(line int) {
	newRow := line - 1
	if newRow > e.GetActiveWindow().GetBuffer().GetRowCount()-1 {
		newRow = e.GetActiveWindow().GetBuffer().GetRowCount() - 1
	}
	if newRow < 0 {
		newRow = 0
	}
	cursor := e.GetCursor()
	cursor.Row = newRow
	cursor.Col = 0
	e.SetCursor(cursor)
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
	e.documentWindows[w1.GetNumber()] = w1
	e.documentWindows[w2.GetNumber()] = w2
	e.focusedWindow = w1
}

func (e *Editor) SplitWindowHorizontally() {
	w1, w2 := e.focusedWindow.SplitHorizontally()
	e.documentWindows[w1.GetNumber()] = w1
	e.documentWindows[w2.GetNumber()] = w2
	e.focusedWindow = w1
}

func (e *Editor) CloseActiveWindow() {
	removedWindow := e.focusedWindow.(*Window)
	e.focusedWindow = e.focusedWindow.Close()
	e.PurgeIfOffscreenDuplicate(removedWindow)
}

func (e *Editor) PurgeIfOffscreenDuplicate(w *Window) {
	// if the window is onscreen, return
	if e.rootWindow.FindWindow(w.GetNumber()) != nil {
		return
	}
	// is this window a duplicate?
	count := 0
	for _, w2 := range e.documentWindows {
		if w2.(*Window).buffer == w.buffer {
			count++
		}
	}
	if (count > 1) && (w.GetNumber() != 0) {
		delete(e.documentWindows, w.GetNumber())
	}
}
