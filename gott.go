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
	"log"
	"os"

	"github.com/nsf/termbox-go"
)

const VERSION = "0.1.2"

func main() {
	// open a log file
	f, err := os.OpenFile(os.Getenv("HOME")+"/.gott.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Output(1, err.Error())
		return
	}
	log.SetOutput(f)
	defer f.Close()

	// open the terminal.
	err = termbox.Init()
	if err != nil {
		log.Output(1, err.Error())
		return
	}
	defer termbox.Close()

	// create our editor.
	e := NewEditor()

	c := Commander{Editor: e}

	// if a file was specified on the command-line, read it.
	if len(os.Args) > 1 {
		filename := os.Args[1]
		err = e.ReadFile(filename)
		if err != nil {
			log.Output(1, err.Error())
		}
	}

	// run the editor event loop.
	for e.Mode != ModeQuit {
		e.Render()
		err = c.ProcessEvent(termbox.PollEvent())
		if err != nil {
			log.Output(1, err.Error())
		}
	}
}
