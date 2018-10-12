package main

import (
	"fmt"
	"os"
)

func main() {
	runCmd := NewRunCommand(os.Stdin, os.Stdout, os.Stderr)
	err := runCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}
