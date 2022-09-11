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

type RestartPolicy int

const (
	Always RestartPolicy = iota
	OnFailure
	Never
)

func main() {
	var wg sync.WaitGroup
	colors := p.GetRandomColors()

	commands := [][]string{
		//{"./demo-exes/03-dynamic-sleep-cpp.exe", "1", "3", "9", "3"},
		//{"./demo-exes/03-dynamic-sleep-cpp.exe", "1", "-1", "1", "-1", "1"},
		//{"./demo-exes/03-dynamic-sleep-cpp.exxe", "f", "f"},
		{"./demo-exes/03-dynamic-sleep-cpp.exe", "1"},
	}

	for i, command := range commands {
		wg.Add(1)
		go runCommandAndKeepAlive(i, &wg, colors, OnFailure, command[0], command[1:]...)
	}

	// TODO: After x minutes running successfully, reset falloff
	wg.Wait()
}

const initialBackoffDelaySeconds = 5

func expBackoffSeconds(attempt int) time.Duration {
	// Cap to 5 minutes
	if attempt >= 6 {
		return time.Second * 300
	}

	if attempt < 0 {
		return 0
	}

	return time.Second * time.Duration(math.Pow(2, float64(attempt))*initialBackoffDelaySeconds)
}

func runCommandAndKeepAlive(i int, group *sync.WaitGroup, colors []int, restartPolicy RestartPolicy, command string, args ...string) {
	// Sync with wait group
	defer group.Done()

	// Validate retry policy
	if restartPolicy != Always && restartPolicy != OnFailure && restartPolicy != Never {
		panic(fmt.Sprintf("Invalid retry policy %d", restartPolicy))
		return
	}

	attempt := -1
	for {
		attempt++

		// If the command never stops, the following line will block forever
		id, exitCode := runCommand(i, attempt, colors, command, args...)

		// If this line is reached, the command exited, either successfully of with an error
		switch restartPolicy {
		case Never:
			{
				p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("terminated with exit code (%d), restart: Never", exitCode)))
				return
			}
		case Always:
			{
				backoff := expBackoffSeconds(attempt)
				p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("terminated with exit code (%d), restart: Always, will sleep for %s and will re-run", exitCode, backoff)))
				time.Sleep(backoff)
			}
		case OnFailure:
			{
				if exitCode == 0 {
					p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("terminated with exit code (%d), restart: OnFailure, will not re-run", exitCode)))
					return
				} else {
					backoff := expBackoffSeconds(attempt)
					p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("terminated with exit code (%d), restart: OnFailure, will sleep for %s and will re-run", exitCode, backoff)))
					time.Sleep(backoff)
				}
			}
		}
	}
}

func runCommand(i int, attempt int, colors []int, command string, args ...string) (name string, pid int) {

	// Execute command
	cmd := exec.Command(command, args...)

	// Get command out pipes
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()

	cmdSummary := fmt.Sprintf("'%s' with args %s", command, args)

	// Start command
	if err = cmd.Start(); err != nil {
		noPidId := fmt.Sprintf("%d:noPID:%d", i, attempt)
		p.PrintLnColor(noPidId, colors, i, p.ErrColor(fmt.Sprintf("cannot start %s: %s", cmdSummary, err.Error())))
		return noPidId, -1
	}

	// ID format: index:PID:attempt where attempt increases by one each time the command is restarted
	id := fmt.Sprintf("%d:%d:%d", i, cmd.Process.Pid, attempt)

	// TODO: Beware of printing all args, since the user might pass sensitive data as env vars for the game.
	p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("running %s PID %d", cmdSummary, cmd.Process.Pid)))

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
		p.PrintLnColor(id, colors, i, p.ErrColor(fmt.Sprintf("%s exited with error", cmdSummary)), err.Error())
		return id, cmd.ProcessState.ExitCode()
	} else {
		p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("%s exited with success code", cmdSummary)))
		return id, 0
	}
}
