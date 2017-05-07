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
package main

import (
	"os"
	"os/exec"
	"testing"
)

const source = "test/gettysburg-address.txt"

func setup(t *testing.T) *Editor {
	editor := NewEditor()
	err := editor.ReadFile(source)
	if err != nil {
		t.Errorf("Read failed: %+v", err)
	}
	return editor
}

func final(t *testing.T, editor *Editor) {
	editor.WriteFile("test-final.txt")
	err := exec.Command("diff", "test-final.txt", source).Run()
	if err != nil {
		t.Errorf("Diff failed: %+v", err)
	} else { // the test succeeded, clean up
		os.Remove("test-final.txt")
	}
}

// read and write a file without changing it
func TestReadWriteInvariance(t *testing.T) {
	editor := setup(t)
	final(t, editor)
}

func TestDeleteRow(t *testing.T) {
	editor := setup(t)
	editor.Cursor = Point{Row: 20, Col: 0}
	editor.Perform(&DeleteRow{}, 20)
	if rowCount := editor.Buffer.RowCount(); rowCount != 20 {
		t.Errorf("Invalid row count after deletion: %d", rowCount)
	}
	editor.PerformUndo()
	final(t, editor)
}

func TestDeleteWord(t *testing.T) {
	editor := setup(t)
	editor.Cursor = Point{Row: 19, Col: 0}
	editor.Perform(&DeleteWord{}, 5)
	expected := "remaining before us--that from these"
	if remainder := editor.Buffer.TextAfter(19, 0); remainder != expected {
		t.Errorf("Unexpected remainder after deletion: '%s'", remainder)
	}
	editor.PerformUndo()
	final(t, editor)
}

func TestDeleteCharacter(t *testing.T) {
	editor := setup(t)
	editor.Cursor = Point{Row: 19, Col: 0}
	editor.Perform(&DeleteCharacter{}, 28)
	expected := "remaining before us--that from these"
	if remainder := editor.Buffer.TextAfter(19, 0); remainder != expected {
		t.Errorf("Unexpected remainder after deletion: '%s'", remainder)
	}
	editor.PerformUndo()
	final(t, editor)
}

func TestInsert(t *testing.T) {
	editor := setup(t)
	editor.Cursor = Point{Row: 1, Col: 0}
	insert := &Insert{Position: InsertAtCursor, Text: "hello, world!"}
	editor.Perform(insert, 1)
	expected := "hello, world!"
	if remainder := editor.Buffer.TextAfter(1, 0); remainder != expected {
		t.Errorf("Unexpected remainder after insertion: '%s'", remainder)
	}
	editor.Cursor = Point{Row: 0, Col: 3}
	insert = &Insert{Position: InsertAfterCursor, Text: "BIG LEAGUE "}
	editor.Perform(insert, 1)
	expected = "THE BIG LEAGUE GETTYSBURG ADDRESS:"
	if remainder := editor.Buffer.TextAfter(0, 0); remainder != expected {
		t.Errorf("Unexpected remainder after insertion: '%s'", remainder)
	}
	editor.Cursor = Point{Row: 3, Col: 3}
	insert = &Insert{Position: InsertAfterEndOfLine, Text: " very"}
	editor.Perform(insert, 1)
	expected = "Four score and seven years ago our fathers brought forth on this very"
	if remainder := editor.Buffer.TextAfter(3, 0); remainder != expected {
		t.Errorf("Unexpected remainder after insertion: '%s'", remainder)
	}
	editor.Cursor = Point{Row: 4, Col: 3}
	insert = &Insert{Position: InsertAtStartOfLine, Text: "nice "}
	editor.Perform(insert, 1)
	expected = "nice continent a new nation, conceived in liberty and dedicated to the"
	if remainder := editor.Buffer.TextAfter(4, 0); remainder != expected {
		t.Errorf("Unexpected remainder after insertion: '%s'", remainder)
	}
	editor.Cursor = Point{Row: 21, Col: 3}
	insert = &Insert{Position: InsertAtNewLineAboveCursor, Text: "most"}
	editor.Perform(insert, 1)
	expected = "most"
	if remainder := editor.Buffer.TextAfter(21, 0); remainder != expected {
		t.Errorf("Unexpected remainder after insertion: '%s'", remainder)
	}
	editor.Cursor = Point{Row: 22, Col: 3}
	insert = &Insert{Position: InsertAtNewLineBelowCursor, Text: "excellent"}
	editor.Perform(insert, 1)
	expected = "excellent"
	if remainder := editor.Buffer.TextAfter(23, 0); remainder != expected {
		t.Errorf("Unexpected remainder after insertion: '%s'", remainder)
	}
	editor.PerformUndo()
	editor.PerformUndo()
	editor.PerformUndo()
	editor.PerformUndo()
	editor.PerformUndo()
	editor.PerformUndo()
	final(t, editor)
}

func TestReverseCase(t *testing.T) {
	editor := setup(t)
	editor.Cursor = Point{Row: 0, Col: 1}
	editor.Perform(&ReverseCaseCharacter{}, 20)
	expected := "The gettysburg addresS:"
	if remainder := editor.Buffer.TextAfter(0, 0); remainder != expected {
		t.Errorf("Unexpected remainder after deletion: '%s'", remainder)
	}
	editor.PerformUndo()
	final(t, editor)
}

func TestReplaceCharacter(t *testing.T) {
	editor := setup(t)
	editor.Cursor = Point{Row: 0, Col: 0}
	editor.Perform(&ReplaceCharacter{Character: 'X'}, 1)
	editor.Cursor = Point{Row: 0, Col: 1}
	editor.Perform(&ReplaceCharacter{Character: 'X'}, 1)
	editor.Cursor = Point{Row: 0, Col: 2}
	editor.Perform(&ReplaceCharacter{Character: 'X'}, 1)
	expected := "XXX GETTYSBURG ADDRESS:"
	if remainder := editor.Buffer.TextAfter(0, 0); remainder != expected {
		t.Errorf("Unexpected remainder after deletion: '%s'", remainder)
	}
	editor.PerformUndo()
	editor.PerformUndo()
	editor.PerformUndo()
	final(t, editor)
}
