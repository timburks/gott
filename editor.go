//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
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

// The Editor handles user commands and displays buffer text.
type Editor struct {
	Mode        int
	ScreenRows  int
	ScreenCols  int
	EditRows    int // actual number of rows used for text editing
	EditCols    int
	CursorRow   int
	CursorCol   int
	Message     string // status message
	RowOffset   int
	ColOffset   int
	Command     string
	CommandKeys string
	SearchText  string
	Debug       bool
	PasteBoard  string
	Multiplier  string
	Buffer      *Buffer
}

func NewEditor() *Editor {
	e := &Editor{}
	e.Buffer = NewBuffer()
	e.Mode = ModeEdit
	return e
}

func (e *Editor) ReadFile(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	e.Buffer.ReadBytes(b)
	e.Buffer.FileName = path
	return nil
}

func (e *Editor) Bytes() []byte {
	return e.Buffer.Bytes()
}

func (e *Editor) WriteFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	b := e.Bytes()
	out, err := gofmt(e.Buffer.FileName, b)
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
			if e.CursorRow > len(e.Buffer.Rows)-1 {
				e.CursorRow = len(e.Buffer.Rows) - 1
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
				filename = e.Buffer.FileName
			}
			e.WriteFile(filename)
		case "wq":
			var filename string
			if len(parts) == 2 {
				filename = parts[1]
			} else {
				filename = e.Buffer.FileName
			}
			e.WriteFile(filename)
			e.Mode = ModeQuit
			return
		case "fmt":
			out, err := gofmt(e.Buffer.FileName, e.Bytes())
			if err == nil {
				e.Buffer.ReadBytes(out)
			}
		case "$":
			e.CursorRow = len(e.Buffer.Rows) - 1
			if e.CursorRow < 0 {
				e.CursorRow = 0
			}
		default:
			e.Message = "nope"
		}
	}
	e.Command = ""
	e.Mode = ModeEdit
}

func (e *Editor) PerformSearch() {
	if len(e.Buffer.Rows) == 0 {
		return
	}
	row := e.CursorRow
	col := e.CursorCol + 1

	for {
		var s string
		if col < e.Buffer.Rows[row].Length() {
			s = e.Buffer.Rows[row].Text[col:]
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
			if row == len(e.Buffer.Rows) {
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

func (e *Editor) ProcessKeyEditMode(event termbox.Event) error {
	if e.CommandKeys == "d" {
		ch := event.Ch
		if ch != 0 {
			switch ch {
			case 'd':
				e.DeleteRow()
				e.KeepCursorInRow()
			case 'w':
				e.DeleteWord()
				e.KeepCursorInRow()
			}
		}
		e.CommandKeys = ""
		return nil
	}
	if e.CommandKeys == "r" {
		ch := event.Ch
		if ch != 0 {
			e.Buffer.Rows[e.CursorRow].ReplaceChar(e.CursorCol, ch)
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
			if e.CursorRow < len(e.Buffer.Rows) {
				e.CursorCol = e.Buffer.Rows[e.CursorRow].Length() - 1
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
		//
		// command multipliers are saved when operations are created
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			e.Multiplier += string(ch)
		//
		// commands go to the message bar
		case ':':
			e.Mode = ModeCommand
			e.Command = ""
		//
		// search queries go to the message bar
		case '/':
			e.Mode = ModeSearch
			e.SearchText = ""
		//
		// cursor movement isn't logged
		case 'h':
			e.MoveCursor(termbox.KeyArrowLeft)
		case 'j':
			e.MoveCursor(termbox.KeyArrowDown)
		case 'k':
			e.MoveCursor(termbox.KeyArrowUp)
		case 'l':
			e.MoveCursor(termbox.KeyArrowRight)
		//
		// operations are saved for undo and repetition
		case 'i':
			// insert at current location
			e.Mode = ModeInsert
		case 'a':
			// insert one character past the current location
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
	return nil
}

func (e *Editor) ProcessKeyInsertMode(event termbox.Event) error {
	key := event.Key
	if key != 0 {
		switch key {
		case termbox.KeyEsc:
			// ESC ends an operation.
			// We need to tell the operation that it finished
			// and what was inserted. It should return its own undo.
			// .. todo ..
			e.Mode = ModeEdit
			e.KeepCursorInRow()
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
	return nil
}

func (e *Editor) ProcessKeyCommandMode(event termbox.Event) error {
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
	return nil
}

func (e *Editor) ProcessKeySearchMode(event termbox.Event) error {
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
	return nil
}

func (e *Editor) ProcessKey(event termbox.Event) error {
	var err error
	switch e.Mode {
	case ModeEdit:
		err = e.ProcessKeyEditMode(event)
	case ModeInsert:
		err = e.ProcessKeyInsertMode(event)
	case ModeCommand:
		err = e.ProcessKeyCommandMode(event)
	case ModeSearch:
		err = e.ProcessKeySearchMode(event)
	}
	return err
}

func (e *Editor) Render() {
	termbox.Clear(termbox.ColorWhite, termbox.ColorBlack)
	w, h := termbox.Size()
	e.ScreenRows = h
	e.ScreenCols = w
	e.EditRows = e.ScreenRows - 2
	e.EditCols = e.ScreenCols

	e.Scroll()
	e.RenderInfoBar()
	e.RenderMessageBar()
	e.Buffer.X = 0
	e.Buffer.Y = 0
	e.Buffer.W = e.ScreenCols
	e.Buffer.H = e.ScreenRows - 2
	e.Buffer.YOffset = e.RowOffset
	e.Buffer.Render()
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
	finalText := fmt.Sprintf(" %d/%d ", e.CursorRow, len(e.Buffer.Rows))
	text := " the gott editor - " + e.Buffer.FileName + " "
	for len(text) < e.ScreenCols-len(finalText)-1 {
		text = text + " "
	}
	text += finalText
	for x, c := range text {
		termbox.SetCell(x, e.ScreenRows-2,
			rune(c),
			termbox.ColorBlack, termbox.ColorWhite)
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

func (e *Editor) MoveCursor(key termbox.Key) {
	switch key {
	case termbox.KeyArrowLeft:
		if e.CursorCol > 0 {
			e.CursorCol--
		}
	case termbox.KeyArrowRight:
		if e.CursorRow < len(e.Buffer.Rows) {
			rowLength := e.Buffer.Rows[e.CursorRow].Length()
			if e.CursorCol < rowLength-1 {
				e.CursorCol++
			}
		}
	case termbox.KeyArrowUp:
		if e.CursorRow > 0 {
			e.CursorRow--
		}
	case termbox.KeyArrowDown:
		if e.CursorRow < len(e.Buffer.Rows)-1 {
			e.CursorRow++
		}
	}
	// don't go past the end of the current line
	if e.CursorRow < len(e.Buffer.Rows) {
		rowLength := e.Buffer.Rows[e.CursorRow].Length()
		if e.CursorCol > rowLength-1 {
			e.CursorCol = rowLength - 1
			if e.CursorCol < 0 {
				e.CursorCol = 0
			}
		}
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
