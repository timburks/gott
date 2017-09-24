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
package screen

import (
	"log"

	"github.com/nsf/termbox-go"
	gott "github.com/timburks/gott/types"
)

// The Screen draws the state of an Editor.
type Screen struct {
	size   gott.Size // screen size
	editor gott.Editor
}

func NewScreen(e gott.Editor) *Screen {
	// Open the terminal.
	err := termbox.Init()
	if err != nil {
		log.Output(1, err.Error())
		return nil
	}
	termbox.SetOutputMode(termbox.Output256)
	s := &Screen{editor: e}
	s.Resize()
	return s
}

func (s *Screen) Close() {
	termbox.Close()
}

func (s *Screen) Resize() {
	var screenSize gott.Size
	screenSize.Cols, screenSize.Rows = termbox.Size()
	s.size = screenSize
	editSize := screenSize
	editSize.Rows -= 1
	s.editor.SetSize(editSize)
	s.editor.LayoutWindows()
}

func (s *Screen) Render(e gott.Editor, c gott.Commander) {
	termbox.Clear(termbox.ColorWhite, termbox.ColorBlack)
	e.RenderWindows(s)
	s.RenderMessageBar(c)
	termbox.Flush()
}

func (s *Screen) SetCell(j int, i int, c rune, color gott.Color) {
	termbox.SetCell(j, i, c, termbox.Attribute(color), 0x01)
}

func (s *Screen) SetCellReversed(j int, i int, c rune, color gott.Color) {
	termbox.SetCell(j, i, c, termbox.Attribute(color), 0x08)
}

func (s *Screen) SetCursor(position gott.Point) {
	termbox.SetCursor(position.Col, position.Row)
}

// The message bar is a single line at the bottom of the screen.
func (s *Screen) RenderMessageBar(c gott.Commander) {
	text := c.GetMessageBarText(s.size.Cols)
	for x, ch := range text {
		termbox.SetCell(x, s.size.Rows-1,
			rune(ch),
			termbox.ColorWhite,
			termbox.ColorBlack)
	}
}

func (s *Screen) GetNextEvent() *gott.Event {
	event := termbox.PollEvent()
	if event.Type == termbox.EventResize {
		s.Resize()
		termbox.Flush()
	}
	return &gott.Event{
		Type: int(event.Type),
		Key:  key(event.Key),
		Ch:   event.Ch,
	}
}

func key(k termbox.Key) gott.Key {
	switch k {
	case termbox.KeyArrowDown:
		return gott.KeyArrowDown
	case termbox.KeyArrowLeft:
		return gott.KeyArrowLeft
	case termbox.KeyArrowRight:
		return gott.KeyArrowRight
	case termbox.KeyArrowUp:
		return gott.KeyArrowUp
	case termbox.KeyBackspace2:
		return gott.KeyBackspace2
	case termbox.KeyCtrlA:
		return gott.KeyCtrlA
	case termbox.KeyCtrlB:
		return gott.KeyCtrlB
	case termbox.KeyCtrlC:
		return gott.KeyCtrlC
	case termbox.KeyCtrlD:
		return gott.KeyCtrlD
	case termbox.KeyCtrlE:
		return gott.KeyCtrlE
	case termbox.KeyCtrlF:
		return gott.KeyCtrlF
	case termbox.KeyCtrlG:
		return gott.KeyCtrlG
	case termbox.KeyCtrlH:
		return gott.KeyCtrlH
	//case termbox.KeyCtrlI:
	//	return gott.KeyCtrlI
	case termbox.KeyCtrlJ:
		return gott.KeyCtrlJ
	case termbox.KeyCtrlK:
		return gott.KeyCtrlK
	case termbox.KeyCtrlL:
		return gott.KeyCtrlL
	//case termbox.KeyCtrlM:
	//	return gott.KeyCtrlM
	case termbox.KeyCtrlN:
		return gott.KeyCtrlN
	case termbox.KeyCtrlO:
		return gott.KeyCtrlO
	case termbox.KeyCtrlP:
		return gott.KeyCtrlP
	case termbox.KeyCtrlQ:
		return gott.KeyCtrlQ
	case termbox.KeyCtrlR:
		return gott.KeyCtrlR
	case termbox.KeyCtrlS:
		return gott.KeyCtrlS
	case termbox.KeyCtrlT:
		return gott.KeyCtrlT
	case termbox.KeyCtrlU:
		return gott.KeyCtrlU
	case termbox.KeyCtrlV:
		return gott.KeyCtrlV
	case termbox.KeyCtrlW:
		return gott.KeyCtrlW
	case termbox.KeyCtrlX:
		return gott.KeyCtrlX
	case termbox.KeyCtrlY:
		return gott.KeyCtrlY
	case termbox.KeyCtrlZ:
		return gott.KeyCtrlZ
	case termbox.KeyEnd:
		return gott.KeyEnd
	case termbox.KeyEnter:
		return gott.KeyEnter
	case termbox.KeyEsc:
		return gott.KeyEsc
	case termbox.KeyHome:
		return gott.KeyHome
	case termbox.KeyPgdn:
		return gott.KeyPgdn
	case termbox.KeyPgup:
		return gott.KeyPgup
	case termbox.KeySpace:
		return gott.KeySpace
	case termbox.KeyTab:
		return gott.KeyTab
	default:
		return gott.KeyUnsupported
	}
}
