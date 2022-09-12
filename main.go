package main

import (
	"bufio"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"math"
	"os"
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

func (r RestartPolicy) String() string {
	return restartPolicyToString[r]
}

var restartPolicyToString = map[RestartPolicy]string{
	Always:    "Always",
	OnFailure: "OnFailure",
	Never:     "Never",
}

var stringToRestartPolicy = map[string]RestartPolicy{
	"Always":    Always,
	"OnFailure": OnFailure,
	"Never":     Never,
}

// Note: struct fields must be public in order for unmarshal to correctly populate the data.
type fleetSpec struct {
	Metadata struct {
		Name string `yaml:"name"`
	}
	Specs []struct {
		Cmd      []string
		Restart  string
		Replicas int
	}
}

func readConf(filename string) (*fleetSpec, error) {

	// Read file
	buff, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Parse data
	data := &fleetSpec{}
	err = yaml.Unmarshal(buff, data)

	if err != nil {
		return nil, err
	}

	return data, err
}

// Eg. Run with: `go run .\main.go --file=./demo-specs/test-spec.yml`
func main() {
	// Define cli params
	filePathPtr := flag.String("file", "", "spec file containing args")
	flag.Parse()

	// Read and parse file
	fleets, err := readConf(*filePathPtr)
	if err != nil {
		panic(err)
	}

	// Execute
	var wg sync.WaitGroup
	colors := p.GetRandomColors()

	count := 0
	for _, f := range fleets.Specs {
		for rep := 0; rep < f.Replicas; rep++ {
			wg.Add(1)
			go runCommandAndKeepAlive(count, &wg, colors, stringToRestartPolicy[f.Restart], f.Cmd[0], f.Cmd[1:]...)
			count++
		}
	}

	wg.Wait()
}

const backoffBaseDelaySeconds = 5
const backoffResetIfUpSeconds = 600

func expBackoffSeconds(attempt int) time.Duration {
	// Cap to 5 minutes
	if attempt >= 6 {
		return time.Second * 300
	}

	if attempt < 0 {
		return 0
	}

	return time.Second * time.Duration(math.Pow(2, float64(attempt))*backoffBaseDelaySeconds)
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
		if elapsed.Seconds() >= backoffResetIfUpSeconds {
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
