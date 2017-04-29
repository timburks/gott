// a good editor

package main

import (
	"log"
	"os"

	"github.com/nsf/termbox-go"
)

const VERSION = "0.1.2"

func main() {
	err := termbox.Init()
	if err != nil {
		log.Printf("ERROR %s", err)
	}
	defer termbox.Close()

	e := NewEditor()

	if len(os.Args) > 1 {
		filename := os.Args[1]
		e.ReadFile(filename)
	}

	// run the editor event loop
	for e.Mode != ModeQuit {
		e.DrawScreen()
		e.ProcessEvent(termbox.PollEvent())
	}
}
