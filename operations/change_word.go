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

// Delete a word

type ChangeWord struct {
	operation
	Text      string
	Inverse   *DeleteCharacter
	Commander gott.Commander
}

func (op *ChangeWord) Perform(e gott.Editor, multiplier int) gott.Operation {
	op.init(e, multiplier)

	if op.Text != "" {
		e.SetCursor(op.Cursor)
	} else {
		op.Cursor = e.GetCursor()
		e.SetInsertOperation(op)
	}

	deletedText, newMode := e.ChangeWordAtCursor(op.Multiplier, op.Text)
	if op.Commander != nil {
		op.Commander.SetMode(newMode)
	}

	delete := &DeleteCharacter{}
	delete.copyForUndo(&op.operation)
	delete.Multiplier = len(op.Text)
	op.Inverse = delete

	reinsert := &Insert{
		Position: gott.InsertAtCursor,
		Text:     string(deletedText),
	}
	reinsert.copyForUndo(&op.operation)
	reinsert.Multiplier = 1

	operations := make([]gott.Operation, 0)
	// first delete inserted characters
	operations = append(operations, delete)

	// then reinsert deleted words
	operations = append(operations, reinsert)
	inverse := &Sequence{
		Operations: operations,
	}
	inverse.copyForUndo(&op.operation)
	inverse.Multiplier = 1
	return inverse
}

func (op *ChangeWord) Length() int {
	return len(op.Text)
}

func (op *ChangeWord) AddCharacter(c rune) {
	op.Text += string(c)
}

func (op *ChangeWord) DeleteCharacter() {
	op.Text = op.Text[0 : len(op.Text)-1]
}

func (op *ChangeWord) Close() {
	op.Inverse.Multiplier = len(op.Text)
}
