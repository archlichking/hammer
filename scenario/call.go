package scenario

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

const (
	RANDOM_RANGE   string = "_random_range_\\([0-9]+,[0-9]+\\)"
	RANDOM_RANGE_X string = "_random_range_\\([0-9]+,[0-9]+,[a-zA-Z]+\\)"
	VALUE_OF       string = "_value_of_\\([a-zA-Z]+\\)"
)

type Call struct {
	RandomWeight      float32
	Weight            float32
	URL, Method, Body string
	Type              string // rest or www or "", default it rest
	RegexpBody        string
	RegexpUrl         string

	_localVarMap            map[string]string
	_reservedMethods        map[string]func(string) string
	_reservedCompiledRegexp map[string]*regexp.Regexp

	GenParam GenCall // to generate URL & Body programmically
	CallBack GenCallBack

	SePoint *Session
}

func (c *Call) normalize() {
	c.Method = strings.ToUpper(c.Method)
	c.Type = strings.ToUpper(c.Type)
}

func (c *Call) parseRegexp() {
	c._localVarMap = make(map[string]string)

	rangeMatcher, _ := regexp.Compile("[0-9]+")
	varMatcher, _ := regexp.Compile("[a-zA-Z]+\\)")

	// initial regexp.compile to reuse
	c._reservedCompiledRegexp = make(map[string]*regexp.Regexp)
	for _, v := range []string{RANDOM_RANGE, RANDOM_RANGE_X, VALUE_OF} {
		reg, _ := regexp.Compile(v)
		c._reservedCompiledRegexp[v] = reg
	}

	// initial reserved regexp methods
	c._reservedMethods = map[string]func(string) string{
		RANDOM_RANGE: func(s string) string {
			r := rangeMatcher.FindAllString(s, -1)
			s0, _ := strconv.Atoi(r[0])
			s1, _ := strconv.Atoi(r[1])

			return strconv.Itoa(rand.Intn(s1-s0) + s0)
		},
		RANDOM_RANGE_X: func(s string) string {
			r := rangeMatcher.FindAllString(s, -1)
			s0, _ := strconv.Atoi(r[0])
			s1, _ := strconv.Atoi(r[1])
			v := strconv.Itoa(rand.Intn(s1-s0) + s0)

			r = varMatcher.FindAllString(s, -1)
			c._localVarMap[r[0]] = v

			return v
		},
		VALUE_OF: func(s string) string {
			r := varMatcher.FindAllString(s, -1)

			if v, not := c._localVarMap[r[0]]; !not {
				return ""
			} else {
				return v
			}
		},
	}

	// 1. filter
	tmpKs := []string{}
	for _, t := range []string{c.URL, c.Body} {
		for pattern, _ := range c._reservedMethods {
			// iterates _reservedMethod to match all to c.Body
			if m, _ := regexp.MatchString(pattern, t); m {
				tmpKs = append(tmpKs, pattern)
			}
		}
	}
	if len(tmpKs) == 0 {
		// no match
		return
	}
	// matches then
	c.RegexpBody = c.Body
	c.RegexpUrl = c.URL
	c.GenParam = GenCall(func(ps ...string) (_m, _t, _u, _b string) {
		// 2. parse
		ret := []string{}
		for _, t := range []string{c.RegexpUrl, c.RegexpBody} {
			for i := range tmpKs {
				genFunc, _ := c._reservedMethods[tmpKs[i]]
				reg, _ := c._reservedCompiledRegexp[tmpKs[i]]
				t = reg.ReplaceAllStringFunc(t, genFunc)
			}
			ret = append(ret, t)
		}
		return c.Method, c.Type, ret[0], ret[1]
	})
}

type GenCall func(ps ...string) (_m, _t, _u, _b string)
type GenCallBack func(se *Session, st int, storage []byte)
