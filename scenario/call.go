package scenario

import (
	"strings"
)

type Call struct {
	RandomWeight      float32
	Weight            float32
	URL, Method, Body string
	Type              string // rest or www or "", default it rest

	GenFunc func() (_method, _type, _url, _body string) // to generate URL & Body programmically
}

func (c *Call) normalize() {
	c.Method = strings.ToUpper(c.Method)
	c.Type = strings.ToUpper(c.Type)
}
