package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os/exec"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	colors := getRandomColors()

	commands := [][]string{
		{"./demo-exes/03-dynamic-sleep-cpp.exe", "1", "1"},
		{"./demo-exes/03-dynamic-sleep-cpp.exe", "1", "-1"},
		{"./demo-exes/03-dynamic-sleep-cpp.exe", "-1", "-1", "f"},
	}

	for i, command := range commands {
		// TODO: Beware of printing all args, since the user might pass sensitive data as env vars for the game.
		printLnColor(colors, i, dim(fmt.Sprintf("running command '%s' with args %s", command[0], command[1:])))
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

func printLnColor(colors []int, i int, msg ...any) {
	colorIndex := i % len(colors)
	colored := fmt.Sprintf("\x1b[%dm%s\x1b[0m", colors[colorIndex], fmt.Sprintf("[%02d]", i))
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
	group.Add(1)
	defer group.Done()

	printLnColor(colors, i, dim("starting"))

	// Prepare command

	// Execute command
	cmd := exec.Command(command, args...)

	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()

	if err = cmd.Start(); err != nil {
		printLnColor(colors, i, err.Error())
	}

	// print the output of the subprocess

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m := scanner.Text()
			printLnColor(colors, i, outColor("STDOUT"), m)
		}
	}()

	go func() {
		scannerErr := bufio.NewScanner(stderr)
		for scannerErr.Scan() {
			m := scannerErr.Text()
			printLnColor(colors, i, errColor("STDERR"), m)
		}
	}()

	if err := cmd.Wait(); err != nil {
		printLnColor(colors, i, dim("terminated with error"), err.Error())
	} else {
		printLnColor(colors, i, dim("terminated"))
	}
}
