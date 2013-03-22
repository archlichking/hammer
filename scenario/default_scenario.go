package scenario

import (
)

type DefaultScenario struct{
	Scenario
}

func (ds *DefaultScenario) InitFromCode(){
	ds.addNewCall(50, func ()(string, string, string, string){
		return "POST", "REST", "http://localhost:9000/hello", "{}"
		})
	ds.addNewCall(50, func ()(string, string, string, string){
		return "POST", "REST", "http://localhost:9000/hello_in_json", "{}"
		})
}


func (ds *DefaultScenario) addNewCall(weight float32, gen func() (_m, _t, _u, _b string)){
	ds._totalWeight = ds._totalWeight + weight
	ds._calls[ds._num].RandomWeight = ds._totalWeight
	ds._calls[ds._num].Method = ""
	ds._calls[ds._num].URL = ""
	ds._calls[ds._num].Body = ""
	ds._calls[ds._num].Type = ""

	ds._calls[ds._num].GenFunc = gen

	ds._calls[ds._num].normalize()

	ds._num++
}

func init() {
	Register("default", newDefaultScenario)
}

func newDefaultScenario() (Profile, error){
	return &DefaultScenario{}, nil
}