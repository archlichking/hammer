package main

import (
	"flag"
	"fmt"
	"github.com/hammer/auth"
	"github.com/hammer/counter"
	"github.com/hammer/scenario"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

// to reduce size of thread, speed up
const SizePerThread = 10000000

//var DefaultTransport RoundTripper = &Transport{Proxy: ProxyFromEnvironment}

// Counter will be an atomic, to count the number of request handled
// which will be used to print PPS, etc.
type Hammer struct {
	counter *counter.Counter
	db_load *counter.MysqlC

	client  *http.Client
	monitor *time.Ticker
	// ideally error should be organized by type TODO
	throttle    <-chan time.Time
	db_throttle <-chan time.Time
}

// init
func (c *Hammer) Init() {
	c.counter = new(counter.Counter)
	c.db_load = new(counter.MysqlC)
	c.db_load.Init("load_log")

	c.db_load.Open(counter.MysqlConfig{
		Mysql: struct {
			Host, User, Password string
		}{
			Host:     "127.0.0.1:3306",
			User:     "root",
			Password: "",
		},
	})
	// set up HTTP proxy
	if proxy != "none" {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			log.Fatal(err)
		}
		c.client = &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConnsPerHost: 200000,
				Proxy:               http.ProxyURL(proxyUrl),
			},
		}
	} else {
		c.client = &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConnsPerHost: 200000,
			},
		}
	}
}

// main goroutine to drive traffic
func (c *Hammer) hammer(rg *rand.Rand) {
	// before send out, update send count
	c.counter.RecordSend()
	call, err := profile.NextCall(rg)

	if err != nil {
		log.Println("next call error: ", err)
		return
	}

	req, err := http.NewRequest(call.Method, call.URL, strings.NewReader(call.Body))
	// log.Println(call, req, err)
	switch auth_method {
	case "oauth":
		_signature := oauth_client.AuthorizationHeaderWithBodyHash(nil, call.Method, call.URL, url.Values{}, call.Body)
		req.Header.Add("Authorization", _signature)
	}

	// Add special haeader for PATCH, PUT and POST
	switch call.Method {
	case "PATCH", "PUT", "POST":
		switch call.Type {
		case "REST":
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
			break
		case "WWW":
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			break
		}
	}

	t1 := time.Now().UnixNano()
	res, err := c.client.Do(req)

	response_time := time.Now().UnixNano() - t1

	/*
		    ###
			disable reading res.body, no need for our purpose for now,
		    by doing this, hope we can save more file descriptor.
			##
	*/
	defer req.Body.Close()

	switch {
	case err != nil:
		log.Println("Response Time: ", float64(response_time)/1.0e9, " Erorr: when", call.Method, call.URL, "with error ", err)
		c.counter.RecordError()
	case res.StatusCode >= 400 && res.StatusCode != 409:
		log.Println("Got error code --> ", res.Status, "for call ", call.Method, " ", call.URL)
		c.counter.RecordError()
	default:
		// only do successful response here
		defer res.Body.Close()
		c.counter.RecordRes(response_time, slowThreshold, call.URL)
		data, _ := ioutil.ReadAll(res.Body)
		if call.CallBack == nil && !debug {
		} else {
			if res.StatusCode == 409 {
				log.Println("Http 409 Res Body : ", string(data))
			}
			if debug {
				log.Println("Req : ", call.Method, call.URL)
				if auth_method != "none" {
					log.Println("Authorization: ", string(req.Header.Get("Authorization")))
				}
				log.Println("Req Body : ", call.Body)
				log.Println("Response: ", res.Status)
				log.Println("Res Body : ", string(data))
			}
			if call.CallBack != nil {
				call.CallBack(call.SePoint, scenario.NEXT, data)
			}
		}
	}

}

func (c *Hammer) monitorHammer() {
	log.Println(c.counter.GeneralStat(), profile.CustomizedReport())
}

func (c *Hammer) launch(rps int64) {
	// var _rps time.Duration

	_p := time.Duration(rps)
	_interval := 1000000000.0 / _p
	c.throttle = time.Tick(_interval * time.Nanosecond)
	c.db_throttle = time.Tick(3 * time.Second)
	// var wg sync.WaitGroup

	log.Println("run with rps -> ", int(_p))
	go func() {
		i := 0
		for {
			if i == 8 {
				i = 0
			}
			<-c.throttle
			go c.hammer(rands[i])
			i++
		}
	}()

	c.monitor = time.NewTicker(time.Second)
	go func() {
		for {
			<-c.monitor.C // rate limit for monitor routine
			go c.monitorHammer()
		}
	}()

	// db flush
	go func() {
		for {
			<-c.db_throttle
			c.db_load.Flush(c.counter)
		}
	}()
}

// init the program from command line
var (
	rps           int64
	profileFile   string
	profileType   string
	slowThreshold int64
	debug         bool
	auth_method   string
	sessionAmount int
	proxy         string

	// rands
	rands []*rand.Rand
	// profile
	profile scenario.Profile

	oauth_client = new(auth.Client)
)

func init() {
	flag.Int64Var(&rps, "rps", 500, "Set Request Per Second")
	flag.StringVar(&profileFile, "profile", "", "The path to the traffic profile")
	flag.Int64Var(&slowThreshold, "threshold", 200, "Set slowness standard (in millisecond)")
	flag.StringVar(&profileType, "type", "default", "Profile type (default|session|your session type)")
	flag.BoolVar(&debug, "debug", false, "debug flag (true|false)")
	flag.StringVar(&auth_method, "auth", "none", "Set authorization flag (oauth|gree(c|s)2s|none)")
	flag.IntVar(&sessionAmount, "size", 100, "session amount")
	flag.StringVar(&proxy, "proxy", "none", "Set HTTP proxy (need to specify scheme. e.g. http://127.0.0.1:8888)")
}

// main func
func main() {

	flag.Parse()
	NCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(2)

	// to speed up
	rands = make([]*rand.Rand, NCPU)
	for i, _ := range rands {
		s := rand.NewSource(time.Now().UnixNano())
		rands[i] = rand.New(s)
		rands[i].Seed(time.Now().UnixNano())
	}

	log.Println("cpu number -> ", NCPU)
	log.Println("rps -> ", rps)
	log.Println("slow threshold -> ", slowThreshold, "ms")
	log.Println("profile type -> ", profileType)
	log.Println("Proxy -> ", proxy)

	profile, _ = scenario.New(profileType, sessionAmount)
	if profileFile != "" {
		profile.InitFromFile(profileFile)
	} else {
		profile.InitFromCode()
	}

	rand.Seed(time.Now().UnixNano())

	ham := new(Hammer)
	ham.Init()

	go ham.launch(rps)

	var input string
	fmt.Scanln(&input)
}
