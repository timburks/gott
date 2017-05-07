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

type Editable interface {
	GetCursor() Point
	SetCursor(cursor Point)

	MoveCursorToStartOfLine()
	MoveCursorToStartOfLineBelowCursor()

	ReplaceCharacterAtCursor(cursor Point, c rune) rune
	DeleteRowsAtCursor(multiplier int) string
	DeleteWordsAtCursor(multiplier int) string
	DeleteCharactersAtCursor(multiplier int, undo bool, finallyDeleteRow bool) string
	InsertChar(c rune)
	InsertText(text string, position int) (Point, int)
	ReverseCaseCharactersAtCursor(multiplier int)

	SetPasteBoard(text string, mode int)
	SetInsertOperation(insert InsertOperation)
	GetPasteMode() int
	GetPasteText() string
}

type Operation interface {
	Perform(e Editable, multiplier int) Operation // performs the operation and returns its inverse
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
}
