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

// ReplaceCharacter replaces a character at the current cursor position.
type ReplaceCharacter struct {
	operation
	Character rune
}

func (op *ReplaceCharacter) Perform(e gott.Editor, multiplier int) gott.Operation {
	op.init(e, multiplier)
	old := e.ReplaceCharacterAtCursor(op.Cursor, op.Character)
	inverse := &ReplaceCharacter{}
	inverse.copyForUndo(&op.operation)
	inverse.Character = old
	return inverse
}
