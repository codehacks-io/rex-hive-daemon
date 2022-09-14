package main

import (
	"bufio"
	"flag"
	"fmt"
	"os/exec"
	"rex-daemon/backoff"
	"sync"
	"time"
)
import p "rex-daemon/rexprint"

// Eg. Run with: `go run .\main.go --file=./demo-specs/test-spec.yml`
func main() {
	// Define cli params
	filePathPtr := flag.String("file", "", "spec file containing args")
	flag.Parse()

	// Read and parse file
	swarmSpec, err := readConf(*filePathPtr)
	if err != nil {
		panic(err)
	}

	runProcessSwarm(swarmSpec)
}

func runProcessSwarm(swarmSpec *ProcessSwarm) {

	if len((*swarmSpec).Spec.ProcessSpecs) < 1 {
		fmt.Println("No process specs to run")
		return
	}

	var usedNumsInSequence = map[int]bool{}
	count := 0

	// Before running any process, validate that we can get all the dynamic args
	for _, s := range swarmSpec.Spec.ProcessSpecs {
		for rep := 0; rep < s.Replicas; rep++ {
			// This line will panic if we cannot get all the dynamic args
			getDynamicArgsOrPanic(s.Cmd[1:], &usedNumsInSequence)
			count++
		}
	}
	usedNumsInSequence = map[int]bool{} // Reset map of used nums after validation
	fmt.Println(fmt.Sprintf("Process specs: %d, total processes: %d", len((*swarmSpec).Spec.ProcessSpecs), count))
	count = 0 // Also reset count
	// End of validations

	swarmChan := make(chan int)

	go func() {
		// Spawn process swarm
		var wg sync.WaitGroup
		colors := p.GetRandomColors()
		for _, s := range swarmSpec.Spec.ProcessSpecs {
			for rep := 0; rep < s.Replicas; rep++ {
				wg.Add(1)
				args := getDynamicArgsOrPanic(s.Cmd[1:], &usedNumsInSequence)
				go runCommandAndKeepAlive(&swarmChan, count, &wg, colors, stringToRestartPolicy[s.Restart], s.Cmd[0], args...)
				count++
			}
		}
		wg.Wait()
		close(swarmChan)
	}()

	for c := range swarmChan {
		fmt.Println("Received PID", c)
	}
}

func runCommandAndKeepAlive(swarmChan *chan int, i int, group *sync.WaitGroup, colors []int, restartPolicy RestartPolicy, command string, args ...string) {
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

		// If the command never stops, the following line will block until command execution terminates
		id, exitCode := runCommand(swarmChan, i, restartPolicy, runCount, colors, command, args...)

		// Get elapsed runtime of command
		elapsed := time.Since(startedAt)

		// Reset backoff
		if elapsed.Seconds() >= backoff.BackoffResetIfUpSeconds {
			backoffCount = 0
		} else {
			backoffCount++
		}

		// If this line is reached, the command exited, either successfully of with an error
		p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("runtime: %s", elapsed)))
		switch restartPolicy {
		case Never:
			{
				p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("won't re-run")))
				return
			}
		case Always:
			{
				delay := backoff.ExpBackoffSeconds(backoffCount)
				p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("will re-run after %s", delay)))
				time.Sleep(delay)
			}
		case OnFailure:
			{
				if exitCode == 0 {
					p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("won't re-run")))
					return
				} else {
					delay := backoff.ExpBackoffSeconds(backoffCount)
					p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("will re-run after %s", delay)))
					time.Sleep(delay)
				}
			}
		}
	}
}

func runCommand(swarmChan *chan int, i int, restart RestartPolicy, attempt int, colors []int, command string, args ...string) (name string, pid int) {

	// Execute command
	cmd := exec.Command(command, args...)

	// Get command out pipes
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()

	cmdSummary := fmt.Sprintf("'%s', args: %s, restart: %s", command, args, restartPolicyToString[restart])

	// Start command
	if err = cmd.Start(); err != nil {
		noPidId := fmt.Sprintf("%d:noPID:%d", i, attempt)
		p.PrintLnColor(noPidId, colors, i, p.ErrColor(fmt.Sprintf("cannot start %s: %s", cmdSummary, err.Error())))
		return noPidId, -1
	}

	// At this point we've got a PID for the process
	*swarmChan <- cmd.Process.Pid

	// ID format: index:PID:attempt where attempt increases by one each time the command is restarted
	id := fmt.Sprintf("%d:%d:%d", i, cmd.Process.Pid, attempt)

	// TODO: Beware of printing all args, since the user might pass sensitive data as env vars for the game.
	p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("running %s, PID %d", cmdSummary, cmd.Process.Pid)))

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
		p.PrintLnColor(id, colors, i, p.ErrColor(fmt.Sprintf("%s. Error-exited with code (%d)", cmdSummary, cmd.ProcessState.ExitCode())), err.Error())
		return id, cmd.ProcessState.ExitCode()
	} else {
		p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("%s. Success-exited with code (%d)", cmdSummary, cmd.ProcessState.ExitCode())))
		return id, 0
	}
}
