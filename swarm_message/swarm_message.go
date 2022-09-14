package swarm_message

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
}
