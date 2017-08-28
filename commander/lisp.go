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
package commander

import (
	"errors"
	"fmt"

	"github.com/steelseries/golisp"
	gott "github.com/timburks/gott/types"
)

// file-global pointers to editor objects
var commander *Commander
var editor gott.Editor

func makePrimitiveFunctionWithMultiplier(name string, action func(multiplier int)) {
	golisp.MakePrimitiveFunction(name, "0|1",
		func(args *golisp.Data, env *golisp.SymbolTableFrame) (result *golisp.Data, err error) {
			if n, err := argumentCountValue(name, args, env); err == nil {
				action(n)
			}
			return nil, err
		})
}

func init() {
	golisp.Global.BindTo(
		golisp.SymbolWithName("TWO"),
		golisp.IntegerWithValue(2))

	makePrimitiveFunctionWithMultiplier("move-down", func(m int) {
		editor.MoveCursor(gott.MoveDown, m)
	})

	makePrimitiveFunctionWithMultiplier("move-up", func(m int) {
		editor.MoveCursor(gott.MoveUp, m)
	})

	makePrimitiveFunctionWithMultiplier("move-left", func(m int) {
		editor.MoveCursor(gott.MoveLeft, m)
	})

	makePrimitiveFunctionWithMultiplier("move-right", func(m int) {
		editor.MoveCursor(gott.MoveRight, m)
	})

	makePrimitiveFunctionWithMultiplier("page-down", func(m int) {
		editor.PageDown(m)
	})

	makePrimitiveFunctionWithMultiplier("page-up", func(m int) {
		editor.PageUp(m)
	})

	makePrimitiveFunctionWithMultiplier("half-page-down", func(m int) {
		editor.HalfPageDown(m)
	})

	makePrimitiveFunctionWithMultiplier("half-page-up", func(m int) {
		editor.HalfPageUp(m)
	})
}

func argumentCountValue(name string, args *golisp.Data, env *golisp.SymbolTableFrame) (int, error) {
	n := 1
	val := golisp.Car(args)
	if val != nil {
		if !golisp.IntegerP(val) {
			return 0, errors.New(fmt.Sprintf("%s requires an integer argument", name))
		}
		n = int(golisp.IntegerValue(val))
	}
	return n, nil
}

func (c *Commander) ParseEval(command string) string {
	commander = c
	editor = c.editor
	value, err := golisp.ParseAndEvalAll(command)
	if err != nil {
		return fmt.Sprintf("ERR %+v", err)
	} else {
		return golisp.String(value)
	}
}
