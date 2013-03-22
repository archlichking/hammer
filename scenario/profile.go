package scenario

import (
	"errors"
)

type Profile interface {
	InitFromFile(string)
	InitFromCode()
	NextCalls() ([]*Call, int)
}

var scenarios = make(map[string]func() (Profile, error))

func Register(name string, scenario func() (Profile, error)) {
	scenarios[name] = scenario
}

func New(scenarioName string) (Profile, error) {
	if scenario, ok := scenarios[scenarioName]; ok {
		return scenario()
	}

	return nil, errors.New("scenario is not registered")
}
