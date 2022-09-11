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

type commandQuery struct {
	command       string
	args          []string
	restartPolicy RestartPolicy
}

func main() {
	var wg sync.WaitGroup
	colors := p.GetRandomColors()

	qs := []commandQuery{
		{"./demo-exes/03-dynamic-sleep-cpp.exe", []string{"2"}, Always},
		//{"./demo-exes/03-dynamic-sleep-cpp.exe", []string{"1", "-1"}, Always},
		//{"./demo-exes/03-dynamic-sleep-cpp.exe", []string{"2"}, OnFailure},
		//{"./demo-exes/03-dynamic-sleep-cpp.exe", []string{"f"}, OnFailure},
		//{"./demo-exes/03-dynamic-sleep-cpp.exe", []string{"15"}, Never},
		//{"./demo-exes/03-dynamic-sleep-cpp.exe", []string{"f"}, Never},
	}

	for i, q := range qs {
		wg.Add(1)
		go runCommandAndKeepAlive(i, &wg, colors, q.restartPolicy, q.command, q.args...)
	}

	wg.Wait()
}

const initialBackoffDelaySeconds = 5
const resetBackoffIfRunForSeconds = 600

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

	runCount := -1
	backoffCount := -1
	for {
		runCount++
		startedAt := time.Now()

		// If the command never stops, the following line will block forever
		id, exitCode := runCommand(i, runCount, colors, command, args...)

		// Get elapsed runtime of command
		elapsed := time.Since(startedAt)

		// Reset backoff
		if elapsed.Seconds() >= resetBackoffIfRunForSeconds {
			backoffCount = 0
		} else {
			backoffCount++
		}

		// If this line is reached, the command exited, either successfully of with an error
		switch restartPolicy {
		case Never:
			{
				p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("terminated with exit code (%d), runtime: %s, restart: Never", exitCode, elapsed)))
				return
			}
		case Always:
			{
				backoff := expBackoffSeconds(backoffCount)
				p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("terminated with exit code (%d), runtime: %s, restart: Always, will sleep for %s and will re-run", exitCode, elapsed, backoff)))
				time.Sleep(backoff)
			}
		case OnFailure:
			{
				if exitCode == 0 {
					p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("terminated with exit code (%d), runtime: %s, restart: OnFailure, will not re-run", exitCode, elapsed)))
					return
				} else {
					backoff := expBackoffSeconds(backoffCount)
					p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("terminated with exit code (%d), runtime: %s, restart: OnFailure, will sleep for %s and will re-run", exitCode, elapsed, backoff)))
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
