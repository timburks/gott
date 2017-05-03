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

// Move directions
const (
	MoveUp    = 0
	MoveDown  = 1
	MoveRight = 2
	MoveLeft  = 3
)

// Insert positions
const (
	InsertAtCursor             = 0
	InsertAfterCursor          = 1
	InsertAtStartOfLine        = 2
	InsertAfterEndOfLine       = 3
	InsertAtNewLineBelowCursor = 4
	InsertAtNewLineAboveCursor = 5
)

// The Editor handles user commands and displays buffer text.
type Editor struct {
	Mode       int              // editor mode
	ScreenRows int              // screen size in rows
	ScreenCols int              // screen size in columns
	EditRows   int              // actual number of rows used for editing
	EditCols   int              // actual number of cols used for editing
	CursorRow  int              // cursor position
	CursorCol  int              // cursor position
	Message    string           // status message
	RowOffset  int              // display offset
	ColOffset  int              // display offset
	Command    string           // command as it is being typed on the command line
	EditKeys   string           // edit key sequences in progress
	Multiplier string           // multiplier string as it is being entered
	SearchText string           // text for searches as it is being typed
	Debug      bool             // debug mode displays information about events (key codes, etc)
	PasteBoard string           // used to cut/copy and paste
	Buffer     *Buffer          // active buffer being edited
	Repeat     Operation        // last operation performed, available to repeat
	Undo       []Operation      // stack of operations to undo
	Insert     *InsertOperation // when in insert mode, the current insert operation
}

func NewEditor() *Editor {
	e := &Editor{}
	e.Buffer = NewBuffer()
	e.Mode = ModeEdit
	return e
}

func (e *Editor) SetCursor(row, col int) {
	e.CursorRow = row
	e.CursorCol = col
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

func (e *Editor) Perform(op Operation) {
	// perform the operation
	inverse := op.Perform(e)
	// save the operation for repeats
	e.Repeat = op
	// save the inverse of the operation for undo
	if inverse != nil {
		e.Undo = append(e.Undo, inverse)
	}
}

func (e *Editor) ProcessKeyEditMode(event termbox.Event) error {
	// multikey commands have highest precedence
	if len(e.EditKeys) > 0 {
		ch := event.Ch
		switch e.EditKeys {
		case "d":
			switch ch {
			case 'd':
				e.Perform(&DeleteRowOperation{})
				e.KeepCursorInRow()
			case 'w':
				e.Perform(&DeleteWordOperation{})
				e.KeepCursorInRow()
			}
		case "r":
			if ch != 0 {
				e.Perform(&ReplaceCharacterOperation{Character: rune(event.Ch)})
			}
		case "y":
			switch ch {
			case 'y': // YankRow
				e.YankRow()
			default:
				break
			}
		}
		e.EditKeys = ""
		return nil
	}
	key := event.Key
	if key != 0 {
		switch key {
		case termbox.KeyEsc:
			break
		case termbox.KeyPgup:
			// move to the top of the screen
			e.CursorRow = e.RowOffset
			// move up by a page
			for i := 0; i < e.EditRows; i++ {
				e.MoveCursor(MoveUp)
			}
		case termbox.KeyPgdn:
			// move to the bottom of the screen
			e.CursorRow = e.RowOffset + e.EditRows - 1
			// move down by a page
			for i := 0; i < e.EditRows; i++ {
				e.MoveCursor(MoveDown)
			}
		case termbox.KeyCtrlA, termbox.KeyHome:
			// move to beginning of line
			e.CursorCol = 0
		case termbox.KeyCtrlE, termbox.KeyEnd:
			// move to end of line
			e.CursorCol = 0
			if e.CursorRow < len(e.Buffer.Rows) {
				e.CursorCol = e.Buffer.Rows[e.CursorRow].Length() - 1
				if e.CursorCol < 0 {
					e.CursorCol = 0
				}
			}
		case termbox.KeyArrowUp:
			e.MoveCursor(MoveUp)
		case termbox.KeyArrowDown:
			e.MoveCursor(MoveDown)
		case termbox.KeyArrowLeft:
			e.MoveCursor(MoveLeft)
		case termbox.KeyArrowRight:
			e.MoveCursor(MoveRight)
		}
	}
	ch := event.Ch
	if ch != 0 {
		switch ch {
		//
		// command multipliers are saved when operations are created
		//
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			e.Multiplier += string(ch)
		//
		// commands go to the message bar
		//
		case ':':
			e.Mode = ModeCommand
			e.Command = ""
		//
		// search queries go to the message bar
		//
		case '/':
			e.Mode = ModeSearch
			e.SearchText = ""
		case 'n': // repeat the last search
			e.PerformSearch()
		//
		// cursor movement isn't logged
		//
		case 'h':
			e.MoveCursor(MoveLeft)
		case 'j':
			e.MoveCursor(MoveDown)
		case 'k':
			e.MoveCursor(MoveUp)
		case 'l':
			e.MoveCursor(MoveRight)
		//
		// "performed" operations are saved for undo and repetition
		//
		case 'i':
			e.Perform(&InsertOperation{Position: InsertAtCursor})
		case 'a':
			e.Perform(&InsertOperation{Position: InsertAfterCursor})
		case 'I':
			e.Perform(&InsertOperation{Position: InsertAtStartOfLine})
		case 'A':
			e.Perform(&InsertOperation{Position: InsertAfterEndOfLine})
		case 'o':
			e.Perform(&InsertOperation{Position: InsertAtNewLineBelowCursor})
		case 'O':
			e.Perform(&InsertOperation{Position: InsertAtNewLineAboveCursor})
		case 'x':
			e.Perform(&DeleteCharacterOperation{})
		case 'p': // PasteText
			e.Perform(&PasteOperation{})
		//
		// a few keys open multi-key commands
		//
		case 'd':
			e.EditKeys = "d"
		case 'y':
			e.EditKeys = "y"
		case 'r':
			e.EditKeys = "r"
		//
		// undo
		//
		case 'u':
			e.PerformUndo()
		}
	}
	return nil
}

func (e *Editor) PerformUndo() {
	if len(e.Undo) > 0 {
		last := len(e.Undo) - 1
		undo := e.Undo[last]
		e.Undo = e.Undo[0:last]
		undo.Perform(e)
	}
}

func (e *Editor) ProcessKeyInsertMode(event termbox.Event) error {
	key := event.Key
	if key != 0 {
		switch key {
		case termbox.KeyEsc: // end an insert operation.
			e.Insert.Close()
			e.Insert = nil
			e.Mode = ModeEdit
			e.KeepCursorInRow()
		case termbox.KeyBackspace2:
			e.BackspaceChar()
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

func (e *Editor) MoveCursor(direction int) {
	switch direction {
	case MoveLeft:
		if e.CursorCol > 0 {
			e.CursorCol--
		}
	case MoveRight:
		if e.CursorRow < len(e.Buffer.Rows) {
			rowLength := e.Buffer.Rows[e.CursorRow].Length()
			if e.CursorCol < rowLength-1 {
				e.CursorCol++
			}
		}
	case MoveUp:
		if e.CursorRow > 0 {
			e.CursorRow--
		}
	case MoveDown:
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

// These editor primitives will make changes in insert mode and associate them with to the current operation.

func (e *Editor) InsertChar(c rune) {
	if e.Insert != nil {
		e.Insert.Text += string(c)
	}
	if c == '\n' {
		e.InsertRow()
		e.CursorRow++
		e.CursorCol = 0
		return
	}
	// if the cursor is past the nmber of rows, add a row
	for e.CursorRow >= len(e.Buffer.Rows) {
		e.AppendBlankRow()
	}
	e.Buffer.Rows[e.CursorRow].InsertChar(e.CursorCol, c)
	e.CursorCol += 1
}

func (e *Editor) InsertRow() {
	if e.CursorRow >= len(e.Buffer.Rows) {
		// we should never get here
		e.AppendBlankRow()
	} else {
		newRow := e.Buffer.Rows[e.CursorRow].Split(e.CursorCol)
		i := e.CursorRow + 1
		// add a dummy row at the end of the Rows slice
		e.AppendBlankRow()
		// move rows to make room for the one we are adding
		copy(e.Buffer.Rows[i+1:], e.Buffer.Rows[i:])
		// add the new row
		e.Buffer.Rows[i] = newRow
	}
}

func (e *Editor) BackspaceChar() rune {
	if len(e.Buffer.Rows) == 0 {
		return rune(0)
	}
	if len(e.Insert.Text) == 0 {
		return rune(0)
	}
	e.Insert.Text = e.Insert.Text[0 : len(e.Insert.Text)-1]
	if e.CursorCol > 0 {
		c := e.Buffer.Rows[e.CursorRow].DeleteChar(e.CursorCol - 1)
		e.CursorCol--
		return c
	} else if e.CursorRow > 0 {
		// remove the current row and join it with the previous one
		oldRowText := e.Buffer.Rows[e.CursorRow].Text
		newCursorCol := len(e.Buffer.Rows[e.CursorRow-1].Text)
		e.Buffer.Rows[e.CursorRow-1].Text += oldRowText
		e.Buffer.Rows = append(e.Buffer.Rows[0:e.CursorRow], e.Buffer.Rows[e.CursorRow+1:]...)
		e.CursorRow--
		e.CursorCol = newCursorCol
		return rune('\n')
	} else {
		return rune(0)
	}
}

func (e *Editor) YankRow() {
	if len(e.Buffer.Rows) == 0 {
		return
	}
	e.PasteBoard = ""
	N := e.MultiplierValue()
	for i := 0; i < N; i++ {
		if i > 0 {
			e.PasteBoard += "\n"
		}
		position := e.CursorRow + i
		if position < len(e.Buffer.Rows) {
			e.PasteBoard += e.Buffer.Rows[position].Text
		}
	}
}

func (e *Editor) KeepCursorInRow() {
	if len(e.Buffer.Rows) == 0 {
		e.CursorCol = 0
	} else {
		if e.CursorRow >= len(e.Buffer.Rows) {
			e.CursorRow = len(e.Buffer.Rows) - 1
		}
		if e.CursorRow < 0 {
			e.CursorRow = 0
		}
		lastIndexInRow := e.Buffer.Rows[e.CursorRow].Length() - 1
		if e.CursorCol > lastIndexInRow {
			e.CursorCol = lastIndexInRow
		}
		if e.CursorCol < 0 {
			e.CursorCol = 0
		}
	}
}

func (e *Editor) AppendBlankRow() {
	e.Buffer.Rows = append(e.Buffer.Rows, NewRow(""))
}

func (e *Editor) InsertLineAboveCursor() {
	e.AppendBlankRow()
	copy(e.Buffer.Rows[e.CursorRow+1:], e.Buffer.Rows[e.CursorRow:])
	e.Buffer.Rows[e.CursorRow] = NewRow("")
	e.CursorCol = 0
}

func (e *Editor) InsertLineBelowCursor() {
	e.AppendBlankRow()
	copy(e.Buffer.Rows[e.CursorRow+2:], e.Buffer.Rows[e.CursorRow+1:])
	e.Buffer.Rows[e.CursorRow+1] = NewRow("")
	e.CursorRow += 1
	e.CursorCol = 0
}
