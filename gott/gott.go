// Package gott is a top-level interface for embedding gott in gomobile apps.

package gott

type Editor struct {
	State  int
	Buffer Buffer
}

type Event struct {
	Code     int
	Key      int
	Modifier int
	X        int
	Y        int
}

type Buffer struct {
	Characters string
	Rows       int
	Cols       int
}

func NewEditor() *Editor {
	return &Editor{}
}

func (e *Editor)HandleEvent(ev *Event) *Buffer {
	return &e.Buffer
}
