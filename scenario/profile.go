package scenario

import(
	"fmt"
	"io"
  	"os"
	"log"
	"math/rand"
	"strings"
	"encoding/json"
)

// type Profile interface{
// 	InitFromFile(path string) 
// 	NextCall() (*Call)
// 	Print() string
// 	AddNewCall(weight float32, method, _type string, genf func() (string, string))
// }

type Scenario struct {
	_totalWeight float32
	_calls       [100]Call
	_num         int
	// _traffic     [100]int64 //to track

	InitFromCode func(*Scenario)
}

func (s *Scenario) InitFromFile(path string) {
	buf := make([]byte, 2048)

	f, _ := os.Open(path)
	f.Read(buf)

	dec := json.NewDecoder(strings.NewReader(string(buf)))
	for {
		var m Call
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			//log.Println(err)
			// TODO, fix error handling
			break
		}

		m.normalize()
		s._calls[s._num] = m

		s._totalWeight = s._totalWeight + m.Weight
		s._calls[s._num].RandomWeight = s._totalWeight
		log.Print(s._calls[s._num])

		s._num++
		fmt.Printf("Import Call -> W: %f URL: %s  Method: %s\n", m.Weight, m.URL, m.Method)
	}
}

func (s *Scenario) AddNewCall(weight float32, method, _type string, gen func() (string, string)){
	s._totalWeight = s._totalWeight + weight
	s._calls[s._num].RandomWeight = s._totalWeight
	s._calls[s._num].Method = strings.ToUpper(method)
	s._calls[s._num].URL = ""
	s._calls[s._num].Body = ""
	s._calls[s._num].Type = _type

	s._calls[s._num].GenFunc = gen

	s._calls[s._num].normalize()

	s._num++
	fmt.Printf("Import Call -> W: %f URL: with func Method: %s\n", weight, method)
}

func (s *Scenario) NextCall() (*Call) {
	r := rand.Float32() * s._totalWeight
	for i := 0; i < s._num; i++ {
		if r <= s._calls[i].RandomWeight {
			if s._calls[i].GenFunc != nil {
				s._calls[i].URL, s._calls[i].Body = s._calls[i].GenFunc()
			} 
			return &s._calls[i]
		}
	}

	log.Fatal("what? should never reach here")
	return &s._calls[1]
}

func (s *Scenario) Print() (string) {
	var x string
	for i := 0; i < s._num; i++ {
		x = x + s._calls[i].Print() + "\n+++++++\n"
	}
	return x
}

func New() (*Scenario){
	return &Scenario{}
}