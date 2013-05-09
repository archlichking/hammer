package main

import (
	"flag"
	"fmt"
	"github.com/hammer/auth"
	"github.com/hammer/counter"
	"github.com/hammer/logg"
	"github.com/hammer/scenario"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
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

	client  *http.Client
	monitor *time.Ticker
	// ideally error should be organized by type TODO
	throttle <-chan time.Time
	// counterArray [][]int64
}

// init
func (c *Hammer) Init() {
	c.counter = new(counter.Counter)
	// c.counterArray = make([][]int64, 0)
	// set up HTTP proxy
	if proxy != "none" {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			log.Fatal(err)
		}
		c.client = &http.Client{
			Transport: &http.Transport{
				// DisableKeepAlives:   false,
				// MaxIdleConnsPerHost: 200000,
				Proxy: http.ProxyURL(proxyUrl),
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
	_interval := 1.0e9 / _p
	c.throttle = time.Tick(_interval * time.Nanosecond)
	// var wg sync.WaitGroup

	log.Println("run with rps -> ", int(_p))
	go func() {
		i := 0
		for {
			if i == len(rands) {
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

	// do log here, so either db, file or no save only display during load test
	log_intv := time.Tick(time.Duration(logIntv) * time.Second)
	go func() {
		for {
			<-log_intv
			logger.Log(c.counter.GetAllStat(), logIntv)
		}
	}()
}
func (c *Hammer) health(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Content-Length", strconv.Itoa(len("health")))
	rw.WriteHeader(200)
	rw.Write([]byte("health"))
}

func (c *Hammer) log(rw http.ResponseWriter, req *http.Request) {
	p := struct {
		Title string
		Data  string
	}{
		Title: fmt.Sprintf(
			"Performance Log [type:%s rps:%d size:%d slow:%d]",
			profileType,
			rps,
			sessionAmount,
			slowThreshold),
		Data: logger.Read(),
	}
	t, _ := template.ParseFiles("log.tpl")
	t.Execute(rw, p)
}

// init the program from command line
var (
	rps              int64
	profileFile      string
	profileType      string
	slowThreshold    int64
	debug            bool
	auth_method      string
	auth_key         string
	sessionAmount    int
	sessionUrlPrefix string
	proxy            string

	chisel bool

	logIntv int
	logType string

	// profile
	profile scenario.Profile
	logger  logg.Logger

	// rands
	rands []*rand.Rand

	oauth_client = new(auth.Client)
)

func init() {
	flag.Int64Var(&rps, "rps", 500, "Set Request Per Second")
	flag.StringVar(&profileFile, "profile", "", "The path to the traffic profile")
	flag.Int64Var(&slowThreshold, "threshold", 200, "Set slowness standard (in millisecond)")
	flag.StringVar(&profileType, "type", "default", "Profile type (default|session|your session type)")
	flag.BoolVar(&debug, "debug", false, "debug flag (true|false)")
	flag.StringVar(&auth_method, "auth", "none", "Set authorization flag (oauth|gree(c|s)2s|none)")
	flag.StringVar(&auth_key, "key", "", "Set authorization key")
	flag.IntVar(&sessionAmount, "size", 100, "session amount")
	flag.StringVar(&sessionUrlPrefix, "url", "", "Session url prefix")
	flag.StringVar(&proxy, "proxy", "none", "Set HTTP proxy (need to specify scheme. e.g. http://127.0.0.1:8888)")
	flag.IntVar(&logIntv, "intv", 6, "Log interval in chart")
	flag.StringVar(&logType, "ltype", "default", "Log type (file|db)")
	flag.BoolVar(&chisel, "chisel", false, "adjust rps dynamically")

}

// main func
func main() {

	flag.Parse()
	NCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(NCPU)

	// to speed up
	rands = make([]*rand.Rand, NCPU)
	for i, _ := range rands {
		s := rand.NewSource(time.Now().UnixNano())
		rands[i] = rand.New(s)
	}

	log.Println("cpu number -> ", NCPU)
	log.Println("rps -> ", rps)
	log.Println("slow threshold -> ", slowThreshold, "ms")
	log.Println("profile type -> ", profileType)
	log.Println("proxy -> ", proxy)
	log.Println("auth method-> ", auth_method)
	log.Println("auth key -> ", auth_key)

	profile, _ = scenario.New(profileType, sessionAmount)
	if profileFile != "" {
		profile.InitFromFile(profileFile)
	} else {
		profile.InitFromCode(sessionUrlPrefix)
	}

	logger, _ = logg.NewLogger(logType, fmt.Sprintf("%s_%d_%d_%d", profileType, rps, sessionAmount, slowThreshold))

	rand.Seed(time.Now().UnixNano())

	hamm := new(Hammer)
	hamm.Init()

	go hamm.launch(rps)

	http.HandleFunc("/log", hamm.log)
	http.HandleFunc("/health", hamm.health)
	http.ListenAndServe(":9090", nil)

	var input string
	for {
		fmt.Scanln(&input)
		if input == "exit" {
			break
		}
	}
}
