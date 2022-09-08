package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

func main() {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	fmt.Println("Daemon: starting")

	// Prepare command
	app := "./demo-exes/02-sleep-from-cpp.exe"
	arg0 := "-e"

	// Execute command
	cmd := exec.Command(app, arg0)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	// Check for errors
	if err != nil {
		fmt.Println("Daemon: failed executing")
		fmt.Println(err.Error())
		return
	}

	// Check for success
	fmt.Println("Daemon: succeeded executing")

	fmt.Println("Daemon: output below------------")
	fmt.Println(stdout.String())

	fmt.Println("Daemon: errors below------------")
	fmt.Println(stderr.String())
}
