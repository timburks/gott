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

// JoinLine joins the current line with the next one.
type JoinLine struct {
	operation
}

func (op *JoinLine) Perform(e gott.Editor, multiplier int) gott.Operation {
	op.init(e, multiplier)
	cursors := e.JoinRow(op.Multiplier)
	operations := make([]gott.Operation, 0)
	for i := len(cursors) - 1; i >= 0; i-- {
		insert := &Insert{}
		insert.Cursor = cursors[i]
		insert.Multiplier = 1
		insert.Undo = true
		insert.Position = gott.InsertAtCursor
		insert.Text = "\n"
		operations = append(operations, insert)
	}
	inverse := &Sequence{
		Operations: operations,
	}
	inverse.copyForUndo(&op.operation)
	inverse.Multiplier = 1
	return inverse
}
