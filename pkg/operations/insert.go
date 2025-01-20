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
	gott "github.com/timburks/gott/pkg/types"
)

// Insert inserts text at a specified position relative to the cursor.
// In edit mode, insert invocations are followed by key inputs until
// the user exits insert mode. When used in undo operations, inserts
// are performed immediately.
type Insert struct {
	operation
	Position  int
	Text      string
	Inverse   *DeleteCharacter
	Commander gott.Commander
}

func (op *Insert) Perform(e gott.Editor, multiplier int) gott.Operation {
	op.init(e, multiplier)

	if op.Text != "" {
		e.SetCursor(op.Cursor)
	} else {
		op.Cursor = e.GetCursor()
		e.SetInsertOperation(op)
	}

	var newMode int
	op.Cursor, newMode = e.InsertText(op.Text, op.Position)
	if op.Commander != nil {
		op.Commander.SetMode(newMode)
	}

	inverse := &DeleteCharacter{}
	inverse.copyForUndo(&op.operation)
	inverse.Multiplier = len(op.Text)
	if op.Position == gott.InsertAtNewLineBelowCursor ||
		op.Position == gott.InsertAtNewLineAboveCursor {
		inverse.FinallyDeleteRow = true
	}
	op.Inverse = inverse
	return inverse
}

// Length returns the length of text added by the insert operation.
func (op *Insert) Length() int {
	return len(op.Text)
}

// AddCharacter adds a character to the insert operation.
func (op *Insert) AddCharacter(c rune) {
	op.Text += string(c)
}

// DeleteCharacter deletes a character from the end of the insert operation.
func (op *Insert) DeleteCharacter() {
	op.Text = op.Text[0 : len(op.Text)-1]
}

// Close completes an insert operation.
func (op *Insert) Close() {
	op.Inverse.Multiplier = len(op.Text)
}
