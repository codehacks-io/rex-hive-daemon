package hive_spec

import (
	"gopkg.in/yaml.v3"
	"os"
	"rex-hive-daemon/machine_meta"
)

// HiveSpec is the formal definition of how one or multiple processes will run in a machine. Once a HiveSpec is executed
// the group of processes that are running is called a "HiveRun". A HiveRun is assigned an ID once registered in DB.
type HiveSpec struct {
	Kind     string `bson:"kind"`
	Metadata struct {
		Name string `bson:"name"`
	} `bson:"metadata"`
	Spec struct {
		ProcessSpecs []struct {
			Name string `bson:"name"`
			Env  []struct {
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
		} `yaml:"processes" bson:"processSpecs"`
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
