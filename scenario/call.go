package scenario

import (
	"sync/atomic"
	"fmt"
	"strings"
)

type Call struct {
	RandomWeight      float32
	Weight            float32
	URL, Method, Body string
	Type              string // rest or www or "", default it rest

	GenFunc func() (_url, _body string) // to generate URL & Body programmically 
	count     int64 // total # of request
	totaltime int64 // total response time.
	backlog   int64
}

func (c *Call) Record(_time int64) {
	atomic.AddInt64(&c.count, 1)
	atomic.AddInt64(&c.totaltime, _time)
}

func (c *Call) Print() string {
	return "API : " + c.Method + "  " + c.URL +
		"\nTotal Call : " + fmt.Sprintf("%d", c.count) +
		"\nResponse Time : " + fmt.Sprintf("%2.4f", float64(c.totaltime)/(float64(c.count)*1.0e9))
}

func (c *Call) normalize() {
	c.Method = strings.ToUpper(c.Method)
	c.Type = strings.ToUpper(c.Type)
}
