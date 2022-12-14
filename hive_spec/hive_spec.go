package hive_spec

import (
	"gopkg.in/yaml.v3"
	"os"
	"rex-hive-daemon/machine_meta"
	"time"
)

type ProcessSpec struct {
	Name string `bson:"name"`
	// ForwardOsEnv when set to true, will forward all the OS env vars (`os.Environ()`) to the process.
	// If instead of running a compiled binary, you're running `go run` e.g: `go run my-main.go`, you need to set this
	// to true.
	ForwardOsEnv bool `yaml:"forwardOsEnv" bson:"forwardOsEnv"`
	Env          []struct {
		Name      string `bson:"name"`
		Value     string `bson:"value"`
		ValueFrom struct {
			SecretKeyRef struct {
				Name string `bson:"name"`
				Key  string `bson:"key"`
			} `bson:"secretKeyRef"`
		} `bson:"valueFrom"`
	} `bson:"env"`
	Cmd      []string `bson:"cmd"`
	Restart  string   `bson:"restart"`
	Replicas int      `bson:"replicas"`
}

// HiveSpec is the formal definition of how one or multiple processes will run in a machine. Once a HiveSpec is executed
// the group of processes that are running is called a "HiveRun". A HiveRun is assigned an ID once registered in DB.
type HiveSpec struct {
	Id       string    `bson:"_id"`
	Time     time.Time `bson:"time"`
	Kind     string    `bson:"kind"`
	Metadata struct {
		Name string `bson:"name"`
	} `bson:"metadata"`
	Spec struct {
		Processes []*ProcessSpec `yaml:"processes" bson:"processes"`
	} `bson:"spec"`
	// This field os not populated by the yml spec but at run time
	RuntimeMachine *machine_meta.MachineMeta `bson:"runtimeMachine,omitempty"`
}

func FromFile(filename string) (*HiveSpec, error) {

	// Read file
	buff, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Parse data
	data := &HiveSpec{}
	err = yaml.Unmarshal(buff, data)

	if err != nil {
		return nil, err
	}

	return data, err
}
