package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"xmachine.net/go/goed/terminal"
)

const VERSION = "0.0.1"

const (
	CTRL_A      = 0x01
	CTRL_E      = 0x05
	CTRL_Q      = 0x11
	ARROW_LEFT  = 1000
	ARROW_RIGHT = 1001
	ARROW_UP    = 1002
	ARROW_DOWN  = 1003
	PAGE_UP     = 1004
	PAGE_DOWN   = 1005
	HOME        = 1006
	END         = 1007
	DELETE      = 1008
)

type Row struct {
	Text string
}

type Editor struct {
	Reader     *bufio.Reader
	ScreenRows int
	ScreenCols int
	EditRows   int // actual number of rows used for text editing
	EditCols   int
	CursorRow  int
	CursorCol  int
	Message    string // status message
	Rows       []Row
	RowOffset  int
	ColOffset  int
}

func NewEditor() *Editor {
	e := &Editor{}
	e.Reader = bufio.NewReader(os.Stdin)
	e.Rows = make([]Row, 0)
	return e
}

func (e *Editor) ReadFile(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	s := string(b)
	lines := strings.Split(s, "\n")
	e.Rows = make([]Row, 0)
	for _, line := range lines {
		e.Rows = append(e.Rows, Row{Text: line})
	}
	return nil
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
			case 0x31:
				switch b[3] {
				case 0x7e:
					return HOME
				}
			case 0x33:
				switch b[3] {
				case 0x7e:
					return DELETE
				}
			case 0x34:
				switch b[3] {
				case 0x7e:
					return END
				}
			case 0x35:
				switch b[3] {
				case 0x7e:
					return PAGE_UP
				}
			case 0x36:
				switch b[3] {
				case 0x7e:
					return PAGE_DOWN
				}
			}
		}
	}
	return int(b[0])
}

func (e *Editor) ProcessKeyPress() error {
	key := e.ReadKey()
	//e.Message += fmt.Sprintf(" key=%d", key)

	switch key {
	case CTRL_Q:
		e.Exit()
		return errors.New("quit")
	case PAGE_UP:
		for times := e.EditRows; times > 0; times-- {
			e.MoveCursor(ARROW_UP)
		}
	case PAGE_DOWN:
		for times := e.EditRows; times > 0; times-- {
			e.MoveCursor(ARROW_DOWN)
		}
	case CTRL_A:
		e.CursorCol = 0
	case CTRL_E:
		e.CursorCol = 0
		if e.CursorRow < len(e.Rows) {
			displayText := e.Rows[e.CursorRow].Text
			displayText = strings.Replace(displayText, "\t", "        ", -1)
			rowLength := len(displayText)
			e.CursorCol = rowLength - 1
		}
	case ARROW_UP, ARROW_DOWN, ARROW_LEFT, ARROW_RIGHT:
		e.MoveCursor(key)
	}
	return nil
}

func (e *Editor) Scroll() {
	if e.CursorRow < e.RowOffset {
		e.RowOffset = e.CursorRow
	}
	if e.CursorRow-e.RowOffset >= e.EditRows {
		e.RowOffset = e.CursorRow - e.EditRows + 1
	}
	if e.CursorCol < e.ColOffset {
		e.ColOffset = e.CursorCol
	}
	if e.CursorCol-e.ColOffset >= e.EditCols {
		e.ColOffset = e.CursorCol - e.EditCols + 1
	}
}

func (e *Editor) RefreshScreen() {
	w, h, err := terminal.GetSize(0)
	if err != nil {
		fmt.Print(err)
	}
	e.ScreenRows = h
	e.ScreenCols = w
	e.EditRows = e.ScreenRows - 2
	e.EditCols = e.ScreenCols

	e.Scroll()
	buffer := make([]byte, 0)
	buffer = append(buffer, []byte("\x1b[?25l")...) // hide cursor
	buffer = append(buffer, []byte("\x1b[1;1H")...) // move cursor to row 1, col 1
	buffer = e.DrawRows(buffer)

	buffer = append(buffer, []byte(fmt.Sprintf("\x1b[%d;%dH", e.CursorRow+1-e.RowOffset, e.CursorCol+1-e.ColOffset))...)

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
	for y := 0; y < e.ScreenRows; y++ {
		if y == e.ScreenRows-2 {
			buffer = append(buffer, []byte("\x1b[7m")...)
			text := fmt.Sprintf(" %d/%d ", e.CursorRow, len(e.Rows))
			for len(text) < e.ScreenCols-1 {
				text = " " + text
			}
			buffer = append(buffer, []byte(text)...)
			buffer = append(buffer, []byte("\x1b[K")...)
			buffer = append(buffer, []byte("\x1b[m")...)
			buffer = append(buffer, []byte("\r\n")...)
		} else if y == e.ScreenRows-1 {
			text := "The last word: " + e.Message
			buffer = append(buffer, []byte(text)...)
			buffer = append(buffer, []byte("\x1b[K")...)
		} else if (y + e.RowOffset) < len(e.Rows) {
			line := e.Rows[y+e.RowOffset].Text
			line = strings.Replace(line, "\t", "        ", -1)

			if len(line) > e.ColOffset {
				line = line[e.ColOffset:]
			} else {
				line = ""
			}
			if len(line) > e.ScreenCols {
				line = line[0:e.ScreenCols]
			}
			buffer = append(buffer, []byte(line)...)
			buffer = append(buffer, []byte("\x1b[K")...)
			buffer = append(buffer, []byte("\r\n")...)
		} else {
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
			buffer = append(buffer, []byte("\r\n")...)
		}
	}
	return buffer
}

func (e *Editor) MoveCursor(key int) {

	switch key {
	case ARROW_LEFT:
		if e.CursorCol > 0 {
			e.CursorCol--
		} else if e.CursorRow > 0 {
			e.CursorRow--
			displayText := e.Rows[e.CursorRow].Text
			displayText = strings.Replace(displayText, "\t", "        ", -1)
			rowLength := len(displayText)
			e.CursorCol = rowLength - 1
		}
	case ARROW_RIGHT:
		if e.CursorRow < len(e.Rows) {
			displayText := e.Rows[e.CursorRow].Text
			displayText = strings.Replace(displayText, "\t", "        ", -1)
			rowLength := len(displayText)
			if e.CursorCol < rowLength-1 {
				e.CursorCol++
			} else if e.CursorRow < len(e.Rows)-1 {
				e.CursorRow++
				e.CursorCol = 0
			}
		}
	case ARROW_UP:
		if e.CursorRow > 0 {
			e.CursorRow--
		}
	case ARROW_DOWN:
		if e.CursorRow < len(e.Rows)-1 {
			e.CursorRow++
		}
	}

	if e.CursorRow < len(e.Rows) {
		displayText := e.Rows[e.CursorRow].Text
		displayText = strings.Replace(displayText, "\t", "        ", -1)
		rowLength := len(displayText)
		if e.CursorCol > rowLength-1 {
			e.CursorCol = rowLength - 1
			if e.CursorCol < 0 {
				e.CursorCol = 0
			}
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
	e.ReadFile("goed.go")
	// input loop
	for {
		e.RefreshScreen()
		err = e.ProcessKeyPress()
		if err != nil {
			break
		}
	}
}
