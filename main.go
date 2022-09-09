package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"math/rand"
	"os/exec"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	colors := getRandomColors()

	commands := [][]string{
		{"./demo-exes/03-dynamic-sleep-cpp.exe", "1", "1", "1", "1"},
		{"./demo-exes/03-dynamic-sleep-cpp.exe", "1", "1", "1", "1"},
	}

	for i, command := range commands {
		wg.Add(1)
		fmt.Println(fmt.Sprintf("[%d] running command '%s' with args %s", i, command[0], command[1:]))
		go run(i, &wg, colors, command[0], command[1:]...)
	}

	// TODO: Use channels to communicate if a goroutine exists, and if so, restart it.
	// TODO: Add a restart policy similar to how docker or k8s or terraform restart pods
	wg.Wait()
}

func getRandomColors() []*color.Color {
	colors := []*color.Color{
		color.New(color.FgBlack),
		color.New(color.FgRed),
		color.New(color.FgGreen),
		color.New(color.FgYellow),
		color.New(color.FgBlue),
		color.New(color.FgMagenta),
		color.New(color.FgCyan),
		color.New(color.FgWhite),
		color.New(color.FgHiBlack),
		color.New(color.FgHiRed),
		color.New(color.FgHiGreen),
		color.New(color.FgHiYellow),
		color.New(color.FgHiBlue),
		color.New(color.FgHiMagenta),
		color.New(color.FgHiCyan),
		color.New(color.FgHiWhite),
	}

	// Shuffle colors array
	rand.Seed(time.Now().UnixNano())
	for i := range colors {
		j := rand.Intn(i + 1)
		colors[i], colors[j] = colors[j], colors[i]
	}
	return colors
}

func printLnColor(colors []*color.Color, i int, msg ...any) {
	colorIndex := i % len(colors)
	_, _ = colors[colorIndex].Print(fmt.Sprintf("[%d] ", i))
	fmt.Println(msg...)
}

func run(i int, group *sync.WaitGroup, colors []*color.Color, command string, args ...string) {
	defer group.Done()

	printLnColor(colors, i, "starting")

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
			printLnColor(colors, i, "STDOUT", m)
		}
	}()

	go func() {
		scannerErr := bufio.NewScanner(stderr)
		for scannerErr.Scan() {
			m := scannerErr.Text()
			printLnColor(colors, i, "STDERR", m)
		}
	}()

	if err := cmd.Wait(); err != nil {
		printLnColor(colors, i, "terminated with error", err.Error())
	} else {
		printLnColor(colors, i, "terminated")
	}
}
