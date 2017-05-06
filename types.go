package main

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
	MoveToStartOfLineBelowCursor()

	ReplaceCharacterAtCursor(cursor Point, c rune) rune
	DeleteRowsAtCursor(multiplier int) string
	DeleteWordsAtCursor(multiplier int) string
	DeleteCharactersAtCursor(multiplier int, undo bool, finallyDeleteRow bool) string
	InsertChar(c rune)
	InsertText(text string, position int) (Point, int)
	ReverseCaseCharactersAtCursor(multiplier int)

	SetPasteBoard(text string, mode int)
	SetInsertOperation(insert *Insert)
	GetPasteMode() int
	GetPasteText() string
}
