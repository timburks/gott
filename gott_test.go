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

// delete rows past the end of file and undo
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
