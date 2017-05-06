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

type Commander struct {
	Editor *Editor

	Debug bool // debug mode displays information about events (key codes, etc)
}

func (c *Commander) ProcessEvent(event termbox.Event) error {
	editor := c.Editor
	if c.Debug {
		editor.Message = fmt.Sprintf("event=%+v", event)
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
	e := c.Editor

	// multikey commands have highest precedence
	if len(e.EditKeys) > 0 {
		key := event.Key
		ch := event.Ch
		switch e.EditKeys {
		case "d":
			switch ch {
			case 'd':
				e.Perform(&DeleteRow{})
				e.KeepCursorInRow()
			case 'w':
				e.Perform(&DeleteWord{})
				e.KeepCursorInRow()
			}
		case "r":
			if key != 0 {
				if key == termbox.KeySpace {
					e.Perform(&ReplaceCharacter{Character: rune(' ')})
				}
			} else if ch != 0 {
				e.Perform(&ReplaceCharacter{Character: rune(event.Ch)})
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
			e.Cursor.Row = e.Offset.Rows
			// move up by a page
			for i := 0; i < e.EditSize.Rows; i++ {
				e.MoveCursor(MoveUp)
			}
		case termbox.KeyPgdn:
			// move to the bottom of the screen
			e.Cursor.Row = e.Offset.Rows + e.EditSize.Rows - 1
			// move down by a page
			for i := 0; i < e.EditSize.Rows; i++ {
				e.MoveCursor(MoveDown)
			}
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
			e.Perform(&Insert{Position: InsertAtCursor})
		case 'a':
			e.Perform(&Insert{Position: InsertAfterCursor})
		case 'I':
			e.Perform(&Insert{Position: InsertAtStartOfLine})
		case 'A':
			e.Perform(&Insert{Position: InsertAfterEndOfLine})
		case 'o':
			e.Perform(&Insert{Position: InsertAtNewLineBelowCursor})
		case 'O':
			e.Perform(&Insert{Position: InsertAtNewLineAboveCursor})
		case 'x':
			e.Perform(&DeleteCharacter{})
		case 'p': // PasteText
			e.Perform(&Paste{})
		case '~': // reverse case
			e.Perform(&ReverseCaseCharacter{})
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
	e := c.Editor

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
	e := c.Editor

	key := event.Key
	if key != 0 {
		switch key {
		case termbox.KeyEsc:
			e.Mode = ModeEdit
		case termbox.KeyEnter:
			c.PerformCommand()
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

func (c *Commander) ProcessKeySearchMode(event termbox.Event) error {
	e := c.Editor

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

func (c *Commander) ProcessKey(event termbox.Event) error {
	e := c.Editor
	var err error
	switch e.Mode {
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
	e := c.Editor

	parts := strings.Split(e.Command, " ")
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
					c.Debug = true
				} else if parts[1] == "off" {
					c.Debug = false
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
			e.Cursor.Row = len(e.Buffer.Rows) - 1
			if e.Cursor.Row < 0 {
				e.Cursor.Row = 0
			}
		default:
			e.Message = "nope"
		}
	}
	e.Command = ""
	e.Mode = ModeEdit
}
