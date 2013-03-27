package scenario

import (
	// "encoding/json"
	// "fmt"
	// "io"
	"log"
	"math/rand"
	// "os"
	// "strings"
)

type DefaultScenario struct {
	_totalWeight float32
	_callGroups  [100]CallGroup
	_count       int
}

func (ds *DefaultScenario) InitFromFile(path string) {
	// buf := make([]byte, 2048)

	// f, _ := os.Open(path)
	// f.Read(buf)

	// dec := json.NewDecoder(strings.NewReader(string(buf)))
	// for {
	// 	var m Call
	// 	if err := dec.Decode(&m); err == io.EOF {
	// 		break
	// 	} else if err != nil {
	// 		//log.Println(err)
	// 		// TODO, fix error handling
	// 		break
	// 	}

	// 	ds._calls[ds._count] = m

	// 	ds._totalWeight = ds._totalWeight + m.Weight
	// 	ds._calls[ds._count].RandomWeight = ds._totalWeight
	// 	log.Print(ds._calls[ds._count])

	// 	ds._count++
	// 	fmt.Printf("Import Call -> W: %f URL: %s  Method: %s\n", m.Weight, m.URL, m.Method)
	// }
}

func (ds *DefaultScenario) NextCalls() (*CallGroup, int) {
	// r := rand.Float32() * ds._totalWeight
	// for i := 0; i < ds._count; i++ {
	// 	if r <= ds._calls[i].RandomWeight {
	// 		if ds._calls[i].GenFunc != nil {
	// 			ds._calls[i].Method, ds._calls[i].Type, ds._calls[i].URL, ds._calls[i].Body = ds._calls[i].GenFunc()
	// 		}
	// 		return []*Call{&ds._calls[i]}, -1
	// 	}
	// }

	// log.Fatal("what? should never reach here")
	// return []*Call{&ds._calls[0]}, -1
	r := rand.Float32() * ds._totalWeight
	for i := 0; i < ds._count; i++ {
		if r <= ds._callGroups[i].RandomWeight {
			for _, c := range ds._callGroups[i].Calls {
				if c.GenFunc != nil {
					c.Method, c.Type, c.URL, c.Body = c.GenFunc()
				}
			}
			return &ds._callGroups[i], -1
		}
	}

	log.Fatal("what? should never reach here")
	return &ds._callGroups[0], -1
}

func (ds *DefaultScenario) InitFromCode() {

	ds.addCallGroup(50, []GenRequest{
		GenRequest(func(...string) (_m, _t, _u, _b string) {
			return "GET", "REST", "http://localhost:9000/hello", "{}"
		}),
	})
	ds.addCallGroup(50, []GenRequest{
		GenRequest(func(...string) (_m, _t, _u, _b string) {
			return "GET", "REST", "http://localhost:9000/hello_in_json", "{}"
		}),
	})
}

func (ds *DefaultScenario) addCallGroup(weight float32, gens []GenRequest) {
	ds._totalWeight = ds._totalWeight + weight
	ds._callGroups[ds._count].RandomWeight = ds._totalWeight
	ds._callGroups[ds._count].Calls = make([]*Call, len(gens))
	for i := 0; i < len(gens); i++ {
		ds._callGroups[ds._count].Calls[i] = new(Call)
		ds._callGroups[ds._count].Calls[i].GenFunc = gens[i]
	}
	ds._count++
}

// func (ds *DefaultScenario) addNewCall(weight float32, gen GenRequest) {
// 	ds._totalWeight = ds._totalWeight + weight
// 	ds._calls[ds._count].RandomWeight = ds._totalWeight
// 	ds._calls[ds._count].Method = ""
// 	ds._calls[ds._count].URL = ""
// 	ds._calls[ds._count].Body = ""
// 	ds._calls[ds._count].Type = ""
// 	ds._calls[ds._count].GenFunc = gen

// 	ds._count++
// }

func init() {
	Register("default", newDefaultScenario)
}

func newDefaultScenario() (Profile, error) {
	return &DefaultScenario{}, nil
}
