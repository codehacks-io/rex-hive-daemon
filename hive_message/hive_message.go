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
	// TempId is a temp UUID assigned at runtime before storing it to mongo DB. It's not persisted in MongoDB.
	// The actual ID of these entities will be a MongoDB ObjectID which is faster and works better for message logs
	// which can be stored a rates of hundreds per second. ObjectIDs also work better at keeping the order in which
	// entities are stored which is important when storing and retrieving log messages.
	TempId         string                    `bson:"-"`
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
