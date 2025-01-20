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

	gott "github.com/timburks/gott/pkg/types"
)

// The Commander converts user input into commands to the editor.
type Commander struct {
	editor         gott.Editor
	batch          bool     // true if commander is running a lisp script
	mode           int      // editor mode
	debug          bool     // debug mode displays information about events (key codes, etc)
	editKeys       string   // edit key sequences in progress
	commandText    string   // command as it is being typed on the command line
	searchText     string   // text for searches as it is being typed
	searchForward  bool     // true to search forward, false to search backward
	lispText       string   // lisp command as it is being typed
	multiplierText string   // multiplier string as it is being entered
	message        string   // status message
	lastKey        gott.Key // last key pressed
	lastCh         rune     // last character pressed (if key == 0)
}

func NewCommander(e gott.Editor) *Commander {
	return &Commander{editor: e, mode: gott.ModeEdit}
}

func (c *Commander) getLastKey() gott.Key {
	return c.lastKey
}

func (c *Commander) getLastCh() rune {
	return c.lastCh
}

func (c *Commander) getMode() int {
	return c.mode
}

func (c *Commander) SetMode(m int) {
	c.mode = m
}

func (c *Commander) getModeName() string {
	switch c.mode {
	case gott.ModeEdit:
		return "edit"
	case gott.ModeInsert:
		return "insert"
	case gott.ModeCommand:
		return "command"
	case gott.ModeSearchForward:
		return "search-forward"
	case gott.ModeSearchBackward:
		return "search-backward"
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
		return c.processKey(event)
	case gott.EventResize:
		return c.processResize(event)
	default:
		return nil
	}
}

func (c *Commander) processResize(event *gott.Event) error {
	return nil
}

func (c *Commander) processKeyEditMode(event *gott.Event) error {
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
				c.parseEval("(change-word)")
			}
		case "d":
			switch ch {
			case 'd':
				c.parseEval("(delete-row)")
			case 'w':
				c.parseEval("(delete-word)")
			}
		case "r":
			if (key != 0 && key == gott.KeySpace) || (ch != 0) {
				c.parseEval("(replace-character)")
			}
		case "y":
			switch ch {
			case 'y': // YankRow
				c.parseEval("(yank-row)")
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
			c.parseEval("(page-up)")
		case gott.KeyCtrlF, gott.KeyPgdn:
			c.parseEval("(page-down)")
		case gott.KeyCtrlD:
			c.parseEval("(half-page-down)")
		case gott.KeyCtrlU:
			c.parseEval("(half-page-up)")
		case gott.KeyCtrlA, gott.KeyHome:
			c.parseEval("(beginning-of-line)")
		case gott.KeyCtrlE, gott.KeyEnd:
			c.parseEval("(end-of-line)")
		case gott.KeyArrowUp:
			c.parseEval("(up)")
		case gott.KeyArrowDown:
			c.parseEval("(down)")
		case gott.KeyArrowLeft:
			c.parseEval("(left)")
		case gott.KeyArrowRight:
			c.parseEval("(right)")
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
			c.parseEval("(command-mode)")
		//
		// lisp commands go to the message bar
		//
		case '(':
			c.parseEval("(lisp-mode)")
		//
		// search queries go to the message bar
		//
		case '/':
			c.parseEval("(search-forward-mode)")
		case '?':
			c.parseEval("(search-backward-mode)")
		//
		// repeat the last search
		//
		case 'n':
			if c.searchForward {
				c.parseEval("(repeat-search-forward)")
			} else {
				c.parseEval("(repeat-search-backward)")
			}
		//
		// cursor movement isn't logged
		//
		case 'h':
			c.parseEval("(left)")
		case 'j':
			c.parseEval("(down)")
		case 'k':
			c.parseEval("(up)")
		case 'l':
			c.parseEval("(right)")
		case 'w':
			c.parseEval("(next-word)")
		case 'b':
			c.parseEval("(previous-word)")
		case '>':
			c.parseEval("(change-window)")
		//
		// "performed" operations are saved for undo and repetition
		//
		case 'i':
			c.parseEval("(insert-at-cursor)")
		case 'a':
			c.parseEval("(insert-after-cursor)")
		case 'I':
			c.parseEval("(insert-at-start-of-line)")
		case 'A':
			c.parseEval("(insert-after-end-of-line)")
		case 'o':
			c.parseEval("(insert-at-new-line-below-cursor)")
		case 'O':
			c.parseEval("(insert-at-new-line-above-cursor)")
		case 'x':
			c.parseEval("(delete-character)")
		case 'J':
			c.parseEval("(join-line)")
		case 'p':
			c.parseEval("(paste)")
		case '~':
			c.parseEval("(reverse-case-character)")
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
			c.parseEval("(undo)")
		//
		// repeat
		//
		case '.':
			c.parseEval("(repeat)")
		}
	}
	return nil
}

func (c *Commander) processKeyInsertMode(event *gott.Event) error {
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

func (c *Commander) processKeyCommandMode(event *gott.Event) error {
	key := event.Key
	ch := event.Ch
	if key != 0 {
		switch key {
		case gott.KeyEsc:
			c.mode = gott.ModeEdit
		case gott.KeyEnter:
			c.performCommand()
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

func (c *Commander) processKeySearchMode(event *gott.Event) error {
	e := c.editor

	key := event.Key
	ch := event.Ch
	if key != 0 {
		switch key {
		case gott.KeyEsc:
			c.mode = gott.ModeEdit
		case gott.KeyEnter:
			if c.mode == gott.ModeSearchForward {
				c.searchForward = true
				e.PerformSearchForward(c.searchText)
			} else {
				c.searchForward = false
				e.PerformSearchBackward(c.searchText)
			}
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

func (c *Commander) processKeyLispMode(event *gott.Event) error {
	key := event.Key
	ch := event.Ch
	if key != 0 {
		switch key {
		case gott.KeyEsc:
			c.mode = gott.ModeEdit
		case gott.KeyEnter:
			c.message = c.parseEval(c.lispText)
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

func (c *Commander) processKey(event *gott.Event) error {
	var err error
	switch c.mode {
	case gott.ModeEdit:
		err = c.processKeyEditMode(event)
	case gott.ModeInsert:
		err = c.processKeyInsertMode(event)
	case gott.ModeCommand:
		err = c.processKeyCommandMode(event)
	case gott.ModeSearchForward:
		err = c.processKeySearchMode(event)
	case gott.ModeSearchBackward:
		err = c.processKeySearchMode(event)
	case gott.ModeLisp:
		err = c.processKeyLispMode(event)
	}
	return err
}

func (c *Commander) performCommand() {

	e := c.editor

	parts := strings.Split(c.commandText, " ")
	if len(parts) > 0 {

		i, err := strconv.ParseInt(parts[0], 10, 64)
		if err == nil {
			e.MoveCursorToLine(int(i))
		}
		switch parts[0] {
		case "q":
			c.mode = gott.ModeQuit
			return
		case "quit":
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
				filename = e.GetFileName()
			}
			e.WriteFile(filename)
		case "wq":
			var filename string
			if len(parts) == 2 {
				filename = parts[1]
			} else {
				filename = e.GetFileName()
			}
			e.WriteFile(filename)
			c.mode = gott.ModeQuit
			return
		case "fmt":
			out, err := e.Gofmt(e.GetFileName(), e.Bytes())
			if err == nil {
				e.LoadBytes(out)
			}
		case "$":
			e.MoveCursorToLine(1e9)
		case "cursor":
			cursor := e.GetCursor()
			c.message = fmt.Sprintf("%d,%d", cursor.Row, cursor.Col)
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
		case "next": // switch to next window
			e.SelectWindowNext()
		case "prev": // switch to previous window
			e.SelectWindowPrevious()
		case "windows":
			e.ListWindows()
		case "clear":
			e.LoadBytes([]byte{})
		case "eval":
			output := c.parseEval(string(e.Bytes()))
			e.SelectWindow(0)
			e.AppendBytes([]byte(output))
		case "split":
			e.SplitWindowVertically()
		case "vsplit":
			e.SplitWindowVertically()
		case "hsplit":
			e.SplitWindowHorizontally()
		case "close":
			e.CloseActiveWindow()
		case "layout":
			e.LayoutWindows()
		default:
			c.message = ""
		}
	}
	c.commandText = ""
	c.mode = gott.ModeEdit
}

func (c *Commander) getMultiplier() int {
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

func (c *Commander) getSearchText() string {
	return c.searchText
}

func (c *Commander) getLispText() string {
	return c.lispText
}

func (c *Commander) getCommandText() string {
	return c.commandText
}

func (c *Commander) getMessage() string {
	return c.message
}

func (c *Commander) GetMessageBarText(length int) string {
	var line string
	switch c.getMode() {
	case gott.ModeCommand:
		line += ":" + c.getCommandText()
	case gott.ModeSearchForward:
		line += "/" + c.getSearchText()
	case gott.ModeSearchBackward:
		line += "?" + c.getSearchText()
	case gott.ModeLisp:
		line += c.getLispText()
	default:
		line += c.getMessage()
	}
	if len(line) > length {
		line = line[0:length]
	}
	return line
}
