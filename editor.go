package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/nsf/termbox-go"
)

// Editor modes
const (
	ModeEdit    = 0
	ModeInsert  = 1
	ModeCommand = 2
	ModeSearch  = 3
	ModeQuit    = 9999
)

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
	PasteBoard  string
	Multiplier  string
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
	e.ReadBytes(b)
	e.FileName = path
	return nil
}

func (e *Editor) ReadBytes(b []byte) {
	s := string(b)
	lines := strings.Split(s, "\n")
	e.Rows = make([]Row, 0)
	for _, line := range lines {
		e.Rows = append(e.Rows, NewRow(line))
	}
}

func (e *Editor) WriteFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	b := e.Bytes()
	out, err := gofmt(e.FileName, b)
	if err == nil {
		f.Write(out)
	} else {
		f.Write(b)
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
		case "fmt":
			out, err := gofmt(e.FileName, e.Bytes())
			if err == nil {
				e.ReadBytes(out)
			}
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
		var s string
		if col < e.Rows[row].Length() {
			s = e.Rows[row].Text[col:]
		} else {
			s = ""
		}
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

func (e *Editor) ProcessEvent(event termbox.Event) error {
	if e.Debug {
		e.Message = fmt.Sprintf("event=%+v", event)
	}
	switch event.Type {
	case termbox.EventResize:
		return e.ProcessResize(event)
	case termbox.EventKey:
		return e.ProcessKey(event)
	default:
		return nil
	}
}

func (e *Editor) ProcessResize(event termbox.Event) error {
	termbox.Flush()
	return nil
}

func (e *Editor) ProcessKey(event termbox.Event) error {
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
		if e.CommandKeys == "r" {
			ch := event.Ch
			if ch != 0 {
				e.Rows[e.CursorRow].ReplaceChar(e.CursorCol, ch)
			}
			e.CommandKeys = ""
			return nil
		}
		if e.CommandKeys == "y" {
			ch := event.Ch
			switch ch {
			case 'y':
				e.YankRow()
			default:
				break
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
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				e.Multiplier += string(ch)
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
			case 'y':
				e.CommandKeys = "y"
			case 'p':
				e.Paste()
			case 'n':
				e.PerformSearch()
			case 'r':
				e.CommandKeys = "r"
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
			case termbox.KeyTab:
				e.InsertChar(' ')
				for {
					if e.CursorCol%8 == 0 {
						break
					}
					e.InsertChar(' ')
				}
			case termbox.KeyEsc:
				e.Mode = ModeEdit
				e.KeepCursorInRow()
			case termbox.KeyEnter:
				e.InsertChar('\n')
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

func (e *Editor) DrawScreen() {
	termbox.Clear(termbox.ColorBlack, termbox.ColorWhite)
	w, h := termbox.Size()
	e.ScreenRows = h
	e.ScreenCols = w
	e.EditRows = e.ScreenRows - 2
	e.EditCols = e.ScreenCols

	e.Scroll()
	e.RenderInfoBar()
	e.RenderMessageBar()
	e.RenderTextArea()
	termbox.SetCursor(e.CursorCol-e.ColOffset, e.CursorRow-e.RowOffset)
	termbox.Flush()
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

func (e *Editor) RenderInfoBar() {
	finalText := fmt.Sprintf(" %d/%d ", e.CursorRow, len(e.Rows))
	text := " the gott editor - " + e.FileName + " "
	for len(text) < e.ScreenCols-len(finalText)-1 {
		text = text + " "
	}
	text += finalText
	for x, c := range text {
		termbox.SetCell(x, e.ScreenRows-2, rune(c), termbox.ColorWhite, termbox.ColorBlack)
	}
}

func (e *Editor) RenderMessageBar() {
	var line string
	if e.Mode == ModeCommand {
		line += ":" + e.Command
	} else if e.Mode == ModeSearch {
		line += "/" + e.SearchText
	} else {
		line += e.Message
	}
	if len(line) > e.ScreenCols {
		line = line[0:e.ScreenCols]
	}
	for x, c := range line {
		termbox.SetCell(x, e.ScreenRows-1, rune(c), termbox.ColorBlack, termbox.ColorWhite)
	}
}

func (e *Editor) RenderTextArea() {
	for y := 0; y < e.ScreenRows-2; y++ {
		var line string
		if (y + e.RowOffset) < len(e.Rows) {
			line = e.Rows[y+e.RowOffset].DisplayText()
			if e.ColOffset < len(line) {
				line = line[e.ColOffset:]
			} else {
				line = ""
			}
		} else {
			line = "~"
			if y == e.ScreenRows/3 {
				welcome := fmt.Sprintf("the gott editor -- version %s", VERSION)
				padding := (e.ScreenCols - len(welcome)) / 2
				for i := 1; i <= padding; i++ {
					line = line + " "
				}
				line += welcome
			}
		}
		if len(line) > e.ScreenCols {
			line = line[0:e.ScreenCols]
		}
		for x, c := range line {
			termbox.SetCell(x, y, rune(c), termbox.ColorBlack, termbox.ColorWhite)
		}
	}
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
	if c == '\n' {
		e.InsertRow()
		e.CursorRow++
		e.CursorCol = 0
		return
	}
	if len(e.Rows) == 0 {
		e.Rows = append(e.Rows, NewRow(""))
	}
	for e.CursorRow >= len(e.Rows) {
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

func (e *Editor) MultiplierValue() int {
	if e.Multiplier == "" {
		return 1
	}
	i, err := strconv.ParseInt(e.Multiplier, 10, 64)
	if err != nil {
		e.Multiplier = ""
		return 1
	}
	e.Multiplier = ""
	return int(i)
}

func (e *Editor) DeleteRow() {
	if len(e.Rows) == 0 {
		return
	}
	e.PasteBoard = ""
	N := e.MultiplierValue()
	for i := 0; i < N; i++ {
		if i > 0 {
			e.PasteBoard += "\n"
		}
		position := e.CursorRow
		e.PasteBoard += e.Rows[position].Text
		e.Rows = append(e.Rows[0:position], e.Rows[position+1:]...)
		if position > len(e.Rows)-1 {
			position = len(e.Rows) - 1
		}
		if position < 0 {
			position = 0
		}
		e.CursorRow = position
	}
}

func (e *Editor) YankRow() {
	if len(e.Rows) == 0 {
		return
	}
	e.PasteBoard = ""
	N := e.MultiplierValue()
	for i := 0; i < N; i++ {
		if i > 0 {
			e.PasteBoard += "\n"
		}
		position := e.CursorRow + i
		if position < len(e.Rows) {
			e.PasteBoard += e.Rows[position].Text
		}
	}
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

func (e *Editor) ReplaceCharacter() {
}

func (e *Editor) Bytes() []byte {
	var s string
	for _, row := range e.Rows {
		s += row.Text + "\n"
	}
	return []byte(s)
}

func (e *Editor) Paste() {
	e.InsertLineBelowCursor()
	for _, c := range e.PasteBoard {
		e.InsertChar(c)
	}
}
