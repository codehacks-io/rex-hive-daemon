package main

import (
	"fmt"
	"os/exec"
)

func main() {
	fmt.Println("Daemon: starting")

	// Prepare command
	app := "./demo-exes/02-sleep-from-cpp.exe"
	arg0 := "-e"

	// Execute command
	cmd := exec.Command(app, arg0)
	stdout, err := cmd.Output()

	// Check for errors
	if err != nil {
		fmt.Println("Daemon: The command failed executing")
		fmt.Println(err.Error())
		return
	}

	// Check for success
	fmt.Println("Daemon: The command succeeded executing")
	fmt.Println(string(stdout))
}
