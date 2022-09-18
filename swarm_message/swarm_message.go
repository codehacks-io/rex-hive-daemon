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
	Index          int                       `bson:"index"`
	Pid            int                       `bson:"pid"`
	Attempt        int                       `bson:"attempt"`
	Type           swarmMessageType          `bson:"type"`
	Data           string                    `bson:"data,omitempty"`
	ExitCode       int                       `bson:"exitCode"`
	RuntimeMachine *machine_meta.MachineMeta `bson:"runtimeMachine,omitempty"`
	Time           time.Time                 `bson:"time"`
}
