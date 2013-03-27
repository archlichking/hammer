package scenario

import (
	"log"
	"math/rand"
)

type SessionScenario struct {
	_totalWeight float32
	_callGroups  [100]CallGroup
	_count       int

	_mustDone int
}

func (ss *SessionScenario) InitFromFile(path string) {
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

	// 	m.normalize()
	// 	ds._calls[ds._count] = m

	// 	ds._totalWeight = ds._totalWeight + m.Weight
	// 	ds._calls[ds._count].RandomWeight = ds._totalWeight
	// 	log.Print(ds._calls[ds._count])

	// 	ds._count++
	// 	fmt.Printf("Import Call -> W: %f URL: %s  Method: %s\n", m.Weight, m.URL, m.Method)
	// }
}

func (ss *SessionScenario) InitFromCode() {
	ss.addCallGroup(100, []GenRequest{
		GenRequest(func(...string) (_m, _t, _u, _b string) {
			// return "POST", "REST", "http://localhost:9988/post", "{\"fsdfsdfsdf\":\"ddddd\"}"
			return "GET", "REST", "http://localhost:9988/get", "{}"
		}),
		GenRequest(func(...string) (_m, _t, _u, _b string) {
			return "GET", "REST", "http://localhost:9988/get", "{}"
		}),
		GenRequest(func(...string) (_m, _t, _u, _b string) {
			return "GET", "REST", "http://localhost:9988/get", "{}"
		}),
	})
}

func (ss *SessionScenario) NextCalls() (*CallGroup, int) {
	r := rand.Float32() * ss._totalWeight
	for i := 0; i < ss._count; i++ {
		if r <= ss._callGroups[i].RandomWeight {
			// for _, c := range ss._callGroups[i].Calls {
			// 	if c.GenFunc != nil {
			// 		c.Method, c.Type, c.URL, c.Body = c.GenFunc()
			// 	}
			// }
			ss._callGroups[i].BufferedChn = make(chan string, 1)
			return &ss._callGroups[i], ss._mustDone
		}
	}

	log.Fatal("what? should never reach here")
	return &ss._callGroups[0], ss._mustDone
}

func (ss *SessionScenario) addCallGroup(weight float32, gens []GenRequest) {
	ss._totalWeight = ss._totalWeight + weight
	ss._callGroups[ss._count].RandomWeight = ss._totalWeight
	ss._callGroups[ss._count].Calls = make([]*Call, len(gens))
	for i := 0; i < len(gens); i++ {
		ss._callGroups[ss._count].Calls[i] = new(Call)
		ss._callGroups[ss._count].Calls[i].GenFunc = gens[i]
	}
	ss._count++
}

func init() {
	Register("session", newSessionScenario)
}

func newSessionScenario() (Profile, error) {
	return &SessionScenario{
		_mustDone: 1,
	}, nil
}
