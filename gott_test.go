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
	editor.Cursor = Point{Row: 0, Col: 4}
	insert = &Insert{Position: InsertAtCursor, Text: "BIG LEAGUE "}
	editor.Perform(insert, 1)
	expected = "THE BIG LEAGUE GETTYSBURG ADDRESS:"
	if remainder := editor.Buffer.TextAfter(0, 0); remainder != expected {
		t.Errorf("Unexpected remainder after insertion: '%s'", remainder)
	}
	editor.PerformUndo()
	editor.PerformUndo()
	final(t, editor)
}
