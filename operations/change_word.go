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
	inverse := &Insert{
		Position: gott.InsertAtCursor,
		Text:     string(deletedText),
	}
	inverse.copyForUndo(&op.operation)
	return inverse
}

func (op *ChangeWord) AddCharacter(c rune) {

}
func (op *ChangeWord) DeleteCharacter() {

}
func (op *ChangeWord) Close() {

}
func (op *ChangeWord) Length() int {
	return 0
}
