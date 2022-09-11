package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os/exec"
	"sync"
	"time"
)

/*
Restart policies
k8s: Always, OnFailure, Never
*/

func main() {
	var wg sync.WaitGroup
	colors := getRandomColors()

	commands := [][]string{
		{"./demo-exes/03-dynamic-sleep-cpp.exe", "2", "2", "2", "1", "1"},
		//{"./demo-exes/03-dynamic-sleep-cpp.exe", "1", "-1", "1", "-1", "1"},
		{"./demo-exes/03-dynamic-sleep-cpp.exe", "1", "f"},
	}

	for i, command := range commands {
		wg.Add(1)
		go runCommand(i, &wg, colors, command[0], command[1:]...)
	}

	// TODO: Use channels to communicate if a goroutine exists, and if so, restart it.
	// TODO: Add a restart policy similar to how docker or k8s or terraform restart pods
	wg.Wait()
}

func printAllColors() {
	colors := getRandomColors()

	for i, c := range colors {
		fmt.Println(fmt.Sprintf("\x1b[%dm%s\x1b[0m", c, fmt.Sprintf("[%d]:%d", i, c)))
	}
}

func getRandomColors() []int {
	colors := []int{
		31,
		32,
		33,
		34,
		35,
		36,
		//37, Barely visible
		//90, Barely visible
		//91, Very similar to red (31)
		92,
		93,
		94,
		95,
		96,
		97,
	}

	// Shuffle colors array
	rand.Seed(time.Now().UnixNano())
	for i := range colors {
		j := rand.Intn(i + 1)
		colors[i], colors[j] = colors[j], colors[i]
	}
	return colors
}

func printLnColor(id string, colors []int, i int, msg ...any) {
	colorIndex := i % len(colors)
	colored := fmt.Sprintf("\x1b[%dm%s\x1b[0m", colors[colorIndex], fmt.Sprintf("[%s]", id))
	msg = append([]any{colored}, msg...)
	fmt.Println(msg...)
}

func getColored(color int, text string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, text)
}

func dim(text string) string {
	return getColored(37, text)
}

func errColor(text string) string {
	return getColored(41, text)
}

func outColor(text string) string {
	return getColored(42, text)
}

func runCommand(i int, group *sync.WaitGroup, colors []int, command string, args ...string) {
	// Sync with wait group
	defer group.Done()

	// Execute command
	cmd := exec.Command(command, args...)

	// Get command out pipes
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()

	// Start command
	if err = cmd.Start(); err != nil {
		printLnColor(fmt.Sprintf("%d", i), colors, i, err.Error())
	}

	// ID format: index:PID:attempt where attempt increases by one each time the command is restarted
	id := fmt.Sprintf("%d:%d:0", i, cmd.Process.Pid)

	// TODO: Beware of printing all args, since the user might pass sensitive data as env vars for the game.
	printLnColor(id, colors, i, dim(fmt.Sprintf("running '%s' with args %s PID %d", command, args, cmd.Process.Pid)))

	// Print realtime stdout from command
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m := scanner.Text()
			printLnColor(id, colors, i, outColor("STDOUT"), m)
		}
	}()

	// Print realtime stderr from command
	go func() {
		scannerErr := bufio.NewScanner(stderr)
		for scannerErr.Scan() {
			m := scannerErr.Text()
			printLnColor(id, colors, i, errColor("STDERR"), m)
		}
	}()

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		printLnColor(id, colors, i, dim("terminated with error"), err.Error())
	} else {
		printLnColor(id, colors, i, dim("terminated"))
	}
}
