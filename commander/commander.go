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
package commander

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nsf/termbox-go"
	"github.com/timburks/gott/operations"
	gott "github.com/timburks/gott/types"
)

// The Commander converts user input into commands for the Editor.
type Commander struct {
	editor     gott.Editor
	mode       int    // editor mode
	debug      bool   // debug mode displays information about events (key codes, etc)
	editKeys   string // edit key sequences in progress
	command    string // command as it is being typed on the command line
	searchText string // text for searches as it is being typed
	message    string // status message
	multiplier string // multiplier string as it is being entered
}

func NewCommander(e gott.Editor) *Commander {
	return &Commander{editor: e, mode: gott.ModeEdit}
}

func (c *Commander) GetMode() int {
	return c.mode
}

func (c *Commander) SetMode(m int) {
	c.mode = m
}

func (c *Commander) ProcessNextEvent() error {
	event := termbox.PollEvent()
	if c.debug {
		c.message = fmt.Sprintf("event=%+v", event)
	}
	switch event.Type {
	case termbox.EventResize:
		return c.ProcessResize(event)
	case termbox.EventKey:
		return c.ProcessKey(event)
	default:
		return nil
	}
}

func (c *Commander) ProcessResize(event termbox.Event) error {
	termbox.Flush()
	return nil
}

func (c *Commander) ProcessKeyEditMode(event termbox.Event) error {
	e := c.editor

	key := event.Key
	ch := event.Ch

	// multikey commands have highest precedence
	if len(c.editKeys) > 0 {
		switch c.editKeys {
		case "c":
			switch ch {
			case 'w':
				e.Perform(&operations.ChangeWord{Commander: c}, c.Multiplier())
			}
		case "d":
			switch ch {
			case 'd':
				e.Perform(&operations.DeleteRow{}, c.Multiplier())
			case 'w':
				e.Perform(&operations.DeleteWord{}, c.Multiplier())
			}
		case "r":
			if key != 0 {
				if key == termbox.KeySpace {
					e.Perform(&operations.ReplaceCharacter{Character: rune(' ')}, c.Multiplier())
				}
			} else if ch != 0 {
				e.Perform(&operations.ReplaceCharacter{Character: rune(event.Ch)}, c.Multiplier())
			}
		case "y":
			switch ch {
			case 'y': // YankRow
				e.YankRow(c.Multiplier())
			default:
				break
			}
		}
		c.editKeys = ""
		return nil
	}
	if key != 0 {
		switch key {
		case termbox.KeyEsc:
			break
		case termbox.KeyPgup:
			e.PageUp()
		case termbox.KeyPgdn:
			e.PageDown()
		case termbox.KeyCtrlA, termbox.KeyHome:
			e.MoveToBeginningOfLine()
		case termbox.KeyCtrlE, termbox.KeyEnd:
			e.MoveToEndOfLine()
		case termbox.KeyArrowUp:
			e.MoveCursor(gott.MoveUp)
		case termbox.KeyArrowDown:
			e.MoveCursor(gott.MoveDown)
		case termbox.KeyArrowLeft:
			e.MoveCursor(gott.MoveLeft)
		case termbox.KeyArrowRight:
			e.MoveCursor(gott.MoveRight)
		}
	}
	if ch != 0 {
		switch ch {
		//
		// command multipliers are saved when operations are created
		//
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			c.multiplier += string(ch)
		//
		// commands go to the message bar
		//
		case ':':
			c.mode = gott.ModeCommand
			c.command = ""
		//
		// search queries go to the message bar
		//
		case '/':
			c.mode = gott.ModeSearch
			c.searchText = ""
		case 'n': // repeat the last search
			e.PerformSearch(c.searchText)
		//
		// cursor movement isn't logged
		//
		case 'h':
			e.MoveCursor(gott.MoveLeft)
		case 'j':
			e.MoveCursor(gott.MoveDown)
		case 'k':
			e.MoveCursor(gott.MoveUp)
		case 'l':
			e.MoveCursor(gott.MoveRight)
		case 'w':
			e.MoveCursorToNextWord(c.Multiplier())
		case 'b':
			e.MoveCursorToPreviousWord(c.Multiplier())
		//
		// "performed" operations are saved for undo and repetition
		//
		case 'i':
			e.Perform(&operations.Insert{Position: gott.InsertAtCursor, Commander: c}, c.Multiplier())
		case 'a':
			e.Perform(&operations.Insert{Position: gott.InsertAfterCursor, Commander: c}, c.Multiplier())
		case 'I':
			e.Perform(&operations.Insert{Position: gott.InsertAtStartOfLine, Commander: c}, c.Multiplier())
		case 'A':
			e.Perform(&operations.Insert{Position: gott.InsertAfterEndOfLine, Commander: c}, c.Multiplier())
		case 'o':
			e.Perform(&operations.Insert{Position: gott.InsertAtNewLineBelowCursor, Commander: c}, c.Multiplier())
		case 'O':
			e.Perform(&operations.Insert{Position: gott.InsertAtNewLineAboveCursor, Commander: c}, c.Multiplier())
		case 'x':
			e.Perform(&operations.DeleteCharacter{}, c.Multiplier())
		case 'J':
			e.Perform(&operations.JoinLine{}, c.Multiplier())
		case 'p': // PasteText
			e.Perform(&operations.Paste{}, c.Multiplier())
		case '~': // reverse case
			e.Perform(&operations.ReverseCaseCharacter{}, c.Multiplier())
		//
		// a few keys open multi-key commands
		//
		case 'c':
			c.editKeys = "c"
		case 'd':
			c.editKeys = "d"
		case 'y':
			c.editKeys = "y"
		case 'r':
			c.editKeys = "r"
		//
		// undo
		//
		case 'u':
			e.PerformUndo()
		//
		// repeat
		//
		case '.':
			e.Repeat()
		}
	}
	return nil
}

func (c *Commander) ProcessKeyInsertMode(event termbox.Event) error {
	e := c.editor

	key := event.Key
	if key != 0 {
		switch key {
		case termbox.KeyEsc: // end an insert operation.
			e.CloseInsert()
			c.mode = gott.ModeEdit
			e.KeepCursorInRow()
		case termbox.KeyBackspace2:
			e.BackspaceChar()
		case termbox.KeyTab:
			e.InsertChar(' ')
			for {
				if e.GetCursor().Col%8 == 0 {
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

func (c *Commander) ProcessKeyCommandMode(event termbox.Event) error {
	key := event.Key
	if key != 0 {
		switch key {
		case termbox.KeyEsc:
			c.mode = gott.ModeEdit
		case termbox.KeyEnter:
			c.PerformCommand()
		case termbox.KeyBackspace2:
			if len(c.command) > 0 {
				c.command = c.command[0 : len(c.command)-1]
			}
		case termbox.KeySpace:
			c.command += " "
		}
	}
	ch := event.Ch
	if ch != 0 {
		c.command = c.command + string(ch)
	}
	return nil
}

func (c *Commander) ProcessKeySearchMode(event termbox.Event) error {
	e := c.editor

	key := event.Key
	if key != 0 {
		switch key {
		case termbox.KeyEsc:
			c.mode = gott.ModeEdit
		case termbox.KeyEnter:
			e.PerformSearch(c.searchText)
			c.mode = gott.ModeEdit
		case termbox.KeyBackspace2:
			if len(c.searchText) > 0 {
				c.searchText = c.searchText[0 : len(c.searchText)-1]
			}
		case termbox.KeySpace:
			c.searchText += " "
		}
	}
	ch := event.Ch
	if ch != 0 {
		c.searchText = c.searchText + string(ch)
	}
	return nil
}

func (c *Commander) ProcessKey(event termbox.Event) error {
	var err error
	switch c.mode {
	case gott.ModeEdit:
		err = c.ProcessKeyEditMode(event)
	case gott.ModeInsert:
		err = c.ProcessKeyInsertMode(event)
	case gott.ModeCommand:
		err = c.ProcessKeyCommandMode(event)
	case gott.ModeSearch:
		err = c.ProcessKeySearchMode(event)
	}
	return err
}

func (c *Commander) PerformCommand() {
	e := c.editor

	parts := strings.Split(c.command, " ")
	if len(parts) > 0 {

		i, err := strconv.ParseInt(parts[0], 10, 64)
		if err == nil {
			newRow := int(i - 1)
			if newRow > e.GetBuffer().GetRowCount()-1 {
				newRow = e.GetBuffer().GetRowCount() - 1
			}
			if newRow < 0 {
				newRow = 0
			}
			cursor := e.GetCursor()
			cursor.Row = newRow
			cursor.Col = 0
			e.SetCursor(cursor)
		}
		switch parts[0] {
		case "q":
			c.mode = gott.ModeQuit
			return
		case "r":
			if len(parts) == 2 {
				filename := parts[1]
				e.ReadFile(filename)
			}
		case "debug":
			if len(parts) == 2 {
				if parts[1] == "on" {
					c.debug = true
				} else if parts[1] == "off" {
					c.debug = false
					c.message = ""
				}
			}
		case "w":
			var filename string
			if len(parts) == 2 {
				filename = parts[1]
			} else {
				filename = e.GetBuffer().GetFileName()
			}
			e.WriteFile(filename)
		case "wq":
			var filename string
			if len(parts) == 2 {
				filename = parts[1]
			} else {
				filename = e.GetBuffer().GetFileName()
			}
			e.WriteFile(filename)
			c.mode = gott.ModeQuit
			return
		case "fmt":
			out, err := e.Gofmt(e.GetBuffer().GetFileName(), e.Bytes())
			if err == nil {
				e.GetBuffer().ReadBytes(out)
			}
		case "$":
			newRow := e.GetBuffer().GetRowCount() - 1
			if newRow < 0 {
				newRow = 0
			}
			cursor := e.GetCursor()
			cursor.Row = newRow
			cursor.Col = 0
			e.SetCursor(cursor)
		default:
			c.message = "nope"
		}
	}
	c.command = ""
	c.mode = gott.ModeEdit
}

func (c *Commander) Multiplier() int {
	if c.multiplier == "" {
		return 1
	}
	i, err := strconv.ParseInt(c.multiplier, 10, 64)
	if err != nil {
		c.multiplier = ""
		return 1
	}
	c.multiplier = ""
	return int(i)
}

func (c *Commander) GetSearchText() string {
	return c.searchText
}

func (c *Commander) GetCommand() string {
	return c.command
}

func (c *Commander) GetMessage() string {
	return c.message
}
