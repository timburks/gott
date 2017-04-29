package main

import (
	"io/ioutil"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

// Run the gofmt tool.
func gofmt(filename string, inputBytes []byte) (outputBytes []byte, err error) {
	if false {
		return inputBytes, nil
	}
	cmd := exec.Command(runtime.GOROOT() + "/bin/gofmt")
	input, _ := cmd.StdinPipe()
	output, _ := cmd.StdoutPipe()
	cmderr, _ := cmd.StderrPipe()
	err = cmd.Start()
	if err != nil {
		return
	}
	input.Write(inputBytes)
	input.Close()

	outputBytes, _ = ioutil.ReadAll(output)
	errors, _ := ioutil.ReadAll(cmderr)
	if len(errors) > 0 {
		errors := strings.Replace(string(errors), "<standard input>", filename, -1)
		log.Printf("Syntax errors in code:\n%s", errors)
		return inputBytes, nil
	}

	return
}
