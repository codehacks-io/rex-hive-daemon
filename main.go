package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	commands := [][]string{
		{"./demo-exes/03-dynamic-sleep-cpp.exe", "1", "-1"},
		{"./demo-exes/03-dynamic-sleep-cpp.exe", "-2", "1", "x"},
	}

	for i, command := range commands {
		wg.Add(1)
		fmt.Println(fmt.Sprintf("[%d] running command '%s' with args %s", i, command[0], command[1:]))
		go run(i, &wg, command[0], command[1:]...)
	}

	// TODO: Use channels to communicate if a goroutine exists, and if so, restart it.
	// TODO: Add a restart policy similar to how docker or k8s or terraform restart pods
	wg.Wait()
}

func run(i int, group *sync.WaitGroup, command string, args ...string) {
	defer group.Done()

	fmt.Println(fmt.Sprintf("\x1b[%dm%s\x1b[0m", 34, fmt.Sprintf("[%d] starting", i)))

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
			colored := fmt.Sprintf("\x1b[%dm%s\x1b[0m", 32, fmt.Sprintf("[%d] STDOUT", i))
			fmt.Println(colored, m)
		}
	}()

	go func() {
		scannerErr := bufio.NewScanner(stderr)
		for scannerErr.Scan() {
			m := scannerErr.Text()
			colored := fmt.Sprintf("\x1b[%dm%s\x1b[0m", 31, fmt.Sprintf("[%d] STDERR", i))
			fmt.Println(colored, m)
		}
	}()

	if err := cmd.Wait(); err != nil {
		fmt.Println(fmt.Sprintf("\x1b[%dm%s\x1b[0m", 31, fmt.Sprintf("[%d] terminated with error: %s", i, err.Error())))
	} else {
		fmt.Println(fmt.Sprintf("\x1b[%dm%s\x1b[0m", 34, fmt.Sprintf("[%d] terminated", i)))
	}
}
