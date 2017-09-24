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
package editor

import (
	"fmt"

	gott "github.com/timburks/gott/types"
)

var lastWindowNumber = -1

// A Window is a view of a buffer.
type Window struct {
	buffer *Buffer
	number int
	origin gott.Point
	size   gott.Size
	cursor gott.Point // cursor position
	offset gott.Size  // display offset
}

func NewWindow() *Window {
	lastWindowNumber++
	w := &Window{}
	w.number = lastWindowNumber
	w.buffer = NewBuffer()
	return w
}

func (w *Window) GetBuffer() gott.Buffer {
	return w.buffer
}

func (w *Window) GetIndex() int {
	return w.number
}

// draw text in an area defined by origin and size with a specified offset into the buffer
func (w *Window) Render(display gott.Display) {
	w.adjustDisplayOffsetForScrolling()

	b := w.buffer
	if !b.Highlighted {
		switch b.languageMode {
		case "go":
			h := NewGoHighlighter()
			h.Highlight(b)
		}
		b.Highlighted = true
	}

	for i := w.origin.Row; i < w.origin.Row+w.size.Rows-1; i++ {
		var line string
		var colors []gott.Color
		if (i + w.offset.Rows) < len(b.rows) {
			line = b.rows[i+w.offset.Rows].DisplayText()
			colors = b.rows[i+w.offset.Rows].Colors
			if w.offset.Cols < len(line) {
				line = line[w.offset.Cols:]
				colors = colors[w.offset.Cols:]
			} else {
				line = ""
			}
		} else {
			line = "~"
			colors = make([]gott.Color, 1, 1)
			colors[0] = gott.ColorWhite
		}
		// truncate line to fit screen
		if len(line) > w.size.Cols {
			line = line[0:w.size.Cols]
			colors = colors[0:w.size.Cols]
		}
		for j, c := range line {
			var color gott.Color = gott.ColorWhite
			if j < len(colors) {
				color = colors[j]
			}
			display.SetCell(j+w.origin.Col, i, rune(c), color)
		}
	}

	// Draw the info bar as a single line at the bottom of the buffer window.
	infoText := w.computeInfoBarText(w.size.Cols)
	infoRow := w.origin.Row + w.size.Rows - 1
	for x, ch := range infoText {
		display.SetCellReversed(x, infoRow, rune(ch), gott.ColorBlack)
	}
}

// Compute the text to display on the info bar.
func (w *Window) computeInfoBarText(length int) string {
	b := w.buffer
	finalText := fmt.Sprintf(" %d/%d ", w.cursor.Row, b.GetRowCount())
	text := fmt.Sprintf(" [%d] %s", w.GetIndex(), b.GetName())
	if b.GetReadOnly() {
		text = text + "(read-only)"
	}
	for len(text) < length-len(finalText)-1 {
		text = text + " "
	}
	text += finalText
	return text
}

// Recompute the display offset to keep the cursor onscreen.
func (w *Window) adjustDisplayOffsetForScrolling() {
	if w.cursor.Row < w.offset.Rows {
		// scroll up
		w.offset.Rows = w.cursor.Row
	}
	// reserve the last row for the info bar
	textRows := w.size.Rows - 1
	if w.cursor.Row-w.offset.Rows >= textRows {
		// scroll down
		w.offset.Rows = w.cursor.Row - textRows + 1
	}
	if w.cursor.Col < w.offset.Cols {
		// scroll left
		w.offset.Cols = w.cursor.Col
	}
	if w.cursor.Col-w.offset.Cols >= w.size.Cols {
		// scroll right
		w.offset.Cols = w.cursor.Col - w.size.Cols + 1
	}
}

func (w *Window) SetCursor(d gott.Display) {
	d.SetCursor(gott.Point{
		Col: w.cursor.Col - w.offset.Cols,
		Row: w.cursor.Row - w.offset.Rows,
	})
}
