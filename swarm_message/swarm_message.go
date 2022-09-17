package swarm_message

import "rex-daemon/machine_meta"

type swarmMessageType int

const (
	ProcessAborted swarmMessageType = iota
	ProcessStarted
	ProcessExited
	ProcessStdOut
	ProcessStdErr
)

type SwarmMessage struct {
	Index    int
	Pid      int
	Attempt  int
	Type     swarmMessageType
	Data     string
	ExitCode int
	Machine  *machine_meta.MachineMeta
}
