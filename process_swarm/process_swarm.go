package process_swarm

import (
	"gopkg.in/yaml.v3"
	"os"
)

type ProcessSwarm struct {
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
}

func FromFile(filename string) (*ProcessSwarm, error) {

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
