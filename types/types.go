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

package types

// The gott editor is modal and is always in one of these modes.
const (
	ModeEdit           = 0
	ModeInsert         = 1
	ModeCommand        = 2
	ModeLisp           = 3
	ModeSearchForward  = 4
	ModeSearchBackward = 5
	ModeQuit           = 9999
)

// These are the possible directions for cursor movement.
const (
	MoveUp    = 0
	MoveDown  = 1
	MoveRight = 2
	MoveLeft  = 3
)

// These describe positions after cursor movements.
// These are typically used to represent desired positions after automated movements.
const (
	AtNextCharacter = 0
	AtNextLine      = 1
	AtEndOfFile     = 2
)

// These specify different positions to begin inserting text in response to edit commands.
const (
	InsertAtCursor             = 0
	InsertAfterCursor          = 1
	InsertAtStartOfLine        = 2
	InsertAfterEndOfLine       = 3
	InsertAtNewLineBelowCursor = 4
	InsertAtNewLineAboveCursor = 5
)

// These specify different modes of pasting text.
// They are usually implied by the way the text was cut or copied.
const (
	PasteAtCursor = 0
	PasteNewLine  = 1
)

// These specify different types of words for use in the editor.
const (
	WordAlphaNumeric = 0
	WordPunctuation = 1
	WordSpace = 2
)

// A Point represents a cursor or character position in a buffer or a window.
type Point struct {
	Row int
	Col int
}

// A Size represents the size of a buffer, window, or screen.
type Size struct {
	Rows int
	Cols int
}

// A Rect represents a rectangular area, typically used to position windows.
type Rect struct {
	Origin Point
	Size   Size
}

// The Editor interface supports text editing in multiple windows.
type Editor interface {
	// Set the size of the screen.
	SetSize(size Size)

	// File operations.
	ReadFile(path string) error
	WriteFile(path string) error

	// Direct content manipulation
	Bytes() []byte
	LoadBytes([]byte)
	AppendBytes([]byte)

	// File information;
	GetFileName() string

	// Text being edited is displayed in windows.
	GetActiveWindow() Window
	SelectWindow(number int) error
	SelectWindowNext() error
	SelectWindowPrevious() error

	// Text being edited is stored in buffers.
	// Buffers can be displayed in any number of windows (including zero).
	ListWindows()

	// Manage the cursor location.
	GetCursor() Point
	SetCursor(cursor Point)
	MoveCursor(direction int, multiplier int)
	MoveCursorToNextWord(multiplier int)
	MoveCursorToPreviousWord(multiplier int)
	MoveCursorToStartOfLine()
	MoveCursorToStartOfLineBelowCursor()
	MoveToBeginningOfLine()
	MoveToEndOfLine()
	MoveCursorToLine(line int)
	KeepCursorInRow()
	PageUp(multiplier int)
	PageDown(multiplier int)
	HalfPageUp(multiplier int)
	HalfPageDown(multiplier int)

	// Low-level editing functions.
	ReplaceCharacterAtCursor(cursor Point, c rune) rune
	DeleteRowsAtCursor(multiplier int) string
	DeleteWordsAtCursor(multiplier int) string
	DeleteCharactersAtCursor(multiplier int, undo bool, finallyDeleteRow bool) string
	InsertChar(c rune)
	BackspaceChar() rune
	InsertText(text string, position int) (Point, int)
	ReverseCaseCharactersAtCursor(multiplier int)
	JoinRow(multiplier int) []Point
	ChangeWordAtCursor(multiplier int, text string) (string, int)

	// Cut/copy and paste support
	YankRow(multiplier int)
	SetPasteBoard(text string, mode int)
	GetPasteMode() int
	GetPasteText() string

	// Operations are the preferred way to make changes.
	// Operations are designed to be repeated and undone.
	Perform(op Operation, multiplier int)
	Repeat()
	PerformUndo()

	// When the editor is in insert mode, the Insert operation collects changes.
	SetInsertOperation(insert InsertOperation)
	GetInsertOperation() InsertOperation
	CloseInsert()

	// Search.
	PerformSearchForward(text string)
	PerformSearchBackward(text string)

	// Additional features.
	Gofmt(filename string, inputBytes []byte) (outputBytes []byte, err error)

	// Display
	LayoutWindows()
	RenderWindows(d Display)

	// Window Operations
	SplitWindowVertically()
	SplitWindowHorizontally()
	CloseActiveWindow()
}

// The Window interface supports text editing in a single focused window.
type Window interface {
	GetNumber() int
	GetName() string
	GetBuffer() Buffer
	GetParent() Window
	SetParent(w Window)

	GetCursor() Point
	SetCursor(cursor Point)

	SetCursorForDisplay(d Display)
	PerformSearchForward(text string)
	PerformSearchBackward(text string)
	MoveCursor(direction int, multiplier int)
	MoveCursorForward() int
	MoveCursorBackward() int
	MoveToBeginningOfLine()
	MoveToEndOfLine()
	MoveCursorToNextWord(multiplier int)
	MoveForwardToFirstNonSpace()
	MoveCursorBackToFirstNonSpace() int
	MoveCursorBackBeforeCurrentWord() int
	MoveCursorBackToStartOfCurrentWord()
	MoveCursorToPreviousWord(multiplier int)
	KeepCursorInRow()
	MoveCursorToStartOfLine()
	MoveCursorToStartOfLineBelowCursor()

	PageUp(multiplier int)
	PageDown(multiplier int)
	HalfPageUp(multiplier int)
	HalfPageDown(multiplier int)

	InsertChar(c rune)
	InsertRow()
	BackspaceChar() rune
	JoinRow(multiplier int) []Point
	YankRow(multiplier int)

	InsertText(text string, position int) (Point, int)
	ReverseCaseCharactersAtCursor(multiplier int)
	ReplaceCharacterAtCursor(cursor Point, c rune) rune
	DeleteRowsAtCursor(multiplier int) string

	DeleteWordsAtCursor(multiplier int) string
	DeleteCharactersAtCursor(multiplier int, undo bool, finallyDeleteRow bool) string
	ChangeWordAtCursor(multiplier int, text string) (string, int)

	// Display
	Layout(r Rect)
	Render(d Display)

	// Window Operations
	SplitVertically() (Window, Window)
	SplitHorizontally() (Window, Window)
	Close() Window
	GetWindowNext() Window
	GetWindowPrevious() Window
	FindWindow(int) Window
}

// The Buffer interface supports file-level text manipulation.
type Buffer interface {
	// Load bytes into a buffer returning the previous buffer contents.
	LoadBytes(bytes []byte) []byte

	// Append bytes to the end of a buffer.
	AppendBytes(bytes []byte)

	// Buffer information.
	GetName() string
	GetReadOnly() bool
	GetFileName() string
	GetRowCount() int
	GetBytes() []byte
	TextFromPosition(row, col int) string

	SetNameAndReadOnly(string, bool)
	SetFileName(string)
}

// The Highlighter interface supports text highlighting.
type Highlighter interface {
	// Perform syntax coloring on text in a buffer.
	Highlight(b *Buffer)
}

// The Operation interface supports repeatable, invertible operations.
type Operation interface {
	// Perform an operation and return its inverse.
	Perform(e Editor, multiplier int) Operation
}

// The InsertOperation interface supports insert operations that respond to user key commands.
type InsertOperation interface {
	Operation
	AddCharacter(c rune)
	DeleteCharacter()
	Close()
	Length() int
}

// The Commander interface supports user- and script-level control of an editor.
type Commander interface {
	SetMode(int)
	GetMessageBarText(length int) string
}

// Color represents a displayable color.
type Color uint16

// These are named colors for use in the Display interface.
const (
	ColorWhite = 0x08
	ColorBlack = 0x01
)

// The Display interface supports text and cursor display.
type Display interface {
	Close() 
	GetNextEvent() *Event
	Render(Editor, Commander)
	SetCell(j int, i int, c rune, color Color)
	SetCellReversed(j int, i int, c rune, color Color)
	SetCursor(position Point)
}

// These types of events can be generated by a Screen.
const (
	EventKey = iota
	EventResize
)

// Key represents a keystroke value.
type Key int16

// These are named key values that can be generated by a Screen..
const (
	KeyUnsupported = iota
	KeyArrowDown
	KeyArrowLeft
	KeyArrowRight
	KeyArrowUp
	KeyBackspace2
	KeyCtrlA
	KeyCtrlB
	KeyCtrlC
	KeyCtrlD
	KeyCtrlE
	KeyCtrlF
	KeyCtrlG
	KeyCtrlH
	KeyCtrlI
	KeyCtrlJ
	KeyCtrlK
	KeyCtrlL
	KeyCtrlM
	KeyCtrlN
	KeyCtrlO
	KeyCtrlP
	KeyCtrlQ
	KeyCtrlR
	KeyCtrlS
	KeyCtrlT
	KeyCtrlU
	KeyCtrlV
	KeyCtrlW
	KeyCtrlX
	KeyCtrlY
	KeyCtrlZ
	KeyEnd
	KeyEnter
	KeyEsc
	KeyHome
	KeyPgdn
	KeyPgup
	KeySpace
	KeyTab
)

// An Event represents user input events, typically keystrokes.
type Event struct {
	Type int
	Key  Key
	Ch   rune
}
