package scenario

type Call struct {
	RandomWeight      float32
	Weight            float32
	URL, Method, Body string
	Type              string // rest or www or "", default it rest

	GenFunc GenRequest // to generate URL & Body programmically
}

type CallGroup struct {
	RandomWeight float32
	Weight       float32
	Calls        []*Call

	// GenFunc func() (_method, _type, _url, _body string) // to generate URL & Body programmically
}

type GenRequest func() (_m, _t, _u, _b string)
