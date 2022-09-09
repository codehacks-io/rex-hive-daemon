package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go run(&wg, "./demo-exes/03-dynamic-sleep-cpp.exe", "1", "1")
	}

	wg.Wait()
}

func run(group *sync.WaitGroup, command string, args ...string) {
	defer group.Done()

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
