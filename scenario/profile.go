package scenario

import (
	"errors"
)

type Profile interface {
	InitFromFile(string)
	InitFromCode()
	NextCall() (*Call, error)
	CustomizedReport() string
}

var scenarios = make(map[string]func(int) (Profile, error))

func Register(name string, scenario func(int) (Profile, error)) {
	scenarios[name] = scenario
}

func New(scenarioName string, sessionSize int) (Profile, error) {
	if scenario, ok := scenarios[scenarioName]; ok {
		return scenario(sessionSize)
	}

	return nil, errors.New("scenario is not registered")
}
