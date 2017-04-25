package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/nsf/termbox-go"
)

const VERSION = "0.1.0"

// Editor modes
const (
	ModeEdit    = 0
	ModeInsert  = 1
	ModeCommand = 2
	ModeSearch  = 3
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
func (r *Row) DeleteChar(position int) rune {
	if r.Length() == 0 {
		return 0
	}
	if position > r.Length()-1 {
		position = r.Length() - 1
	}
	ch := rune(r.Text[position])
	r.Text = r.Text[0:position] + r.Text[position+1:]
	return ch
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
	SearchText  string
	Debug       bool
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
		case "debug":
			if len(parts) == 2 {
				if parts[1] == "on" {
					e.Debug = true
				} else if parts[1] == "off" {
					e.Debug = false
					e.Message = ""
				}
			}
		case "w":
			var filename string
			if len(parts) == 2 {
				filename = parts[1]
			} else {
				filename = e.FileName
			}
			e.WriteFile(filename)
		case "wq":
			var filename string
			if len(parts) == 2 {
				filename = parts[1]
			} else {
				filename = e.FileName
			}
			e.WriteFile(filename)
			e.Mode = ModeQuit
			return
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

func (e *Editor) PerformSearch() {
	if len(e.Rows) == 0 {
		return
	}
	row := e.CursorRow
	col := e.CursorCol + 1

	for {
		s := e.Rows[row].Text[col:]
		i := strings.Index(s, e.SearchText)
		if i != -1 {
			// found it
			e.CursorRow = row
			e.CursorCol = col + i
			return
		} else {
			col = 0
			row = row + 1
			if row == len(e.Rows) {
				row = 0
			}
		}
		if row == e.CursorRow {
			break
		}
	}
}

func (e *Editor) ProcessNextEvent() error {
	event := termbox.PollEvent()

	if e.Debug {
		e.Message = fmt.Sprintf("event=%+v", event)
	}
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
	//
	// EDIT MODE
	//
	case ModeEdit:

		if e.CommandKeys == "d" {

			ch := event.Ch
			if ch != 0 {
				switch ch {
				case 'd':
					e.DeleteRow()
				case 'w':
					e.DeleteWord()
				}
				e.KeepCursorInRow()
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
			case '/':
				e.Mode = ModeSearch
				e.SearchText = ""
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
			case 'a':
				e.CursorCol++
				e.Mode = ModeInsert
			case 'I':
				e.MoveCursorToStartOfLine()
				e.Mode = ModeInsert
			case 'A':
				e.MoveCursorPastEndOfLine()
				e.Mode = ModeInsert
			case 'o':
				e.InsertLineBelowCursor()
				e.Mode = ModeInsert
			case 'O':
				e.InsertLineAboveCursor()
				e.Mode = ModeInsert
			case 'x':
				e.DeleteCharacterUnderCursor()
			case 'd':
				e.CommandKeys = "d"
			case 'n':
				e.PerformSearch()
			}
		}
	//
	// INSERT MODE
	//
	case ModeInsert:
		key := event.Key
		if key != 0 {
			switch key {
			case termbox.KeyBackspace2:
				e.MoveCursor(termbox.KeyArrowLeft)
				e.DeleteCharacterUnderCursor()
			case termbox.KeyEsc:
				e.Mode = ModeEdit
				e.KeepCursorInRow()
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

	//
	// COMMAND MODE
	//
	case ModeCommand:
		key := event.Key
		if key != 0 {
			switch key {
			case termbox.KeyEsc:
				e.Mode = ModeEdit
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

	//
	// SEARCH MODE
	//
	case ModeSearch:
		key := event.Key
		if key != 0 {
			switch key {
			case termbox.KeyEsc:
				e.Mode = ModeEdit
			case termbox.KeyEnter:
				e.PerformSearch()
				e.Mode = ModeEdit
			case termbox.KeyBackspace2:
				if len(e.SearchText) > 0 {
					e.SearchText = e.SearchText[0 : len(e.SearchText)-1]
				}
			case termbox.KeySpace:
				e.SearchText += " "
			}
		}
		ch := event.Ch
		if ch != 0 {
			e.SearchText = e.SearchText + string(ch)
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
			// draw bottom bar: message or command line
			if e.Mode == ModeCommand {
				buffer = append(buffer, []byte(":"+e.Command)...)
				buffer = append(buffer, []byte("\x1b[K")...)
			} else if e.Mode == ModeSearch {
				buffer = append(buffer, []byte("/"+e.SearchText)...)
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

func (e *Editor) DeleteWord() {
	if len(e.Rows) == 0 {
		return
	}

	c := e.Rows[e.CursorRow].DeleteChar(e.CursorCol)
	for {
		if e.CursorCol > e.Rows[e.CursorRow].Length()-1 {
			break
		}
		if c == ' ' {
			break
		}
		c = e.Rows[e.CursorRow].DeleteChar(e.CursorCol)
	}
	if e.CursorCol > e.Rows[e.CursorRow].Length()-1 {
		e.CursorCol--
	}
	if e.CursorCol < 0 {
		e.CursorCol = 0
	}
}

func (e *Editor) DeleteCharacterUnderCursor() {
	if len(e.Rows) == 0 {
		return
	}
	e.Rows[e.CursorRow].DeleteChar(e.CursorCol)
	if e.CursorCol > e.Rows[e.CursorRow].Length()-1 {
		e.CursorCol--
	}
	if e.CursorCol < 0 {
		e.CursorCol = 0
	}
}

func (e *Editor) MoveCursorToStartOfLine() {
	e.CursorCol = 0
}

func (e *Editor) MoveCursorPastEndOfLine() {
	if len(e.Rows) == 0 {
		return
	}
	e.CursorCol = e.Rows[e.CursorRow].Length()
}

func (e *Editor) KeepCursorInRow() {
	if len(e.Rows) == 0 {
		e.CursorCol = 0
	} else {
		if e.CursorRow >= len(e.Rows) {
			e.CursorRow = len(e.Rows) - 1
		}
		if e.CursorRow < 0 {
			e.CursorRow = 0
		}
		lastIndexInRow := e.Rows[e.CursorRow].Length() - 1
		if e.CursorCol > lastIndexInRow {
			e.CursorCol = lastIndexInRow
		}
		if e.CursorCol < 0 {
			e.CursorCol = 0
		}
	}
}

func (e *Editor) InsertLineAboveCursor() {
	if len(e.Rows) == 0 {
		e.InsertChar(' ')
	}
	i := e.CursorRow
	e.Rows = append(e.Rows, NewRow(""))
	copy(e.Rows[i+1:], e.Rows[i:])
	e.Rows[i] = NewRow("")
	e.CursorRow = i
	e.CursorCol = 0
}

func (e *Editor) InsertLineBelowCursor() {
	if len(e.Rows) == 0 {
		e.InsertChar(' ')
	}
	i := e.CursorRow
	e.Rows = append(e.Rows, NewRow(""))
	copy(e.Rows[i+2:], e.Rows[i+1:])
	e.Rows[i+1] = NewRow("")
	e.CursorRow = i + 1
	e.CursorCol = 0
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
