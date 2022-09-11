package main

import (
	"bufio"
	"fmt"
	"math"
	"os/exec"
	"sync"
	"time"
)
import p "rex-daemon/rexprint"

/*
Restart policies
k8s: Always, OnFailure, Never
*/

func main() {
	var wg sync.WaitGroup
	colors := p.GetRandomColors()

	commands := [][]string{
		//{"./demo-exes/03-dynamic-sleep-cpp.exe", "1", "3", "9", "3"},
		//{"./demo-exes/03-dynamic-sleep-cpp.exe", "1", "-1", "1", "-1", "1"},
		{"./demo-exes/03-dynamic-sleep-cpp.exe", "f", "f"},
	}

	for i, command := range commands {
		wg.Add(1)
		go runCommandAndKeepAlive(i, &wg, colors, command[0], command[1:]...)
	}

	// TODO: Use channels to communicate if a goroutine exists, and if so, restart it.
	// TODO: Add a restart policy similar to how docker or k8s or terraform restart pods
	wg.Wait()
}

const initialBackoffDelaySeconds = 5

func expBackoffSeconds(attempt int) time.Duration {
	// Cap to 5 minutes (2^7 * 5 = 640)
	if attempt >= 7 {
		return time.Second * 600
	}

	if attempt < 0 {
		return 0
	}

	return time.Second * time.Duration(math.Pow(2, float64(attempt))*initialBackoffDelaySeconds)
}

func runCommandAndKeepAlive(i int, group *sync.WaitGroup, colors []int, command string, args ...string) {
	// Sync with wait group
	defer group.Done()
	attempt := 0
	for {
		runCommand(i, attempt, colors, command, args...)
		attempt++
		backoff := expBackoffSeconds(attempt)
		id := fmt.Sprintf("%d:noPID:%d", i, attempt)
		p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("process terminated, will sleep for %s then will try again", backoff)))
		time.Sleep(backoff)
	}
}

func runCommand(i int, attempt int, colors []int, command string, args ...string) {

	// Execute command
	cmd := exec.Command(command, args...)

	// Get command out pipes
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()

	// Start command
	if err = cmd.Start(); err != nil {
		p.PrintLnColor(fmt.Sprintf("%d", i), colors, i, err.Error())
	}

	// ID format: index:PID:attempt where attempt increases by one each time the command is restarted
	id := fmt.Sprintf("%d:%d:%d", i, cmd.Process.Pid, attempt)

	// TODO: Beware of printing all args, since the user might pass sensitive data as env vars for the game.
	p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("running '%s' with args %s PID %d", command, args, cmd.Process.Pid)))

	// Print realtime stdout from command
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m := scanner.Text()
			p.PrintLnColor(id, colors, i, p.OutColor("STDOUT"), m)
		}
	}()

	// Print realtime stderr from command
	go func() {
		scannerErr := bufio.NewScanner(stderr)
		for scannerErr.Scan() {
			m := scannerErr.Text()
			p.PrintLnColor(id, colors, i, p.ErrColor("STDERR"), m)
		}
	}()

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		p.PrintLnColor(id, colors, i, p.Dim("terminated with error"), err.Error())
	} else {
		p.PrintLnColor(id, colors, i, p.Dim("terminated"))
	}
}
