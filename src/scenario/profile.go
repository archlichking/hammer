package scenario

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

const (
	VAR                string = "\\${[0-9a-zA-Z]+}"
	RANDOM_RANGE_FLOAT string = "_random_range_float_\\([0-9]+,[0-9]+\\)"
	RANDOM_RANGE_INT   string = "_random_range_int_\\([0-9]+,[0-9]+\\)"
)

type Profile struct {
	Client      string
	TotalWeight float32
	Scenarios   []*Scenario
	Variables   map[string][]string

	numMatcher            *regexp.Regexp
	varMatcher            *regexp.Regexp
	regexpMethods         map[string]func(string) string
	regexpCompiledMatcher map[string]*regexp.Regexp
}

func New(file string) (*Profile, error) {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	p := new(Profile)
	err = json.Unmarshal(f, p)
	if err != nil {
		fmt.Printf("%#s\n", err)
		return nil, err
	}

	// calculate weight for each call
	for i, _ := range p.Scenarios {

		w := p.Scenarios[i].Weight

		l := len(p.Scenarios[i].Groups)
		for j, _ := range p.Scenarios[i].Groups {
			iw := float32(w / float32(l))
			p.TotalWeight += iw
			p.Scenarios[i].Groups[j].Weight = p.TotalWeight

			if strings.ToUpper(p.Scenarios[i].Type) == "SESSION" {
				p.Scenarios[i].Groups[j].Sessions = make([]*Session, 400)
				for k, _ := range p.Scenarios[i].Groups[j].Sessions {
					p.Scenarios[i].Groups[j].Sessions[k] = new(Session)
					p.Scenarios[i].Groups[j].Sessions[k].Lock = make(chan int, 1)
					p.Scenarios[i].Groups[j].Sessions[k].Lock <- 0
					p.Scenarios[i].Groups[j].Sessions[k].Calls = p.Scenarios[i].Groups[j].Calls
				}
			}

			// fmt.Printf("%#v\n", p.Scenarios[i].Groups[j])
		}
		p.Scenarios[i].Weight = p.TotalWeight
		// fmt.Printf("%#v\n", p.Scenarios[i])
	}

	// compile varialbles if it's not empty
	p.numMatcher, _ = regexp.Compile("[0-9]+")
	p.varMatcher, _ = regexp.Compile("[0-9a-zA-Z]+")

	p.regexpCompiledMatcher = make(map[string]*regexp.Regexp)
	for _, v := range []string{VAR, RANDOM_RANGE_FLOAT, RANDOM_RANGE_INT} {
		c, _ := regexp.Compile(v)
		p.regexpCompiledMatcher[v] = c
	}

	p.regexpMethods = map[string]func(string) string{
		VAR: func(s string) string {
			r := p.varMatcher.FindAllString(s, -1)
			for _, v := range r {
				if p.Variables[v] != nil {
					return p.Variables[v][rand.Intn(len(p.Variables[v]))]
				}
			}
			return s
		},
		RANDOM_RANGE_FLOAT: func(s string) string {
			r := p.numMatcher.FindAllString(s, -1)
			s0, _ := strconv.ParseFloat(r[0], 32)
			s1, _ := strconv.ParseFloat(r[1], 32)

			return strconv.FormatFloat(rand.Float64()*(s1-s0)+s0, 'f', 16, 64)
		},
		RANDOM_RANGE_INT: func(s string) string {
			r := p.numMatcher.FindAllString(s, -1)
			s0, _ := strconv.Atoi(r[0])
			s1, _ := strconv.Atoi(r[1])

			return strconv.Itoa(rand.Intn(s1-s0) + s0)
		},
	}

	// fmt.Printf("%#v\n", p)

	return p, nil
}

func (p *Profile) varMatch(in string) string {
	c := in
	for pattern, function := range p.regexpMethods {
		cp := p.regexpCompiledMatcher[pattern]
		c = cp.ReplaceAllStringFunc(c, function)
	}
	return c
}

func (p *Profile) NextCall(rg *rand.Rand) (*Call, *Session, int, error) {
	r := rg.Float32() * p.TotalWeight

	for i := 0; i < len(p.Scenarios); i++ {
		if r <= p.Scenarios[i].Weight {
			call, session, index, err := p.Scenarios[i].NextAvailable(r, rg)
			c := new(Call)
			if call != nil {
				if call.Body != "" {
					c.Body = p.varMatch(call.Body)
				}
				if call.URL != "" {
					c.URL = p.varMatch(call.URL)
				}
				c.BodyType = call.BodyType
				c.Header = call.Header
				c.Method = call.Method
			}
			return c, session, index, err
		}
	}

	return nil, nil, -1, errors.New("something wrong with randomize number in Profile.NextCall")
}

func (p *Profile) CustomizedReport() string {
	return ""
}
