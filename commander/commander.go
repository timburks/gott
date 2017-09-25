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

	gott "github.com/timburks/gott/types"
)

// The Commander converts user input into commands for the Editor.
type Commander struct {
	editor         gott.Editor
	batch          bool     // true if commander is running a lisp script
	mode           int      // editor mode
	debug          bool     // debug mode displays information about events (key codes, etc)
	editKeys       string   // edit key sequences in progress
	commandText    string   // command as it is being typed on the command line
	searchText     string   // text for searches as it is being typed
	lispText       string   // lisp command as it is being typed
	multiplierText string   // multiplier string as it is being entered
	message        string   // status message
	lastKey        gott.Key // last key pressed
	lastCh         rune     // last character pressed (if key == 0)
}

func NewCommander(e gott.Editor) *Commander {
	return &Commander{editor: e, mode: gott.ModeEdit}
}

func (c *Commander) GetLastKey() gott.Key {
	return c.lastKey
}

func (c *Commander) GetLastCh() rune {
	return c.lastCh
}

func (c *Commander) GetMode() int {
	return c.mode
}

func (c *Commander) SetMode(m int) {
	c.mode = m
}

func (c *Commander) GetModeName() string {
	switch c.mode {
	case gott.ModeEdit:
		return "edit"
	case gott.ModeInsert:
		return "insert"
	case gott.ModeCommand:
		return "command"
	case gott.ModeSearchForward:
		return "search"
	case gott.ModeLisp:
		return "lisp"
	case gott.ModeQuit:
		return "quit"
	default:
		return "unknown"
	}
}

func (c *Commander) IsRunning() bool {
	return c.mode != gott.ModeQuit
}

func (c *Commander) ProcessEvent(event *gott.Event) error {
	if c.debug {
		c.message = fmt.Sprintf("event=%+v", event)
	}
	switch event.Type {
	case gott.EventKey:
		return c.ProcessKey(event)
	case gott.EventResize:
		return c.ProcessResize(event)
	default:
		return nil
	}
}

func (c *Commander) ProcessResize(event *gott.Event) error {
	return nil
}

func (c *Commander) ProcessKeyEditMode(event *gott.Event) error {
	key := event.Key
	ch := event.Ch

	c.lastKey = event.Key
	c.lastCh = event.Ch

	// multikey commands have highest precedence
	if len(c.editKeys) > 0 {
		switch c.editKeys {
		case "c":
			switch ch {
			case 'w':
				c.ParseEval("(change-word)")
			}
		case "d":
			switch ch {
			case 'd':
				c.ParseEval("(delete-row)")
			case 'w':
				c.ParseEval("(delete-word)")
			}
		case "r":
			if (key != 0 && key == gott.KeySpace) || (ch != 0) {
				c.ParseEval("(replace-character)")
			}
		case "y":
			switch ch {
			case 'y': // YankRow
				c.ParseEval("(yank-row)")
			default:
				break
			}
		}
		c.editKeys = ""
		return nil
	}
	if key != 0 {
		switch key {
		case gott.KeyEsc:
			break
		case gott.KeyCtrlB, gott.KeyPgup:
			c.ParseEval("(page-up)")
		case gott.KeyCtrlF, gott.KeyPgdn:
			c.ParseEval("(page-down)")
		case gott.KeyCtrlD:
			c.ParseEval("(half-page-down)")
		case gott.KeyCtrlU:
			c.ParseEval("(half-page-up)")
		case gott.KeyCtrlA, gott.KeyHome:
			c.ParseEval("(beginning-of-line)")
		case gott.KeyCtrlE, gott.KeyEnd:
			c.ParseEval("(end-of-line)")
		case gott.KeyArrowUp:
			c.ParseEval("(up)")
		case gott.KeyArrowDown:
			c.ParseEval("(down)")
		case gott.KeyArrowLeft:
			c.ParseEval("(left)")
		case gott.KeyArrowRight:
			c.ParseEval("(right)")
		}
	}
	if ch != 0 {
		switch ch {
		//
		// command multipliers are saved when operations are created
		//
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			c.multiplierText += string(ch)
		//
		// commands go to the message bar
		//
		case ':':
			c.ParseEval("(command-mode)")
		//
		// lisp commands go to the message bar
		//
		case '(':
			c.ParseEval("(lisp-mode)")
		//
		// search queries go to the message bar
		//
		case '/':
			c.ParseEval("(search-mode)")
		//
		// repeat the last search
		//
		case 'n':
			c.ParseEval("(repeat-search)")
		//
		// cursor movement isn't logged
		//
		case 'h':
			c.ParseEval("(left)")
		case 'j':
			c.ParseEval("(down)")
		case 'k':
			c.ParseEval("(up)")
		case 'l':
			c.ParseEval("(right)")
		case 'w':
			c.ParseEval("(next-word)")
		case 'b':
			c.ParseEval("(previous-word)")
		//
		// "performed" operations are saved for undo and repetition
		//
		case 'i':
			c.ParseEval("(insert-at-cursor)")
		case 'a':
			c.ParseEval("(insert-after-cursor)")
		case 'I':
			c.ParseEval("(insert-at-start-of-line)")
		case 'A':
			c.ParseEval("(insert-after-end-of-line)")
		case 'o':
			c.ParseEval("(insert-at-new-line-below-cursor)")
		case 'O':
			c.ParseEval("(insert-at-new-line-above-cursor)")
		case 'x':
			c.ParseEval("(delete-character)")
		case 'J':
			c.ParseEval("(join-line)")
		case 'p':
			c.ParseEval("(paste)")
		case '~':
			c.ParseEval("(reverse-case-character)")
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
			c.ParseEval("(undo)")
		//
		// repeat
		//
		case '.':
			c.ParseEval("(repeat)")
		}
	}
	return nil
}

func (c *Commander) ProcessKeyInsertMode(event *gott.Event) error {
	e := c.editor

	key := event.Key
	ch := event.Ch
	if key != 0 {
		switch key {
		case gott.KeyEsc: // end an insert operation.
			e.CloseInsert()
			c.mode = gott.ModeEdit
			e.KeepCursorInRow()
		case gott.KeyBackspace2:
			e.BackspaceChar()
		case gott.KeyTab:
			e.InsertChar(' ')
			for {
				if e.GetCursor().Col%8 == 0 {
					break
				}
				e.InsertChar(' ')
			}
		case gott.KeyEnter:
			e.InsertChar('\n')
		case gott.KeySpace:
			e.InsertChar(' ')
		}
	}
	if ch != 0 {
		e.InsertChar(ch)
	}
	return nil
}

func (c *Commander) ProcessKeyCommandMode(event *gott.Event) error {
	key := event.Key
	ch := event.Ch
	if key != 0 {
		switch key {
		case gott.KeyEsc:
			c.mode = gott.ModeEdit
		case gott.KeyEnter:
			c.PerformCommand()
		case gott.KeyBackspace2:
			if len(c.commandText) > 0 {
				c.commandText = c.commandText[0 : len(c.commandText)-1]
			}
		case gott.KeySpace:
			c.commandText += " "
		}
	}
	if ch != 0 {
		c.commandText = c.commandText + string(ch)
	}
	return nil
}

func (c *Commander) ProcessKeySearchMode(event *gott.Event) error {
	e := c.editor

	key := event.Key
	ch := event.Ch
	if key != 0 {
		switch key {
		case gott.KeyEsc:
			c.mode = gott.ModeEdit
		case gott.KeyEnter:
			e.PerformSearch(c.searchText)
			c.mode = gott.ModeEdit
		case gott.KeyBackspace2:
			if len(c.searchText) > 0 {
				c.searchText = c.searchText[0 : len(c.searchText)-1]
			}
		case gott.KeySpace:
			c.searchText += " "
		}
	}
	if ch != 0 {
		c.searchText = c.searchText + string(ch)
	}
	return nil
}

func (c *Commander) ProcessKeyLispMode(event *gott.Event) error {
	key := event.Key
	ch := event.Ch
	if key != 0 {
		switch key {
		case gott.KeyEsc:
			c.mode = gott.ModeEdit
		case gott.KeyEnter:
			c.message = c.ParseEval(c.lispText)
			// if evaluation didn't change the mode, set it back to edit
			if c.mode == gott.ModeLisp {
				c.mode = gott.ModeEdit
			}
		case gott.KeyBackspace2:
			if len(c.lispText) > 0 {
				c.lispText = c.lispText[0 : len(c.lispText)-1]
			}
		case gott.KeySpace:
			c.lispText += " "
		}
	}
	if ch != 0 {
		c.lispText = c.lispText + string(ch)
	}

	return nil
}

func (c *Commander) ProcessKey(event *gott.Event) error {
	var err error
	switch c.mode {
	case gott.ModeEdit:
		err = c.ProcessKeyEditMode(event)
	case gott.ModeInsert:
		err = c.ProcessKeyInsertMode(event)
	case gott.ModeCommand:
		err = c.ProcessKeyCommandMode(event)
	case gott.ModeSearchForward:
		err = c.ProcessKeySearchMode(event)
	case gott.ModeLisp:
		err = c.ProcessKeyLispMode(event)
	}
	return err
}

func (c *Commander) PerformCommand() {

	e := c.editor

	parts := strings.Split(c.commandText, " ")
	if len(parts) > 0 {

		i, err := strconv.ParseInt(parts[0], 10, 64)
		if err == nil {
			newRow := int(i - 1)
			if newRow > e.GetActiveWindow().GetBuffer().GetRowCount()-1 {
				newRow = e.GetActiveWindow().GetBuffer().GetRowCount() - 1
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
				filename = e.GetActiveWindow().GetBuffer().GetFileName()
			}
			e.WriteFile(filename)
		case "wq":
			var filename string
			if len(parts) == 2 {
				filename = parts[1]
			} else {
				filename = e.GetActiveWindow().GetBuffer().GetFileName()
			}
			e.WriteFile(filename)
			c.mode = gott.ModeQuit
			return
		case "fmt":
			out, err := e.Gofmt(e.GetActiveWindow().GetBuffer().GetFileName(), e.Bytes())
			if err == nil {
				e.GetActiveWindow().GetBuffer().LoadBytes(out)
			}
		case "$":
			newRow := e.GetActiveWindow().GetBuffer().GetRowCount() - 1
			if newRow < 0 {
				newRow = 0
			}
			cursor := e.GetCursor()
			cursor.Row = newRow
			cursor.Col = 0
			e.SetCursor(cursor)
		case "window":
			if len(parts) > 1 {
				number, err := strconv.Atoi(parts[1])
				if err == nil {
					err = e.SelectWindow(number)
					if err != nil {
						c.message = err.Error()
					} else {
						c.message = ""
					}
				} else {
					c.message = err.Error()
				}
			}
		case "next": // switch to next buffer
			e.SelectWindowNext()
		case "prev": // switch to previous buffer
			e.SelectWindowPrevious()
		case "windows":
			e.ListWindows()
		case "clear":
			e.GetActiveWindow().GetBuffer().LoadBytes([]byte{})
		case "eval":
			output := c.ParseEval(string(e.Bytes()))
			e.SelectWindow(0)
			e.GetActiveWindow().GetBuffer().AppendBytes([]byte(output))
		case "split":
			e.SplitWindowVertically()
		case "vsplit":
			e.SplitWindowVertically()
		case "hsplit":
			e.SplitWindowHorizontally()
		default:
			c.message = ""
		}
	}
	c.commandText = ""
	c.mode = gott.ModeEdit
}

func (c *Commander) GetMultiplier() int {
	if c.multiplierText == "" {
		return 1
	}
	i, err := strconv.ParseInt(c.multiplierText, 10, 64)
	if err != nil {
		c.multiplierText = ""
		return 1
	}
	c.multiplierText = ""
	return int(i)
}

func (c *Commander) GetSearchText() string {
	return c.searchText
}

func (c *Commander) GetLispText() string {
	return c.lispText
}

func (c *Commander) GetCommandText() string {
	return c.commandText
}

func (c *Commander) GetMessage() string {
	return c.message
}

func (c *Commander) GetMessageBarText(length int) string {
	var line string
	switch c.GetMode() {
	case gott.ModeCommand:
		line += ":" + c.GetCommandText()
	case gott.ModeSearchForward:
		line += "/" + c.GetSearchText()
	case gott.ModeSearchBackward:
		line += "?" + c.GetSearchText()
	case gott.ModeLisp:
		line += c.GetLispText()
	default:
		line += c.GetMessage()
	}
	if len(line) > length {
		line = line[0:length]
	}
	return line
}
