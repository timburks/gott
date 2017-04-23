package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

type Editor struct {
	Reader *bufio.Reader
}

func NewEditor() *Editor {
	editor := &Editor{}
	editor.Reader = bufio.NewReader(os.Stdin)
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
	os.Stdout.Write([]byte("\x1b[2J"))   // clear screen
	os.Stdout.Write([]byte("\x1b[1;1H")) // move cursor to row 1, col 1
}

func (e *Editor) Exit() {
	os.Stdout.Write([]byte("\x1b[2J"))   // clear screen
	os.Stdout.Write([]byte("\x1b[1;1H")) // move cursor to row 1, col 1

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
