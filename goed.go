package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

const VERSION = "0.0.1"

type Editor struct {
	Reader *bufio.Reader

	ScreenRows int
	ScreenCols int
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

const CTRL_Q = 0x11

func (e *Editor) ReadKey() rune {
	rune, _, err := e.Reader.ReadRune()

	if err != nil {
		fmt.Print(err)
	}

	return rune
}

func (e *Editor) ProcessKeyPress() error {
	key := e.ReadKey()

	// print out the unicode value i.e. A -> 65, a -> 97
	fmt.Print(key)
	if key == CTRL_Q {
		e.Exit()
		return errors.New("quit")
	}
	return nil
}

func (e *Editor) RefreshScreen() {
	buffer := make([]byte, 0)
	buffer = append(buffer, []byte("\x1b[?25l")...) // hide cursor
	buffer = append(buffer, []byte("\x1b[1;1H")...) // move cursor to row 1, col 1
	buffer = e.DrawRows(buffer)
	buffer = append(buffer, []byte("\x1b[1;1H")...) // move cursor to row 1, col 1
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
		}
	}
	return buffer
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
	e.RefreshScreen()
	// input loop
	for {
		err = e.ProcessKeyPress()
		if err != nil {
			break
		}
	}
}
