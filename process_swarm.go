package main

import (
	"gopkg.in/yaml.v3"
	"os"
)

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
