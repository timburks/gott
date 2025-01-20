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

// operation is a utility class that is composed into all operation classes.
// It provides common data storage and services for operations.
type operation struct {
	Cursor     gott.Point
	Multiplier int
	Undo       bool
}

func (op *operation) init(e gott.Editor, multiplier int) {
	if op.Undo {
		e.SetCursor(op.Cursor)
	} else {
		op.Cursor = e.GetCursor()
		if op.Multiplier == 0 {
			op.Multiplier = multiplier
		}
	}
}

func (op *operation) copyForUndo(other *operation) {
	op.Cursor = other.Cursor
	op.Multiplier = other.Multiplier
	op.Undo = true
}
