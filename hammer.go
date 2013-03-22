package main

import (
	"flag"
	"fmt"
	"github.com/hammer/scenario"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// to reduce size of thread, speed up
const SizePerThread = 10000000

//var DefaultTransport RoundTripper = &Transport{Proxy: ProxyFromEnvironment}

// Counter will be an atomic, to count the number of request handled
// which will be used to print PPS, etc.
type Counter struct {
	totalReq     int64 // total # of request
	totalResTime int64 // total response time
	totalErr     int64 // how many error
	totalResSlow int64 // how many slow response
	totalSend    int64

	lastSend int64
	lastReq  int64

	client  *http.Client
	monitor *time.Ticker
	// ideally error should be organized by type TODO
	throttle <-chan time.Time

	// profile
	profile *scenario.Profile
}

// init
func (c *Counter) Init(p *scenario.Profile) {

	c.client = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:   false,
			MaxIdleConnsPerHost: 2000,
		},
	}

	c.profile = p
}

// increase the count and record response time.
func (c *Counter) recordRes(_time int64) {
	atomic.AddInt64(&c.totalReq, 1)
	atomic.AddInt64(&c.totalResTime, _time)

	// if longer that 200ms, it is a slow response
	if _time > slowThreshold*1000000 {
		atomic.AddInt64(&c.totalResSlow, 1)
		log.Println("slow response -> ", float64(_time)/1.0e9)
	}
}

func (c *Counter) recordError() {
	atomic.AddInt64(&c.totalErr, 1)
}

func (c *Counter) recordSend() {
	atomic.AddInt64(&c.totalSend, 1)
}

// main goroutine to drive traffic
func (c *Counter) hammer(call *scenario.Call) {
	var req *http.Request
	var err error

	t1 := time.Now().UnixNano()

	// before send out, update send count
	c.recordSend()

	req, err = http.NewRequest(call.Method, call.URL, strings.NewReader(call.Body))

	// Add special haeader for PATCH, PUT and POST
	switch call.Method {
	case "PATCH", "PUT", "POST":
		switch call.Type {
		case "REST":
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
		case "WWW":
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	res, err := c.client.Do(req)

	response_time := time.Now().UnixNano() - t1

	if err != nil {
		log.Println("Response Time: ", float64(response_time)/1.0e9, " Erorr: when", call.Method, call.URL, "with error ", err)
		c.recordError()
		return
	}

	/*
		    ###
			disable reading res.body, no need for our purpose for now,
		    by doing this, hope we can save more file descriptor.
			##
	*/
	defer req.Body.Close()
	defer res.Body.Close()

	// check response code here
	// 409 conflict is ok for PATCH request

	if res.StatusCode >= 400 && res.StatusCode != 409 {
		//fmt.Println(res.Status, string(data))
		log.Println("Got error code --> ", res.Status, "for call ", call.Method, " ", call.URL)
		c.recordError()
		return
	}

	// reference --> https://github.com/tenntenn/gae-go-testing/blob/master/recorder_test.go

	// only record time for "good" call
	c.recordRes(response_time)
}

func (c *Counter) monitorHammer() {
	sps := c.totalSend - c.lastSend
	pps := c.totalReq - c.lastReq
	backlog := c.totalSend - c.totalReq - c.totalErr

	atomic.StoreInt64(&c.lastReq, c.totalReq)
	atomic.StoreInt64(&c.lastSend, c.totalSend)

	avgT := float64(c.totalResTime) / (float64(c.totalReq) * 1.0e9)

	log.Println(
		" total: ", fmt.Sprintf("%4d", c.totalSend),
		" req/s: ", fmt.Sprintf("%4d", sps),
		" res/s: ", fmt.Sprintf("%4d", pps),
		" avg: ", fmt.Sprintf("%2.4f", avgT),
		" pending: ", backlog,
		" err:", c.totalErr,
		"|", fmt.Sprintf("%2.2f%s", (float64(c.totalErr)*100.0/float64(c.totalErr+c.totalReq)), "%"),
		" slow: ", fmt.Sprintf("%2.2f%s", (float64(c.totalResSlow)*100.0/float64(c.totalResSlow+c.totalReq)), "%"))
}

func (c *Counter) launch(rps int64) {
	// var _rps time.Duration

	_p := time.Duration(rps)
	_interval := 1000000000.0 / _p
	c.throttle = time.Tick(_interval * time.Nanosecond)
	var wg sync.WaitGroup

	log.Println("run with rps -> ", int(_p))
	go func() {
		for {
			calls, gap := (*c.profile).NextCalls()
			if gap != -1 {
				// session scenario, need to promise [0:gap-1]*calls executes first
				for i := 0; i < gap; i++ {
					<-c.throttle
					// c.hammer(calls[i])
					wg.Add(1)
					go func() {
						c.hammer(calls[i])
						wg.Done()
					}()
				}
				wg.Wait()

				// then with ticker
				for i := gap; i < len(calls); i++ {
					<-c.throttle
					// threshold
					go c.hammer(calls[i])
				}
			} else {
				// default scenario, launch only with ticker
				for i := 0; i < len(calls); i++ {
					<-c.throttle
					// threshold
					go c.hammer(calls[i])
				}
			}

		}
	}()

	c.monitor = time.NewTicker(time.Second)
	go func() {
		for {
			<-c.monitor.C // rate limit for monitor routine
			go c.monitorHammer()
		}
	}()
}

// init the program from command line
var (
	rps           int64
	profileFile   string
	profileType   string
	slowThreshold int64
)

func init() {
	flag.Int64Var(&rps, "rps", 500, "Set Request Per Second")
	flag.StringVar(&profileFile, "profile", "", "The path to the traffic profile")
	flag.Int64Var(&slowThreshold, "threshold", 200, "Set slowness standard (in millisecond)")
	flag.StringVar(&profileType, "type", "default", "profile type: default or session")
}

// main func
func main() {
	flag.Parse()
	NCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(NCPU + 3)

	log.Println("cpu number -> ", NCPU)
	log.Println("rps -> ", rps)
	log.Println("slow threshold -> ", slowThreshold, "ms")
	log.Println("profile type -> ", profileType)

	profile, _ := scenario.New(profileType)
	if profileFile != "" {
		profile.InitFromFile(profileFile)
	} else {
		profile.InitFromCode()
	}

	rand.Seed(time.Now().UnixNano())

	counter := new(Counter)
	counter.Init(&profile)

	go counter.launch(rps)

	var input string
	fmt.Scanln(&input)
}
