package scenario

import (
	"os"
	"encoding/json"
	"strings"
	"io"
	"log"
	"math/rand"
	"fmt"
	"errors"
)

type Scenario struct {
	_totalWeight float32
	_calls       []*Call
	_count         int
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
		s._calls[s._count] = &m

		s._totalWeight = s._totalWeight + m.Weight
		s._calls[s._count].RandomWeight = s._totalWeight
		log.Print(s._calls[s._count])

		s._count++
		fmt.Printf("Import Call -> W: %f URL: %s  Method: %s\n", m.Weight, m.URL, m.Method)
	}
}

func (s *Scenario) NextCall() (*Call, error) {
	r := rand.Float32() * s._totalWeight
	for i := 0; i < s._count; i++ {
		if r <= s._calls[i].RandomWeight {
			if s._calls[i].GenParam != nil {
				s._calls[i].Method, s._calls[i].Type, s._calls[i].URL, s._calls[i].Body = s._calls[i].GenParam()
			} 
			return s._calls[i], nil
		}
	}

	return nil, errors.New("something wrong with randomize number")
}