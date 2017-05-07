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

// Delete a row

type DeleteRow struct {
	Op
}

func (op *DeleteRow) Perform(e gott.Editable, multiplier int) gott.Operation {
	e.MoveCursorToStartOfLine()
	op.init(e, multiplier)
	deletedText := e.DeleteRowsAtCursor(op.Multiplier)
	e.SetPasteBoard(deletedText, gott.PasteNewLine)
	inverse := &Insert{
		Position: gott.InsertAtCursor,
		Text:     deletedText,
	}
	inverse.copyForUndo(&op.Op)
	return inverse
}
