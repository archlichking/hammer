package scenario

type EQScenario struct {
	DefaultScenario
}

func (ds *EQScenario) InitFromCode() {
	ds.addNewCall(50, func() (string, string, string, string) {
		return "POST", "REST", "http://localhost:9000/hello", "{}"
	})
	ds.addNewCall(50, func() (string, string, string, string) {
		return "POST", "REST", "http://localhost:9000/hello_in_json", "{}"
	})
}

func init() {
	Register("eq", newEQScenario)
}

func newEQScenario() (Profile, error) {
	return &EQScenario{}, nil
}
