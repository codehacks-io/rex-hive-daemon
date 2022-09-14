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

type swarmMessageType int

const (
	processAborted swarmMessageType = iota
	processStarted
	processExited
	processStdOut
	processStdErr
)

type SwarmMessage struct {
	Index    int
	Pid      int
	Attempt  int
	Type     swarmMessageType
	Data     string
	ExitCode int
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

	swarmChan := make(chan SwarmMessage)

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
		fmt.Println(fmt.Sprintf("Received msg %+v", c))
	}
}

func runCommandAndKeepAlive(swarmChan *chan SwarmMessage, i int, group *sync.WaitGroup, colors []int, restartPolicy RestartPolicy, command string, args ...string) {
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

const invalidPid = -1
const noExitCode = -1

func runCommand(swarmChan *chan SwarmMessage, i int, restart RestartPolicy, attempt int, colors []int, command string, args ...string) (name string, pid int) {

	// Execute command
	cmd := exec.Command(command, args...)

	// Get command out pipes
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()

	cmdSummary := fmt.Sprintf("'%s', args: %s, restart: %s", command, args, restartPolicyToString[restart])

	// Start command
	if err = cmd.Start(); err != nil {
		noPidId := fmt.Sprintf("%d:%d:%d", i, invalidPid, attempt)
		p.PrintLnColor(noPidId, colors, i, p.ErrColor(fmt.Sprintf("cannot start %s: %s", cmdSummary, err.Error())))
		*swarmChan <- SwarmMessage{
			Index:    i,
			Pid:      invalidPid,
			Attempt:  attempt,
			Type:     processAborted,
			Data:     err.Error(),
			ExitCode: noExitCode,
		}
		return noPidId, invalidPid
	}

	// At this point we've got a PID for the process

	// ID format: index:PID:attempt where attempt increases by one each time the command is restarted
	id := fmt.Sprintf("%d:%d:%d", i, cmd.Process.Pid, attempt)

	// TODO: Beware of printing all args, since the user might pass sensitive data as env vars for the game.
	p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("running %s, PID %d", cmdSummary, cmd.Process.Pid)))

	*swarmChan <- SwarmMessage{
		Index:    i,
		Pid:      cmd.Process.Pid,
		Attempt:  attempt,
		Type:     processStarted,
		Data:     "",
		ExitCode: noExitCode,
	}

	// Print realtime stdout from command
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m := scanner.Text()
			p.PrintLnColor(id, colors, i, p.OutColor("STDOUT"), m)
			*swarmChan <- SwarmMessage{
				Index:    i,
				Pid:      cmd.Process.Pid,
				Attempt:  attempt,
				Type:     processStdOut,
				Data:     m,
				ExitCode: noExitCode,
			}
		}
	}()

	// Print realtime stderr from command
	go func() {
		scannerErr := bufio.NewScanner(stderr)
		for scannerErr.Scan() {
			m := scannerErr.Text()
			p.PrintLnColor(id, colors, i, p.ErrColor("STDERR"), m)
			*swarmChan <- SwarmMessage{
				Index:    i,
				Pid:      cmd.Process.Pid,
				Attempt:  attempt,
				Type:     processStdErr,
				Data:     m,
				ExitCode: noExitCode,
			}
		}
	}()

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		p.PrintLnColor(id, colors, i, p.ErrColor(fmt.Sprintf("%s. Error-exited with code (%d)", cmdSummary, cmd.ProcessState.ExitCode())), err.Error())
		*swarmChan <- SwarmMessage{
			Index:    i,
			Pid:      cmd.Process.Pid,
			Attempt:  attempt,
			Type:     processExited,
			Data:     err.Error(),
			ExitCode: cmd.ProcessState.ExitCode(),
		}
		return id, cmd.ProcessState.ExitCode()
	} else {
		p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("%s. Success-exited with code (%d)", cmdSummary, cmd.ProcessState.ExitCode())))
		*swarmChan <- SwarmMessage{
			Index:    i,
			Pid:      cmd.Process.Pid,
			Attempt:  attempt,
			Type:     processExited,
			Data:     "",
			ExitCode: 0, // 0 = success
		}
		return id, 0
	}
}
