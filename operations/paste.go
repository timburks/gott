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

package operations

import (
	gott "github.com/timburks/gott/types"
)

// Paste pastes the contents of the pasteboard into a buffer.
type Paste struct {
	operation
}

func (op *Paste) Perform(e gott.Editor, multiplier int) gott.Operation {
	if e.GetPasteMode() == gott.PasteNewLine {
		e.MoveCursorToStartOfLineBelowCursor()
	}

	op.init(e, multiplier)

	cursor := op.Cursor

	for i := 0; i < op.Multiplier; i++ {
		for _, c := range e.GetPasteText() {
			e.InsertChar(c)
		}
	}
	if e.GetPasteMode() == gott.PasteNewLine {
		e.SetCursor(cursor)
		inverse := &DeleteCharacter{}
		inverse.copyForUndo(&op.operation)
		inverse.Multiplier = len(e.GetPasteText()) * op.Multiplier
		inverse.Cursor.Col = 0
		return inverse
	} else {
		return nil
	}
}
