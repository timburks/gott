package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/nsf/termbox-go"
)

const VERSION = "0.0.1"

// Editor modes
const (
	ModeEdit    = 0
	ModeInsert  = 1
	ModeCommand = 2
	ModeQuit    = 9999
)

// A row of text in the editor
type Row struct {
	Text string
}

func NewRow(text string) Row {
	r := Row{}
	r.Text = strings.Replace(text, "\t", "    ", -1)
	return r
}

func (r *Row) DisplayText() string {
	return r.Text
}

func (r *Row) Length() int {
	return len(r.DisplayText())
}

func (r *Row) InsertChar(position int, c rune) {
	line := ""
	if position <= len(r.Text) {
		line += r.Text[0:position]
	} else {
		line += r.Text
	}
	line += string(c)
	if position < len(r.Text) {
		line += r.Text[position:]
	}
	r.Text = line
}

// delete
func (r *Row) DeleteChar(position int) {
	if r.Length() == 0 {
		return
	}
	if position > r.Length()-1 {
		position = r.Length() - 1
	}
	r.Text = r.Text[0:position] + r.Text[position+1:]
}

// split
func (r *Row) Split(position int) Row {
	before := r.Text[0:position]
	after := r.Text[position:]
	r.Text = before
	return NewRow(after)
}

// The Editor
type Editor struct {
	Mode        int
	ScreenRows  int
	ScreenCols  int
	EditRows    int // actual number of rows used for text editing
	EditCols    int
	CursorRow   int
	CursorCol   int
	Message     string // status message
	Rows        []Row
	RowOffset   int
	ColOffset   int
	FileName    string
	Command     string
	CommandKeys string
}

func NewEditor() *Editor {
	e := &Editor{}
	e.Rows = make([]Row, 0)
	e.Mode = ModeEdit
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
		e.Rows = append(e.Rows, NewRow(line))
	}
	e.FileName = path
	return nil
}

func (e *Editor) WriteFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, row := range e.Rows {
		f.WriteString(row.Text + "\n")
	}
	return nil
}

func (e *Editor) PerformCommand() {

	parts := strings.Split(e.Command, " ")
	if len(parts) > 0 {

		i, err := strconv.ParseInt(parts[0], 10, 64)
		if err == nil {
			e.CursorRow = int(i - 1)
			if e.CursorRow > len(e.Rows)-1 {
				e.CursorRow = len(e.Rows) - 1
			}
			if e.CursorRow < 0 {
				e.CursorRow = 0
			}
		}
		switch parts[0] {
		case "q":
			e.Mode = ModeQuit
			return
		case "r":
			if len(parts) == 2 {
				filename := parts[1]
				e.ReadFile(filename)
			}
		case "w":
			var filename string
			if len(parts) == 2 {
				filename = parts[1]
			} else {
				filename = e.FileName
			}
			e.WriteFile(filename)
		case "$":
			e.CursorRow = len(e.Rows) - 1
			if e.CursorRow < 0 {
				e.CursorRow = 0
			}
		default:
			e.Message = "hey hey hey"
		}
	}
	e.Command = ""
	e.Mode = ModeEdit
}

func (e *Editor) ProcessNextEvent() error {
	event := termbox.PollEvent()

	//e.Message = fmt.Sprintf("event=%+v", event)

	switch event.Type {
	case termbox.EventResize:
		termbox.Flush()
		return nil
	case termbox.EventKey:
		break // handle these below
	default:
		return nil
	}

	switch e.Mode {

	case ModeEdit:

		if e.CommandKeys == "d" {

			ch := event.Ch
			if ch != 0 {
				switch ch {
				case 'd':
					e.DeleteRow()
				}
			}
			e.CommandKeys = ""
			return nil
		}

		key := event.Key
		if key != 0 {
			switch key {
			case termbox.KeyEsc:
				break
			case termbox.KeyPgup:
				e.CursorRow = e.RowOffset
				for i := 0; i < e.EditRows; i++ {
					e.MoveCursor(termbox.KeyArrowUp)
				}
			case termbox.KeyPgdn:
				e.CursorRow = e.RowOffset + e.EditRows - 1
				for i := 0; i < e.EditRows; i++ {
					e.MoveCursor(termbox.KeyArrowDown)
				}
			case termbox.KeyCtrlA, termbox.KeyHome:
				e.CursorCol = 0
			case termbox.KeyCtrlE, termbox.KeyEnd:
				e.CursorCol = 0
				if e.CursorRow < len(e.Rows) {
					e.CursorCol = e.Rows[e.CursorRow].Length() - 1
					if e.CursorCol < 0 {
						e.CursorCol = 0
					}
				}
			case termbox.KeyArrowUp, termbox.KeyArrowDown, termbox.KeyArrowLeft, termbox.KeyArrowRight:
				e.MoveCursor(key)
			}
		}
		ch := event.Ch
		if ch != 0 {
			switch ch {
			case ':':
				e.Mode = ModeCommand
				e.Command = ""
			case 'h':
				e.MoveCursor(termbox.KeyArrowLeft)
			case 'j':
				e.MoveCursor(termbox.KeyArrowDown)
			case 'k':
				e.MoveCursor(termbox.KeyArrowUp)
			case 'l':
				e.MoveCursor(termbox.KeyArrowRight)
			case 'i':
				e.Mode = ModeInsert
			case 'x':
				e.DeleteChar()
			case 'd':
				e.CommandKeys = "d"
			}
		}

	case ModeInsert:
		key := event.Key
		if key != 0 {
			switch key {
			case termbox.KeyEsc:
				e.Mode = ModeEdit
			case termbox.KeyEnter:
				e.InsertRow()
				e.CursorRow++
				e.CursorCol = 0
			case termbox.KeySpace:
				e.InsertChar(' ')
			}
		}
		ch := event.Ch
		if ch != 0 {
			e.InsertChar(ch)
		}

	case ModeCommand:
		key := event.Key
		if key != 0 {
			switch key {
			case termbox.KeyEnter:
				e.PerformCommand()
			case termbox.KeyBackspace2:
				if len(e.Command) > 0 {
					e.Command = e.Command[0 : len(e.Command)-1]
				}
			case termbox.KeySpace:
				e.Command += " "
			}
		}
		ch := event.Ch
		if ch != 0 {
			e.Command = e.Command + string(ch)
		}
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

func (e *Editor) DrawScreen() {
	w, h := termbox.Size()
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

	termbox.SetCursor(e.CursorRow+1-e.RowOffset, e.CursorCol+1-e.ColOffset)

	buffer = append(buffer, []byte("\x1b[?25h")...) // show cursor
	os.Stdout.Write(buffer)
}

func (e *Editor) DrawRows(buffer []byte) []byte {
	for y := 0; y < e.ScreenRows; y++ {
		if y == e.ScreenRows-2 {
			// draw status bar
			buffer = append(buffer, []byte("\x1b[7m")...)
			finalText := fmt.Sprintf(" %d/%d ", e.CursorRow, len(e.Rows))
			text := " " + e.FileName + " "
			for len(text) < e.ScreenCols-len(finalText)-1 {
				text = text + " "
			}
			text += finalText
			buffer = append(buffer, []byte(text)...)
			buffer = append(buffer, []byte("\x1b[K")...)
			buffer = append(buffer, []byte("\x1b[m")...)
			buffer = append(buffer, []byte("\r\n")...)
		} else if y == e.ScreenRows-1 {

			if e.Mode == ModeCommand {
				buffer = append(buffer, []byte(":"+e.Command)...)
				buffer = append(buffer, []byte("\x1b[K")...)
			} else {
				// draw message bar
				text := e.Message
				if len(text) > e.ScreenCols {
					text = text[0:e.ScreenCols]
				}
				buffer = append(buffer, []byte(text)...)
				buffer = append(buffer, []byte("\x1b[K")...)
			}
		} else if (y + e.RowOffset) < len(e.Rows) {
			// draw editor text
			line := e.Rows[y+e.RowOffset].DisplayText()
			if e.ColOffset < len(line) {
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

func (e *Editor) MoveCursor(key termbox.Key) {
	switch key {
	case termbox.KeyArrowLeft:
		if e.CursorCol > 0 {
			e.CursorCol--
		}
	case termbox.KeyArrowRight:
		if e.CursorRow < len(e.Rows) {
			rowLength := e.Rows[e.CursorRow].Length()
			if e.CursorCol < rowLength-1 {
				e.CursorCol++
			}
		}
	case termbox.KeyArrowUp:
		if e.CursorRow > 0 {
			e.CursorRow--
		}
	case termbox.KeyArrowDown:
		if e.CursorRow < len(e.Rows)-1 {
			e.CursorRow++
		}
	}
	// don't go past the end of the current line
	if e.CursorRow < len(e.Rows) {
		rowLength := e.Rows[e.CursorRow].Length()
		if e.CursorCol > rowLength-1 {
			e.CursorCol = rowLength - 1
			if e.CursorCol < 0 {
				e.CursorCol = 0
			}
		}
	}
}

func (e *Editor) InsertChar(c rune) {
	if len(e.Rows) == 0 {
		e.Rows = append(e.Rows, NewRow(""))
	}
	e.Rows[e.CursorRow].InsertChar(e.CursorCol, c)
	e.CursorCol += 1
}

func (e *Editor) DeleteChar() {
	if len(e.Rows) > 0 {
		e.Rows[e.CursorRow].DeleteChar(e.CursorCol)
		if e.CursorCol > e.Rows[e.CursorRow].Length()-1 {
			e.CursorCol--
		}
		if e.CursorCol < 0 {
			e.CursorCol = 0
		}
	}
}

func (e *Editor) InsertRow() {
	if e.CursorRow > len(e.Rows)-1 {
		e.Rows = append(e.Rows, NewRow(""))
		e.CursorRow = len(e.Rows) - 1
	} else {
		position := e.CursorRow
		newRow := e.Rows[position].Split(e.CursorCol)
		e.Message = newRow.Text
		i := position + 1
		e.Rows = append(e.Rows, NewRow(""))
		copy(e.Rows[i+1:], e.Rows[i:])
		e.Rows[i] = newRow
	}
}

func (e *Editor) DeleteRow() {
	if len(e.Rows) == 0 {
		return
	}
	position := e.CursorRow
	e.Rows = append(e.Rows[0:position], e.Rows[position+1:]...)
	if position > len(e.Rows)-1 {
		position = len(e.Rows) - 1
	}
	if position < 0 {
		position = 0
	}
	e.CursorRow = position
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	e := NewEditor()

	if len(os.Args) > 1 {
		filename := os.Args[1]
		e.ReadFile(filename)
	}

	for e.Mode != ModeQuit {
		e.DrawScreen()
		e.ProcessNextEvent()
	}
}
