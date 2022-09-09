package main

import (
	"bufio"
	"fmt"
	"os/exec"
)

func main() {
	run("./demo-exes/03-dynamic-sleep-cpp.exe", "1", "-1", "1", "do-fail")
}

func run(command string, args ...string) {
	fmt.Println(fmt.Sprintf("\x1b[%dm%s\x1b[0m", 34, "Daemon: starting"))

	// Prepare command

	// Execute command
	cmd := exec.Command(command, args...)

	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()

	if err = cmd.Start(); err != nil {
		fmt.Println(err)
	}

	// print the output of the subprocess

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m := scanner.Text()
			colored := fmt.Sprintf("\x1b[%dm%s\x1b[0m", 32, "STDOUT")
			fmt.Println(colored, m)
		}
	}()

	go func() {
		scannerErr := bufio.NewScanner(stderr)
		for scannerErr.Scan() {
			m := scannerErr.Text()
			colored := fmt.Sprintf("\x1b[%dm%s\x1b[0m", 31, "STDERR")
			fmt.Println(colored, m)
		}
	}()

	if err := cmd.Wait(); err != nil {
		fmt.Println(fmt.Sprintf("\x1b[%dm%s\x1b[0m", 34, "Daemon: command failed"))
		fmt.Println(fmt.Sprintf("\x1b[%dm%s\x1b[0m", 31, err.Error()))
	} else {
		fmt.Println(fmt.Sprintf("\x1b[%dm%s\x1b[0m", 34, "Daemon: complete"))
	}
}
