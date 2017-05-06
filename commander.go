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
	"strconv"
	"strings"

	"github.com/nsf/termbox-go"
)

// The Commander converts user input into commands for the Editor.
type Commander struct {
	editor     *Editor
	mode       int    // editor mode
	debug      bool   // debug mode displays information about events (key codes, etc)
	editKeys   string // edit key sequences in progress
	command    string // command as it is being typed on the command line
	searchText string // text for searches as it is being typed
	message    string // status message
	multiplier string // multiplier string as it is being entered
}

func NewCommander(e *Editor) *Commander {
	return &Commander{editor: e, mode: ModeEdit}
}

func (c *Commander) GetMode() int {
	return c.mode
}

func (c *Commander) SetMode(m int) {
	c.mode = m
}

func (c *Commander) ProcessEvent(event termbox.Event) error {
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

	// multikey commands have highest precedence
	if len(c.editKeys) > 0 {
		key := event.Key
		ch := event.Ch
		switch c.editKeys {
		case "d":
			switch ch {
			case 'd':
				e.Perform(&DeleteRow{}, c.Multiplier())
				e.KeepCursorInRow()
			case 'w':
				e.Perform(&DeleteWord{}, c.Multiplier())
				e.KeepCursorInRow()
			}
		case "r":
			if key != 0 {
				if key == termbox.KeySpace {
					e.Perform(&ReplaceCharacter{Character: rune(' ')}, c.Multiplier())
				}
			} else if ch != 0 {
				e.Perform(&ReplaceCharacter{Character: rune(event.Ch)}, c.Multiplier())
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
	key := event.Key
	if key != 0 {
		switch key {
		case termbox.KeyEsc:
			break
		case termbox.KeyPgup:
			e.PageUp()
		case termbox.KeyPgdn:
			e.PageDown()
		case termbox.KeyCtrlA, termbox.KeyHome:
			// move to beginning of line
			e.Cursor.Col = 0
		case termbox.KeyCtrlE, termbox.KeyEnd:
			// move to end of line
			e.Cursor.Col = 0
			if e.Cursor.Row < len(e.Buffer.Rows) {
				e.Cursor.Col = e.Buffer.Rows[e.Cursor.Row].Length() - 1
				if e.Cursor.Col < 0 {
					e.Cursor.Col = 0
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
			c.multiplier += string(ch)
		//
		// commands go to the message bar
		//
		case ':':
			c.mode = ModeCommand
			c.command = ""
		//
		// search queries go to the message bar
		//
		case '/':
			c.mode = ModeSearch
			c.searchText = ""
		case 'n': // repeat the last search
			e.PerformSearch(c.searchText)
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
			e.Perform(&Insert{Position: InsertAtCursor, Commander: c}, c.Multiplier())
		case 'a':
			e.Perform(&Insert{Position: InsertAfterCursor, Commander: c}, c.Multiplier())
		case 'I':
			e.Perform(&Insert{Position: InsertAtStartOfLine, Commander: c}, c.Multiplier())
		case 'A':
			e.Perform(&Insert{Position: InsertAfterEndOfLine, Commander: c}, c.Multiplier())
		case 'o':
			e.Perform(&Insert{Position: InsertAtNewLineBelowCursor, Commander: c}, c.Multiplier())
		case 'O':
			e.Perform(&Insert{Position: InsertAtNewLineAboveCursor, Commander: c}, c.Multiplier())
		case 'x':
			e.Perform(&DeleteCharacter{}, c.Multiplier())
		case 'p': // PasteText
			e.Perform(&Paste{}, c.Multiplier())
		case '~': // reverse case
			e.Perform(&ReverseCaseCharacter{}, c.Multiplier())
		//
		// a few keys open multi-key commands
		//
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
			c.mode = ModeEdit
			e.KeepCursorInRow()
		case termbox.KeyBackspace2:
			e.BackspaceChar()
		case termbox.KeyTab:
			e.InsertChar(' ')
			for {
				if e.Cursor.Col%8 == 0 {
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
			c.mode = ModeEdit
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
			c.mode = ModeEdit
		case termbox.KeyEnter:
			e.PerformSearch(c.searchText)
			c.mode = ModeEdit
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
	case ModeEdit:
		err = c.ProcessKeyEditMode(event)
	case ModeInsert:
		err = c.ProcessKeyInsertMode(event)
	case ModeCommand:
		err = c.ProcessKeyCommandMode(event)
	case ModeSearch:
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
			e.Cursor.Row = int(i - 1)
			if e.Cursor.Row > len(e.Buffer.Rows)-1 {
				e.Cursor.Row = len(e.Buffer.Rows) - 1
			}
			if e.Cursor.Row < 0 {
				e.Cursor.Row = 0
			}
		}
		switch parts[0] {
		case "q":
			c.mode = ModeQuit
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
			c.mode = ModeQuit
			return
		case "fmt":
			out, err := gofmt(e.Buffer.FileName, e.Bytes())
			if err == nil {
				e.Buffer.ReadBytes(out)
			}
		case "$":
			e.Cursor.Row = len(e.Buffer.Rows) - 1
			if e.Cursor.Row < 0 {
				e.Cursor.Row = 0
			}
		default:
			c.message = "nope"
		}
	}
	c.command = ""
	c.mode = ModeEdit
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

func (c *Commander) SearchText() string {
	return c.searchText
}

func (c *Commander) Command() string {
	return c.command
}

func (c *Commander) Message() string {
	return c.message
}
