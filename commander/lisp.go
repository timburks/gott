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
	"io/ioutil"
	"os"

	"github.com/steelseries/golisp"
	"github.com/timburks/gott/operations"
	gott "github.com/timburks/gott/types"
)

// file-global pointers to editor objects
var commander *Commander
var editor gott.Editor

func argumentCountValue(name string, args *golisp.Data, env *golisp.SymbolTableFrame) (int, error) {
	n := 1
	val := golisp.Car(args)
	if val != nil {
		if !golisp.IntegerP(val) {
			return 0, errors.New(fmt.Sprintf("%s requires an integer argument", name))
		}
		n = int(golisp.IntegerValue(val))
	} else {
		n = commander.Multiplier()
	}
	return n, nil
}

func makePrimitiveFunctionWithMultiplier(name string, action func(multiplier int)) {
	golisp.MakePrimitiveFunction(name, "0|1",
		func(args *golisp.Data, env *golisp.SymbolTableFrame) (result *golisp.Data, err error) {
			if n, err := argumentCountValue(name, args, env); err == nil {
				action(n)
			}
			return nil, err
		})
}

func argumentStringValue(name string, args *golisp.Data, env *golisp.SymbolTableFrame) (string, error) {
	n := ""
	val := golisp.Car(args)
	if val != nil {
		if !golisp.StringP(val) {
			return "", errors.New(fmt.Sprintf("%s requires a string argument", name))
		}
		n = golisp.StringValue(val)
	}
	return n, nil
}

func makePrimitiveFunctionWithString(name string, action func(s string)) {
	golisp.MakePrimitiveFunction(name, "1",
		func(args *golisp.Data, env *golisp.SymbolTableFrame) (result *golisp.Data, err error) {
			if n, err := argumentStringValue(name, args, env); err == nil {
				action(n)
			}
			return nil, err
		})
}

func init() {
	golisp.Global.BindTo(
		golisp.SymbolWithName("TWO"),
		golisp.IntegerWithValue(2))

	makePrimitiveFunctionWithMultiplier("down", func(m int) {
		editor.MoveCursor(gott.MoveDown, m)
	})

	makePrimitiveFunctionWithMultiplier("up", func(m int) {
		editor.MoveCursor(gott.MoveUp, m)
	})

	makePrimitiveFunctionWithMultiplier("left", func(m int) {
		editor.MoveCursor(gott.MoveLeft, m)
	})

	makePrimitiveFunctionWithMultiplier("right", func(m int) {
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

	makePrimitiveFunctionWithMultiplier("beginning-of-line", func(m int) {
		editor.MoveToBeginningOfLine()
	})

	makePrimitiveFunctionWithMultiplier("end-of-line", func(m int) {
		editor.MoveToEndOfLine()
	})

	makePrimitiveFunctionWithMultiplier("next-word", func(m int) {
		editor.MoveCursorToNextWord(m)
	})

	makePrimitiveFunctionWithMultiplier("previous-word", func(m int) {
		editor.MoveCursorToPreviousWord(m)
	})

	makePrimitiveFunctionWithMultiplier("insert-at-cursor", func(m int) {
		editor.Perform(&operations.Insert{Position: gott.InsertAtCursor, Commander: commander}, m)
	})

	makePrimitiveFunctionWithMultiplier("insert-after-cursor", func(m int) {
		editor.Perform(&operations.Insert{Position: gott.InsertAfterCursor, Commander: commander}, m)
	})

	makePrimitiveFunctionWithMultiplier("insert-at-start-of-line", func(m int) {
		editor.Perform(&operations.Insert{Position: gott.InsertAtStartOfLine, Commander: commander}, m)
	})

	makePrimitiveFunctionWithMultiplier("insert-after-end-of-line", func(m int) {
		editor.Perform(&operations.Insert{Position: gott.InsertAfterEndOfLine, Commander: commander}, m)
	})

	makePrimitiveFunctionWithMultiplier("insert-at-new-line-below-cursor", func(m int) {
		editor.Perform(&operations.Insert{Position: gott.InsertAtNewLineBelowCursor, Commander: commander}, m)
	})

	makePrimitiveFunctionWithMultiplier("insert-at-new-line-above-cursor", func(m int) {
		editor.Perform(&operations.Insert{Position: gott.InsertAtNewLineAboveCursor, Commander: commander}, m)
	})

	makePrimitiveFunctionWithMultiplier("delete-character", func(m int) {
		editor.Perform(&operations.DeleteCharacter{}, m)
	})

	makePrimitiveFunctionWithMultiplier("join-line", func(m int) {
		editor.Perform(&operations.JoinLine{}, m)
	})

	makePrimitiveFunctionWithMultiplier("paste", func(m int) {
		editor.Perform(&operations.Paste{}, m)
	})

	makePrimitiveFunctionWithMultiplier("reverse-case-character", func(m int) {
		editor.Perform(&operations.ReverseCaseCharacter{}, m)
	})

	makePrimitiveFunctionWithMultiplier("perform-undo", func(m int) {
		editor.PerformUndo()
	})

	makePrimitiveFunctionWithMultiplier("repeat-last-command", func(m int) {
		editor.Repeat()
	})

	makePrimitiveFunctionWithMultiplier("change-word", func(m int) {
		editor.Perform(&operations.ChangeWord{Commander: commander}, m)
	})

	makePrimitiveFunctionWithMultiplier("delete-row", func(m int) {
		editor.Perform(&operations.DeleteRow{}, m)
	})

	makePrimitiveFunctionWithMultiplier("delete-word", func(m int) {
		editor.Perform(&operations.DeleteWord{}, m)
	})

	makePrimitiveFunctionWithMultiplier("yank-row", func(m int) {
		editor.YankRow(m)
	})

	makePrimitiveFunctionWithString("print", func(s string) {
		if commander.batch {
			// if we are running in batch (eval) mode, write to output
			os.Stdout.Write([]byte(s + "\n"))
		} else {
			// if we are running in the editor, write to buffer 0
			editor.SelectBuffer(0)
			editor.GetBuffer().AppendBytes([]byte(s))
		}
	})
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

func (c *Commander) ParseEvalFile(filename string) string {
	bytes, err := ioutil.ReadFile(filename)
	if err == nil {
		contents := string(bytes)
		c.batch = true
		return c.ParseEval(contents)
	} else {
		return err.Error()
	}
}
