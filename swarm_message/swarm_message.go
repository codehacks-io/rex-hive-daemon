package swarm_message

import (
	"rex-daemon/machine_meta"
	"time"
)

type swarmMessageType string

const (
	ProcessAborted swarmMessageType = "aborted"
	ProcessStarted swarmMessageType = "started"
	ProcessExited  swarmMessageType = "exited"
	ProcessStdOut  swarmMessageType = "stdout"
	ProcessStdErr  swarmMessageType = "stderr"
)

type SwarmMessage struct {
	Index    int
	Pid      int
	Attempt  int
	Type     swarmMessageType
	Data     string
	ExitCode int
	Machine  *machine_meta.MachineMeta
	Time     time.Time
}
