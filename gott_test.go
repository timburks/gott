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

	"github.com/timburks/gott/editor"
	"github.com/timburks/gott/operations"
	gott "github.com/timburks/gott/types"
)

const source = "test/gettysburg-address.txt"

func setup(t *testing.T) gott.Editor {
	e := editor.NewEditor()
	err := e.ReadFile(source)
	if err != nil {
		t.Errorf("Read failed: %+v", err)
	}
	return e
}

func final(t *testing.T, e gott.Editor) {
	e.WriteFile("test-final.txt")
	err := exec.Command("diff", "test-final.txt", source).Run()
	if err != nil {
		t.Errorf("Diff failed: %+v", err)
	} else { // the test succeeded, clean up
		os.Remove("test-final.txt")
	}
}

// read and write a file without changing it
func TestReadWriteInvariance(t *testing.T) {
	e := setup(t)
	final(t, e)
}

func TestDelete3Rows(t *testing.T) {
	e := setup(t)
	originalRowCount := e.GetActiveWindow().GetBuffer().GetRowCount()
	e.SetCursor(gott.Point{Row: 20, Col: 0})
	e.Perform(&operations.DeleteRow{}, 3)
	if rowCount := e.GetActiveWindow().GetBuffer().GetRowCount(); rowCount != originalRowCount-3 {
		t.Errorf("Invalid row count after deletion: %d", rowCount)
	}
	e.PerformUndo()
	final(t, e)
}

func TestDelete20Rows(t *testing.T) {
	e := setup(t)
	e.SetCursor(gott.Point{Row: 20, Col: 0})
	e.Perform(&operations.DeleteRow{}, 20)
	if rowCount := e.GetActiveWindow().GetBuffer().GetRowCount(); rowCount != 20 {
		t.Errorf("Invalid row count after deletion: %d", rowCount)
	}
	e.PerformUndo()
	final(t, e)
}

func TestDeleteWord(t *testing.T) {
	e := setup(t)
	e.SetCursor(gott.Point{Row: 19, Col: 0})
	e.Perform(&operations.DeleteWord{}, 5)
	expected := "remaining before us--that from these"
	if remainder := e.GetActiveWindow().GetBuffer().TextAfter(19, 0); remainder != expected {
		t.Errorf("Unexpected remainder after deletion: '%s'", remainder)
	}
	e.PerformUndo()
	final(t, e)
}

func TestDeleteCharacter(t *testing.T) {
	e := setup(t)
	e.SetCursor(gott.Point{Row: 19, Col: 0})
	e.Perform(&operations.DeleteCharacter{}, 28)
	expected := "remaining before us--that from these"
	if remainder := e.GetActiveWindow().GetBuffer().TextAfter(19, 0); remainder != expected {
		t.Errorf("Unexpected remainder after deletion: '%s'", remainder)
	}
	e.PerformUndo()
	final(t, e)
}

func TestInsert(t *testing.T) {
	e := setup(t)
	e.SetCursor(gott.Point{Row: 1, Col: 0})
	insert := &operations.Insert{Position: gott.InsertAtCursor, Text: "hello, world!"}
	e.Perform(insert, 1)
	expected := "hello, world!"
	if remainder := e.GetActiveWindow().GetBuffer().TextAfter(1, 0); remainder != expected {
		t.Errorf("Unexpected remainder after insertion: '%s'", remainder)
	}
	e.SetCursor(gott.Point{Row: 0, Col: 3})
	insert = &operations.Insert{Position: gott.InsertAfterCursor, Text: "BIG LEAGUE "}
	e.Perform(insert, 1)
	expected = "THE BIG LEAGUE GETTYSBURG ADDRESS:"
	if remainder := e.GetActiveWindow().GetBuffer().TextAfter(0, 0); remainder != expected {
		t.Errorf("Unexpected remainder after insertion: '%s'", remainder)
	}
	e.SetCursor(gott.Point{Row: 3, Col: 3})
	insert = &operations.Insert{Position: gott.InsertAfterEndOfLine, Text: " very"}
	e.Perform(insert, 1)
	expected = "Four score and seven years ago our fathers brought forth on this very"
	if remainder := e.GetActiveWindow().GetBuffer().TextAfter(3, 0); remainder != expected {
		t.Errorf("Unexpected remainder after insertion: '%s'", remainder)
	}
	e.SetCursor(gott.Point{Row: 4, Col: 3})
	insert = &operations.Insert{Position: gott.InsertAtStartOfLine, Text: "nice "}
	e.Perform(insert, 1)
	expected = "nice continent a new nation, conceived in liberty and dedicated to the"
	if remainder := e.GetActiveWindow().GetBuffer().TextAfter(4, 0); remainder != expected {
		t.Errorf("Unexpected remainder after insertion: '%s'", remainder)
	}
	e.SetCursor(gott.Point{Row: 21, Col: 3})
	insert = &operations.Insert{Position: gott.InsertAtNewLineAboveCursor, Text: "most"}
	e.Perform(insert, 1)
	expected = "most"
	if remainder := e.GetActiveWindow().GetBuffer().TextAfter(21, 0); remainder != expected {
		t.Errorf("Unexpected remainder after insertion: '%s'", remainder)
	}
	e.SetCursor(gott.Point{Row: 22, Col: 3})
	insert = &operations.Insert{Position: gott.InsertAtNewLineBelowCursor, Text: "excellent"}
	e.Perform(insert, 1)
	expected = "excellent"
	if remainder := e.GetActiveWindow().GetBuffer().TextAfter(23, 0); remainder != expected {
		t.Errorf("Unexpected remainder after insertion: '%s'", remainder)
	}
	e.PerformUndo()
	e.PerformUndo()
	e.PerformUndo()
	e.PerformUndo()
	e.PerformUndo()
	e.PerformUndo()
	final(t, e)
}

func TestReverseCase(t *testing.T) {
	e := setup(t)
	e.SetCursor(gott.Point{Row: 0, Col: 1})
	e.Perform(&operations.ReverseCaseCharacter{}, 20)
	expected := "The gettysburg addresS:"
	if remainder := e.GetActiveWindow().GetBuffer().TextAfter(0, 0); remainder != expected {
		t.Errorf("Unexpected remainder after deletion: '%s'", remainder)
	}
	e.PerformUndo()
	final(t, e)
}

func TestReplaceCharacter(t *testing.T) {
	e := setup(t)
	e.SetCursor(gott.Point{Row: 0, Col: 0})
	e.Perform(&operations.ReplaceCharacter{Character: 'X'}, 1)
	e.SetCursor(gott.Point{Row: 0, Col: 1})
	e.Perform(&operations.ReplaceCharacter{Character: 'X'}, 1)
	e.SetCursor(gott.Point{Row: 0, Col: 2})
	e.Perform(&operations.ReplaceCharacter{Character: 'X'}, 1)
	expected := "XXX GETTYSBURG ADDRESS:"
	if remainder := e.GetActiveWindow().GetBuffer().TextAfter(0, 0); remainder != expected {
		t.Errorf("Unexpected remainder after deletion: '%s'", remainder)
	}
	e.PerformUndo()
	e.PerformUndo()
	e.PerformUndo()
	final(t, e)
}

func TestCopyPaste(t *testing.T) {
	e := setup(t)
	e.SetCursor(gott.Point{Row: 3, Col: 3})
	// copy three rows
	e.YankRow(3)
	e.SetCursor(gott.Point{Row: 2, Col: 0})
	// paste them three times
	e.Perform(&operations.Paste{}, 3)
	// verify that we added 9 rows
	if rowCount := e.GetActiveWindow().GetBuffer().GetRowCount(); rowCount != (38 + 9) {
		t.Errorf("Invalid row count after paste: %d", rowCount)
	}
	// sample the expected text
	expected := "Four score and seven years ago our fathers brought forth on this"
	if sample := e.GetActiveWindow().GetBuffer().TextAfter(3, 0); sample != expected {
		t.Errorf("Unexpected sample after paste: '%s'", sample)
	}
	if sample := e.GetActiveWindow().GetBuffer().TextAfter(6, 0); sample != expected {
		t.Errorf("Unexpected sample after paste: '%s'", sample)
	}
	if sample := e.GetActiveWindow().GetBuffer().TextAfter(9, 0); sample != expected {
		t.Errorf("Unexpected sample after paste: '%s'", sample)
	}
	if sample := e.GetActiveWindow().GetBuffer().TextAfter(12, 0); sample != expected {
		t.Errorf("Unexpected sample after paste: '%s'", sample)
	}
	e.PerformUndo()
	final(t, e)
}

func TestJoinRow(t *testing.T) {
	e := setup(t)
	e.SetCursor(gott.Point{Row: 0, Col: 3})
	// join three lines
	e.Perform(&operations.JoinLine{}, 3)
	// sample the expected text
	expected := "THE GETTYSBURG ADDRESS:Four score and seven years ago our fathers brought forth on this"
	if sample := e.GetActiveWindow().GetBuffer().TextAfter(0, 0); sample != expected {
		t.Errorf("Unexpected sample after paste: '%s'", sample)
	}
	e.PerformUndo()
	final(t, e)
}

func TestChangeWord(t *testing.T) {
	e := setup(t)
	e.SetCursor(gott.Point{Row: 3, Col: 0})
	// change four words
	e.Perform(&operations.ChangeWord{Text: "47 "}, 4)
	// sample the expected text
	expected := "47 years ago our fathers brought forth on this"
	if sample := e.GetActiveWindow().GetBuffer().TextAfter(3, 0); sample != expected {
		t.Errorf("Unexpected sample after paste: '%s'", sample)
	}
	e.PerformUndo()
	final(t, e)
}
