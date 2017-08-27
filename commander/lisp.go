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

func init() {
	golisp.Global.BindTo(
		golisp.SymbolWithName("TWO"),
		golisp.IntegerWithValue(2))

	golisp.MakePrimitiveFunction("move-down", "0|1",
		func(args *golisp.Data, env *golisp.SymbolTableFrame) (result *golisp.Data, err error) {
			if n, err := optionalFirstArgumentCountValue(args, env); err == nil {
				editor.MoveCursor(gott.MoveDown, n)
			}
			return nil, err
		})

	golisp.MakePrimitiveFunction("move-up", "0|1",
		func(args *golisp.Data, env *golisp.SymbolTableFrame) (result *golisp.Data, err error) {
			if n, err := optionalFirstArgumentCountValue(args, env); err == nil {
				editor.MoveCursor(gott.MoveUp, n)
			}
			return nil, err
		})

	golisp.MakePrimitiveFunction("move-left", "0|1",
		func(args *golisp.Data, env *golisp.SymbolTableFrame) (result *golisp.Data, err error) {
			if n, err := optionalFirstArgumentCountValue(args, env); err == nil {
				editor.MoveCursor(gott.MoveLeft, n)
			}
			return nil, err
		})

	golisp.MakePrimitiveFunction("move-right", "0|1",
		func(args *golisp.Data, env *golisp.SymbolTableFrame) (result *golisp.Data, err error) {
			if n, err := optionalFirstArgumentCountValue(args, env); err == nil {
				editor.MoveCursor(gott.MoveRight, n)
			}
			return nil, err
		})
}

func optionalFirstArgumentCountValue(args *golisp.Data, env *golisp.SymbolTableFrame) (int, error) {
	n := 1
	val := golisp.Car(args)
	if val != nil {
		if !golisp.IntegerP(val) {
			return 0, errors.New("move-down requires an integer argument")
		}
		n = int(golisp.IntegerValue(val))
	}
	return n, nil
}

func (c *Commander) ParseEval(command string) string {
	commander = c
	editor = c.editor
	value, err := golisp.ParseAndEval(command)
	if err != nil {
		return fmt.Sprintf("ERR %+v", err)
	} else {
		return fmt.Sprintf("VALUE %+v", golisp.String(value))
	}
}
