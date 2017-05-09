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

type Point struct {
	Row int
	Col int
}

type Size struct {
	Rows int
	Cols int
}

type Rect struct {
	Origin Point
	Size   Size
}

type Editor interface {
	// Set the size of the screen.
	SetSize(size Size)

	// Text being edited is stored in buffers.
	GetBuffer() Buffer

	// File operations.
	ReadFile(path string) error
	WriteFile(path string) error
	Bytes() []byte

	// Manage the cursor location.
	GetCursor() Point
	SetCursor(cursor Point)
	MoveCursor(direction int)
	MoveCursorToStartOfLine()
	MoveCursorToStartOfLineBelowCursor()
	MoveToBeginningOfLine()
	MoveToEndOfLine()
	KeepCursorInRow()
	PageUp()
	PageDown()

	// Recompute the display offset to keep the cursor onscreen.
	Scroll()
	GetOffset() Size

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
	CloseInsert()

	// Search.
	PerformSearch(text string)

	// Additional features.
	Gofmt(filename string, inputBytes []byte) (outputBytes []byte, err error)
}

type Buffer interface {
	// Read bytes into a buffer.
	ReadBytes(bytes []byte)

	// Buffer information.
	GetFileName() string
	GetRowCount() int
	TextAfter(row, col int) string

	// Draw the buffer contents.
	Render(origin Point, size Size, offset Size)
}

type Operation interface {
	// Perform an operation and return its inverse.
	Perform(e Editor, multiplier int) Operation
}

type InsertOperation interface {
	Operation
	AddCharacter(c rune)
	DeleteCharacter()
	Close()
	Length() int
}

type Commander interface {
	SetMode(int)
	GetMode() int
	GetSearchText() string
	GetCommand() string
	GetMessage() string
}
