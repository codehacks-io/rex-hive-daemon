package main

import (
	"bufio"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
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

type ProcessSwarm struct {
	Kind     string
	Metadata struct {
		Name string
	}
	Spec struct {
		ProcessSpecs []struct {
			Name string
			Env  []struct {
				Name      string
				Value     string
				ValueFrom struct {
					SecretKeyRef struct {
						Name string
						Key  string
					}
				}
			}
			Cmd      []string
			Restart  string
			Replicas int
		} `yaml:"processes"`
	}
}

func readConf(filename string) (*ProcessSwarm, error) {

	// Read file
	buff, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Parse data
	data := &ProcessSwarm{}
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
	swarmSpec, err := readConf(*filePathPtr)
	if err != nil {
		panic(err)
	}

	runProcessSwarm(swarmSpec)
}

var getUniqueInSequenceRegex = regexp.MustCompile(`{unique-in-sequence:(?P<from>\d+)-(?P<to>\d+)}`)

func getDynamicArgs(originalArgs []string, used *map[int]bool) []string {

	replacedArgs := make([]string, len(originalArgs))
	copy(replacedArgs, originalArgs)

	for i, a := range originalArgs {
		match := getUniqueInSequenceRegex.FindStringSubmatch(a)

		result := make(map[string]string)
		for ii, name := range getUniqueInSequenceRegex.SubexpNames() {
			if ii != 0 && name != "" && len(match) > ii {
				result[name] = match[ii]
			}
		}
		if len(result["from"]) > 0 && len(result["to"]) > 0 {
			from, _ := strconv.Atoi(result["from"])
			to, _ := strconv.Atoi(result["to"])

			// Swap is from is greater than to
			if from > to {
				oldTo := to
				to = from
				from = oldTo
			}

			didAssign := false
			for seq := from; seq <= to; seq++ {
				if !(*used)[seq] {
					replacedArgs[i] = strconv.Itoa(seq)
					(*used)[seq] = true
					didAssign = true
					break
				}
			}
			if !didAssign {
				panic(fmt.Sprintf("dynamic argument %s cannot be allocated a value, all values in the sequense have been reserved", a))
			}
		}
	}

	return replacedArgs
}

func runProcessSwarm(swarmSpec *ProcessSwarm) {

	if len((*swarmSpec).Spec.ProcessSpecs) < 1 {
		fmt.Println("No process specs to run")
		return
	} else {
		fmt.Println(fmt.Sprintf("Found %d process specs", len((*swarmSpec).Spec.ProcessSpecs)))
	}

	// Execute
	var wg sync.WaitGroup
	colors := p.GetRandomColors()
	var usedNumsInSequence = map[int]bool{}

	count := 0
	for _, s := range swarmSpec.Spec.ProcessSpecs {
		for rep := 0; rep < s.Replicas; rep++ {
			wg.Add(1)
			args := getDynamicArgs(s.Cmd[1:], &usedNumsInSequence)
			go runCommandAndKeepAlive(count, &wg, colors, stringToRestartPolicy[s.Restart], s.Cmd[0], args...)
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
		id, exitCode := runCommand(i, restartPolicy, runCount, colors, command, args...)

		// Get elapsed runtime of command
		elapsed := time.Since(startedAt)

		// Reset backoff
		if elapsed.Seconds() >= backoffResetIfUpSeconds {
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
				backoff := expBackoffSeconds(backoffCount)
				p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("will re-run after %s", backoff)))
				time.Sleep(backoff)
			}
		case OnFailure:
			{
				if exitCode == 0 {
					p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("won't re-run")))
					return
				} else {
					backoff := expBackoffSeconds(backoffCount)
					p.PrintLnColor(id, colors, i, p.Dim(fmt.Sprintf("will re-run after %s", backoff)))
					time.Sleep(backoff)
				}
			}
		}
	}
}

func runCommand(i int, restart RestartPolicy, attempt int, colors []int, command string, args ...string) (name string, pid int) {

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
