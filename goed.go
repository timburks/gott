package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"time"

	"./terminal"
)

const VERSION = "0.0.1"

const (
	CTRL_Q      = 0x11
	ARROW_LEFT  = 1000
	ARROW_RIGHT = 1001
	ARROW_UP    = 1002
	ARROW_DOWN  = 1003
	PAGE_UP     = 1004
	PAGE_DOWN   = 1005
	DELETE      = 1006
)

type Editor struct {
	Reader *bufio.Reader

	ScreenRows int
	ScreenCols int

	CursorRow int
	CursorCol int

	Message string
}

func NewEditor() *Editor {
	editor := &Editor{}
	editor.Reader = bufio.NewReader(os.Stdin)
	w, h, err := terminal.GetSize(0)
	if err != nil {
		fmt.Print(err)
	}
	editor.ScreenRows = h
	editor.ScreenCols = w
	return editor
}

func (e *Editor) ReadKey() int {
	var err error
	b := make([]byte, 10)
	n := 0
	for n == 0 {
		n, _ = os.Stdin.Read(b)
		if n == 0 {
			time.Sleep(time.Microsecond)
		}
	}
	e.Message = fmt.Sprintf(" code=%02x", b[0:n])
	if err != nil {
		fmt.Print(err)
	}
	switch b[0] {
	case 0x1b:
		switch b[1] {
		case 0x5b:
			switch b[2] {
			case 'A':
				return ARROW_UP
			case 'B':
				return ARROW_DOWN
			case 'C':
				return ARROW_RIGHT
			case 'D':
				return ARROW_LEFT
			case 0x33:
				switch b[3] {
				case 0x7e:
					return DELETE
				}
			}
		}
	}
	return int(b[0])
}

func (e *Editor) ProcessKeyPress() error {
	key := e.ReadKey()
	e.Message += fmt.Sprintf(" key=%d", key)

	switch key {
	case CTRL_Q:
		e.Exit()
		return errors.New("quit")
	case ARROW_UP, ARROW_DOWN, ARROW_LEFT, ARROW_RIGHT:
		e.MoveCursor(key)
	}
	return nil
}

func (e *Editor) RefreshScreen() {
	buffer := make([]byte, 0)
	buffer = append(buffer, []byte("\x1b[?25l")...) // hide cursor
	buffer = append(buffer, []byte("\x1b[1;1H")...) // move cursor to row 1, col 1
	buffer = e.DrawRows(buffer)

	buffer = append(buffer, []byte(fmt.Sprintf("\x1b[%d;%dH", e.CursorRow+1, e.CursorCol+1))...)

	buffer = append(buffer, []byte("\x1b[?25h")...) // show cursor
	os.Stdout.Write(buffer)
}

func (e *Editor) Exit() {
	buffer := make([]byte, 0)
	buffer = append(buffer, []byte("\x1b[2J")...)   // clear screen
	buffer = append(buffer, []byte("\x1b[1;1H")...) // move cursor to row 1, col 1
	os.Stdout.Write(buffer)
}

func (e *Editor) DrawRows(buffer []byte) []byte {
	for y := 1; y <= e.ScreenRows; y++ {
		if y == e.ScreenRows/3 {
			welcome := fmt.Sprintf("goed editor -- version %s", VERSION)
			padding := (e.ScreenCols - len(welcome)) / 2
			buffer = append(buffer, []byte("~")...)
			for i := 1; i <= padding; i++ {
				buffer = append(buffer, []byte(" ")...)
			}
			buffer = append(buffer, []byte(welcome)...)
		} else {
			buffer = append(buffer, []byte("~")...)
		}
		buffer = append(buffer, []byte("\x1b[K")...)
		if y < e.ScreenRows {
			buffer = append(buffer, []byte("\r\n")...)
		} else {
			buffer = append(buffer, []byte(e.Message)...)
		}
	}
	return buffer
}

func (e *Editor) MoveCursor(key int) {
	switch key {
	case ARROW_LEFT:
		if e.CursorCol > 0 {
			e.CursorCol--
		}
	case ARROW_RIGHT:
		if e.CursorCol < e.ScreenCols-1 {
			e.CursorCol++
		}
	case ARROW_UP:
		if e.CursorRow > 0 {
			e.CursorRow--
		}
	case ARROW_DOWN:
		if e.CursorRow < e.ScreenRows-1 {
			e.CursorRow++
		}
	}
}

func main() {

	// put the terminal into raw mode
	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	// restore terminal however we exit
	defer terminal.Restore(0, oldState)

	e := NewEditor()
	// input loop
	for {
		e.RefreshScreen()
		err = e.ProcessKeyPress()
		if err != nil {
			break
		}
	}
}
