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
package window

import (
	"fmt"

	"github.com/nsf/termbox-go"

	"github.com/timburks/gott/editor"
	gott "github.com/timburks/gott/types"
)

// The Window draws the state of an Editor.
type Window struct {
	size gott.Size // screen size
}

func NewWindow() *Window {
	return &Window{}
}

func (w *Window) Render(e *editor.Editor, c gott.Commander) {
	termbox.Clear(termbox.ColorWhite, termbox.ColorBlack)
	var windowSize gott.Size
	windowSize.Cols, windowSize.Rows = termbox.Size()
	w.size = windowSize

	editSize := windowSize
	editSize.Rows -= 2
	e.SetSize(editSize)

	e.Scroll()
	w.RenderInfoBar(e, c)
	w.RenderMessageBar(e, c)
	bufferOrigin := gott.Point{Row: 0, Col: 0}
	bufferSize := gott.Size{Rows: w.size.Rows - 2, Cols: w.size.Cols}
	e.Buffer.Render(bufferOrigin, bufferSize, e.Offset)
	termbox.SetCursor(e.Cursor.Col-e.Offset.Cols, e.Cursor.Row-e.Offset.Rows)
	termbox.Flush()
}

func (w *Window) RenderInfoBar(e *editor.Editor, c gott.Commander) {
	finalText := fmt.Sprintf(" %d/%d ", e.Cursor.Row, e.Buffer.RowCount())
	text := " the gott editor - " + e.Buffer.FileName + " "
	for len(text) < w.size.Cols-len(finalText)-1 {
		text = text + " "
	}
	text += finalText
	for x, ch := range text {
		termbox.SetCell(x, w.size.Rows-2,
			rune(ch),
			termbox.ColorBlack, termbox.ColorWhite)
	}
}

func (w *Window) RenderMessageBar(e *editor.Editor, c gott.Commander) {
	var line string
	switch c.GetMode() {
	case gott.ModeCommand:
		line += ":" + c.GetCommand()
	case gott.ModeSearch:
		line += "/" + c.GetSearchText()
	default:
		line += c.GetMessage()
	}
	if len(line) > w.size.Cols {
		line = line[0:w.size.Cols]
	}
	for x, ch := range line {
		termbox.SetCell(x, w.size.Rows-1, rune(ch), termbox.ColorBlack, termbox.ColorWhite)
	}
}
