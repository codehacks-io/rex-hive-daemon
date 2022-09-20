package hive_message

import (
	"rex-hive-daemon/machine_meta"
	"time"
)

type hiveMessageType string

const (
	ProcessAborted hiveMessageType = "aborted"
	ProcessStarted hiveMessageType = "started"
	ProcessExited  hiveMessageType = "exited"
	ProcessStdOut  hiveMessageType = "stdout"
	ProcessStdErr  hiveMessageType = "stderr"
)

type HiveMessage struct {
	Id             string                    `bson:"_id"`
	Index          int                       `bson:"index"`
	Pid            int                       `bson:"pid"`
	Attempt        int                       `bson:"attempt"`
	Type           hiveMessageType           `bson:"type"`
	Data           string                    `bson:"data,omitempty"`
	ExitCode       int                       `bson:"exitCode"`
	HiveRunId      interface{}               `bson:"hiveRunId"`
	RuntimeMachine *machine_meta.MachineMeta `bson:"runtimeMachine,omitempty"`
	Time           time.Time                 `bson:"time"`
}
