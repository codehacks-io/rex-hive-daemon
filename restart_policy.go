package main

type RestartPolicy int

const (
	Always RestartPolicy = iota
	OnFailure
	Never
)

func (r RestartPolicy) String() string {
	return restartPolicyToString[r]
}

var restartPolicyToString = map[RestartPolicy]string{
	Always:    "Always",
	OnFailure: "OnFailure",
	Never:     "Never",
}

var stringToRestartPolicy = map[string]RestartPolicy{
	"Always":    Always,
	"OnFailure": OnFailure,
	"Never":     Never,
}
