package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"rex-hive-daemon/backoff"
	"rex-hive-daemon/hive_message"
	"rex-hive-daemon/hive_spec"
	"rex-hive-daemon/message_handler"
	"rex-hive-daemon/slice_tools"
	"sync"
	"syscall"
	"time"
)
import p "rex-hive-daemon/rexprint"

var (
	tearingDown     = false
	runningCommands []*exec.Cmd
	// Locks reads and writes to tearingDown and runningCommands.
	killingLock sync.Mutex
)

func killAllProcesses() {
	killingLock.Lock()
	tearingDown = true
	fmt.Println("Trying to kill", len(runningCommands), "processes")
	for _, cmd := range runningCommands {
		if cmd == nil || cmd.Process == nil {
			fmt.Println("Command", cmd, "has a nil process, cannot be stopped or it's already stopped")
			continue
		}
		if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
			fmt.Println("Could not send SIGINT signal to process. Error:", err, "PID:", cmd.Process.Pid, cmd, "will try to kill it...")
			if err = cmd.Process.Kill(); err != nil {
				fmt.Println("Could not kill process. Error:", err, "PID:", cmd.Process.Pid, cmd)
				continue
			}
		}
		fmt.Println("process successfully killed", cmd.Process.Pid, cmd)
	}
	killingLock.Unlock()
}

func listenForTermination() {
	sigsChan := make(chan os.Signal, 1)
	signal.Notify(sigsChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigsChan
		fmt.Println("Just received a signal:", sig, "will wait to flush before exiting")
		killAllProcesses()
	}()
}

// Eg. Run with: `go run .\main.go --file=./demo-specs/test-spec.yml`
func main() {
	listenForTermination()

	// Define cli params
	filePathPtr := flag.String("file", "", "spec file containing args")
	flag.Parse()

	// Read and parse file
	hiveSpec, err := hive_spec.FromFile(*filePathPtr)
	if err != nil {
		panic(err)
	}

	go message_handler.Run(hiveSpec)
	runHiveSpec(hiveSpec)

	// Wait for messages to be stored in DB (flushing)
	fmt.Println(p.Dim("HiveRun finished, waiting to flush"))
	flushChan := make(chan bool)
	message_handler.Flush(&flushChan)
	<-flushChan
	close(flushChan)
}

func runHiveSpec(hiveSpec *hive_spec.HiveSpec) {

	if len((*hiveSpec).Spec.Processes) < 1 {
		fmt.Println("No process specs to run")
		return
	}

	var usedNumsInSequence = map[int]bool{}
	count := 0

	// Before running any process, validate that we can get all the dynamic args
	for _, s := range hiveSpec.Spec.Processes {
		for rep := 0; rep < s.Replicas; rep++ {
			// This line will panic if we cannot get all the dynamic args
			args := s.Cmd[1:]
			getDynamicArgsOrPanic(&args, &usedNumsInSequence)
			count++
		}
	}
	usedNumsInSequence = map[int]bool{} // Reset map of used nums after validation
	fmt.Println(fmt.Sprintf("Process specs: %d, total processes: %d", len((*hiveSpec).Spec.Processes), count))
	count = 0 // Also reset count
	// End of validations

	hiveChan := make(chan *hive_message.HiveMessage)

	go func() {
		// Spawn processes in spec
		var wg sync.WaitGroup
		colors := p.GetRandomColors()
		for _, processSpec := range hiveSpec.Spec.Processes {
			for rep := 0; rep < processSpec.Replicas; rep++ {
				wg.Add(1)
				args := processSpec.Cmd[1:]
				replacedArgs := getDynamicArgsOrPanic(&args, &usedNumsInSequence)
				go runCommandAndKeepAlive(&hiveChan, count, &wg, colors, processSpec, replacedArgs...)
				count++
			}
		}
		wg.Wait()
		close(hiveChan)
	}()

	for c := range hiveChan {
		message_handler.OnHiveMessage(c)
	}
}

func runCommandAndKeepAlive(hiveChan *chan *hive_message.HiveMessage, i int, group *sync.WaitGroup, colors []int, processSpec *hive_spec.ProcessSpec, args ...string) {
	restartPolicy := stringToRestartPolicy[processSpec.Restart]

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
		id, exitCode := runCommand(hiveChan, i, runCount, colors, processSpec, args...)

		killingLock.Lock()
		if tearingDown {
			p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("tearing down, wont re-run ANY process")))
			killingLock.Unlock()
			return
		}
		killingLock.Unlock()

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

func runCommand(hiveChan *chan *hive_message.HiveMessage, i int, attempt int, colors []int, processSpec *hive_spec.ProcessSpec, args ...string) (name string, pid int) {
	preSpawnId := fmt.Sprintf("%d:%d:%d", i, invalidPid, attempt)

	killingLock.Lock()
	if tearingDown {
		p.PrintLnColor(preSpawnId, colors, i, p.Dim(fmt.Sprintf("tearing down, skipping process")))
		killingLock.Unlock()
		return preSpawnId, invalidPid
	}

	// Execute command
	cmd := exec.Command(processSpec.Cmd[0], args...)
	runningCommands = append(runningCommands, cmd)
	killingLock.Unlock()

	// Remove the cmd from the array of running commands.
	defer func() {
		killingLock.Lock()
		if !tearingDown {
			// Only remove from the running commands if not tearing down.
			// When tearing down, we are actively killing the processes one by one and wee need the cmd array for that.
			runningCommands = *slice_tools.RemoveFirst(&runningCommands, func(x *exec.Cmd) bool { return x == cmd })
		}
		killingLock.Unlock()
	}()

	// Get command out pipes
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()

	cmdSummary := fmt.Sprintf("'%s', args: %s, restart: %s", processSpec.Cmd[0], args, processSpec.Restart)

	func() {
		if len(processSpec.Env) <= 0 {
			p.PrintLnColor(preSpawnId, colors, i, p.Dim("process spec has no env vars"))
		}

		if processSpec.ForwardOsEnv {
			cmd.Env = os.Environ()
			fmt.Println(cmd.Env)
		}

		for _, envEntry := range processSpec.Env {
			p.PrintLnColor(preSpawnId, colors, i, p.Dim(fmt.Sprintf("setting env %s=%s", envEntry.Name, envEntry.Value)))
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", envEntry.Name, envEntry.Value))
		}
	}()

	// Start command
	if err = cmd.Start(); err != nil {
		p.PrintLnColor(preSpawnId, colors, i, p.ErrColor(fmt.Sprintf("cannot start %s: %s", cmdSummary, err.Error())))
		*hiveChan <- &hive_message.HiveMessage{
			Index:    i,
			Pid:      invalidPid,
			Attempt:  attempt,
			Type:     hive_message.ProcessAborted,
			Data:     err.Error(),
			ExitCode: noExitCode,
		}
		return preSpawnId, invalidPid
	}

	// At this point we've got a PID for the process

	// ID format: index:PID:attempt where attempt increases by one each time the command is restarted
	id := fmt.Sprintf("%d:%d:%d", i, cmd.Process.Pid, attempt)

	// TODO: Beware of printing all args, since the user might pass sensitive data as env vars for the game.
	p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("running %s, PID %d", cmdSummary, cmd.Process.Pid)))

	*hiveChan <- &hive_message.HiveMessage{
		Index:    i,
		Pid:      cmd.Process.Pid,
		Attempt:  attempt,
		Type:     hive_message.ProcessStarted,
		Data:     "",
		ExitCode: noExitCode,
	}

	// Print realtime stdout from command
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m := scanner.Text()
			p.PrintLnColor(id, colors, i, p.OutColor("STDOUT"), m)
			*hiveChan <- &hive_message.HiveMessage{
				Index:    i,
				Pid:      cmd.Process.Pid,
				Attempt:  attempt,
				Type:     hive_message.ProcessStdOut,
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
			*hiveChan <- &hive_message.HiveMessage{
				Index:    i,
				Pid:      cmd.Process.Pid,
				Attempt:  attempt,
				Type:     hive_message.ProcessStdErr,
				Data:     m,
				ExitCode: noExitCode,
			}
		}
	}()

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		p.PrintLnColor(id, colors, i, p.ErrColor(fmt.Sprintf("%s. Error-exited with code (%d)", cmdSummary, cmd.ProcessState.ExitCode())), err.Error())
		*hiveChan <- &hive_message.HiveMessage{
			Index:    i,
			Pid:      cmd.Process.Pid,
			Attempt:  attempt,
			Type:     hive_message.ProcessExited,
			Data:     err.Error(),
			ExitCode: cmd.ProcessState.ExitCode(),
		}
		return id, cmd.ProcessState.ExitCode()
	} else {
		p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("%s. Success-exited with code (%d)", cmdSummary, cmd.ProcessState.ExitCode())))
		*hiveChan <- &hive_message.HiveMessage{
			Index:    i,
			Pid:      cmd.Process.Pid,
			Attempt:  attempt,
			Type:     hive_message.ProcessExited,
			Data:     "",
			ExitCode: 0, // 0 = success
		}
		return id, 0
	}
}
