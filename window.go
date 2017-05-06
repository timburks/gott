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

	"github.com/nsf/termbox-go"
)

// The Window draws the state of an Editor.
type Window struct {
	size Size // screen size
}

func NewWindow() *Window {
	return &Window{}
}

func (window *Window) Render(e *Editor, c *Commander) {
	termbox.Clear(termbox.ColorWhite, termbox.ColorBlack)
	var windowSize Size
	windowSize.Cols, windowSize.Rows = termbox.Size()
	window.size = windowSize

	editSize := windowSize
	editSize.Rows -= 2
	e.SetSize(editSize)

	e.Scroll()
	window.RenderInfoBar(e, c)
	window.RenderMessageBar(e, c)
	bufferOrigin := Point{Row: 0, Col: 0}
	bufferSize := Size{Rows: window.size.Rows - 2, Cols: window.size.Cols}
	e.Buffer.Render(bufferOrigin, bufferSize, e.Offset)
	termbox.SetCursor(e.Cursor.Col-e.Offset.Cols, e.Cursor.Row-e.Offset.Rows)
	termbox.Flush()
}

func (window *Window) RenderInfoBar(e *Editor, c *Commander) {
	finalText := fmt.Sprintf(" %d/%d ", e.Cursor.Row, len(e.Buffer.rows))
	text := " the gott editor - " + e.Buffer.FileName + " "
	for len(text) < window.size.Cols-len(finalText)-1 {
		text = text + " "
	}
	text += finalText
	for x, ch := range text {
		termbox.SetCell(x, window.size.Rows-2,
			rune(ch),
			termbox.ColorBlack, termbox.ColorWhite)
	}
}

func (window *Window) RenderMessageBar(e *Editor, c *Commander) {
	var line string
	switch c.GetMode() {
	case ModeCommand:
		line += ":" + c.Command()
	case ModeSearch:
		line += "/" + c.SearchText()
	default:
		line += c.Message()
	}
	if len(line) > window.size.Cols {
		line = line[0:window.size.Cols]
	}
	for x, ch := range line {
		termbox.SetCell(x, window.size.Rows-1, rune(ch), termbox.ColorBlack, termbox.ColorWhite)
	}
}
