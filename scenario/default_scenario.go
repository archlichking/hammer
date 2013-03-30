package scenario

import (
)

type DefaultScenario struct {
	Scenario
}

func (ds *DefaultScenario) InitFromCode() {
	ds._calls = make([]*Call, 100)
	ds.addCall(5, GenCall(func(...string) (_m, _t, _u, _b string) {
			return "POST", "REST", "http://localhost:9988/post", "{\"fsdfsdfsdf\":\"ddddd\"}"
		}))
	ds.addCall(35, GenCall(func(...string) (_m, _t, _u, _b string) {
			return "GET", "REST", "http://localhost:9988/get", "{}"
		}))
	ds.addCall(60, GenCall(func(...string) (_m, _t, _u, _b string) {
			return "GET", "REST", "http://localhost:9988/get", "{}"
		}))
}

func (ds *DefaultScenario) addCall(weight float32, gp GenCall) {
	ds._totalWeight = ds._totalWeight + weight
	ds._calls[ds._count] = new(Call)
	ds._calls[ds._count].RandomWeight = ds._totalWeight
	ds._calls[ds._count].GenParam = gp

	ds._calls[ds._count].normalize()
	ds._count++
}

func init() {
	Register("default", newDefaultScenario)
}

func newDefaultScenario() (Profile, error) {
	return &DefaultScenario{}, nil
}
