package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	// "sync"
	"sync/atomic"
	"time"
	"github.com/hammer/scenario"
)

// to reduce size of thread, speed up
const SizePerThread = 10000000

//var DefaultTransport RoundTripper = &Transport{Proxy: ProxyFromEnvironment}

// Counter will be an atomic, to count the number of request handled
// which will be used to print PPS, etc.
type Counter struct {
	lasttime      int64 // time of last print, in secon, to calculate RPS
	lastcount     int64 // count of last print
	lasttotaltime int64 // last total response time

	count     int64 // total # of request
	totaltime int64 // total response time

	totalerrors int64 // how many error

	totalslowresp int64 // how many slow response

	// to calculate send count
	s_lasttime  int64
	s_lastcount int64
	s_count     int64

	// book keeping just for faster stats report so we do not do it again
	avg_time      float64
	last_avg_time float64
	backlog       int64

	client (*http.Client)

	monitor (*time.Ticker)

	// ideally error should be organized by type TODO
	throttle <-chan time.Time

	runinfo <-chan bool // to indicate current run is good or bad

	// auto find pps
	currentRPS  time.Duration
	lastGoodRPS time.Duration
	lastBadRPS  time.Duration

	// profile
	profile *scenario.Scenario
}

// var TrafficProfile = new(trafficprofiles.Profile)
var _DEBUG bool
var _HOST string

// init
func (c *Counter) _init(p *scenario.Scenario) {
	tr := &http.Transport{
		DisableKeepAlives: false,
		MaxIdleConnsPerHost: 2000,
	}
	// init http client
	c.client = &http.Client{Transport: tr}

	// make channel for auto finder mode
	c.runinfo = make(chan bool)

	c.profile = p

	c.monitor = time.NewTicker(time.Second)
	go func() {
		for {
			<-c.monitor.C // rate limit for monitor routine
			go c.pperf()
		}
	}()
}

// increase the count and record response time.
func (c *Counter) record(_time int64) {
	atomic.AddInt64(&c.count, 1)
	atomic.AddInt64(&c.totaltime, _time)

	// if longer that 200ms, it is a slow response
	if _time > slownessLimit * 1000000 {
		atomic.AddInt64(&c.totalslowresp, 1)
		log.Println("Slow response -> ", float64(_time)/1.0e9)
	}
}

// when error happened, increase counter
// TODO: maybe add error type later
func (c *Counter) recordError() {
	atomic.AddInt64(&c.totalerrors, 1)

	// we do not record time for errors.
	// and there will not be count incr for calls as well
}

func (c *Counter) recordSend() {
	atomic.AddInt64(&c.s_count, 1)
}

// main goroutine to drive traffic
func (c *Counter) hammer() {
	var req *http.Request
	var err error

	t1 := time.Now().UnixNano()

	// before send out, update send count
	c.recordSend()

	call := (*c.profile).NextCall()
	req, err = http.NewRequest(call.Method, call.URL, strings.NewReader(call.Body))

	// Add special haeader for PATCH, PUT and POST
	if call.Method == "PATCH" || call.Method == "PUT" || call.Method == "POST" {
		if call.Type == "REST" {
			// _params.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
		} else if call.Type == "WWW" {
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
	c.record(response_time)
	call.Record(response_time)
}

// to print out performance counter
// run every second, will also update last count
func (c *Counter) pperf() {
	sps := c.s_count - c.s_lastcount
	pps := c.count - c.lastcount
	c.backlog = c.s_count - c.count - c.totalerrors

	atomic.StoreInt64(&c.lastcount, c.count)
	atomic.StoreInt64(&c.s_lastcount, c.s_count)

	c.avg_time = float64(c.totaltime) / (float64(c.count) * 1.0e9)
	// c.last_avg_time = TODO!!

/*
	log.Println(" SendPS: ", fmt.Sprintf("%4d", sps),
		" ReceivePS: ", fmt.Sprintf("%4d", pps), fmt.Sprintf("%2.4f", c.avg_time),
		" Pending Requests: ", c.backlog,
		" Error:", c.totalerrors,
		"|", fmt.Sprintf("%2.2f%s", (float64(c.totalerrors)*100.0/float64(c.totalerrors+c.count)), "%"),
		" Slow Ratio: ", fmt.Sprintf("%2.2f%s", (float64(c.totalslowresp)*100.0/float64(c.totalerrors+c.count)), "%"))
*/
	log.Println(
    " Count: ", fmt.Sprintf("%4d", c.s_count),
    " SendPS: ", fmt.Sprintf("%4d", sps),
		" ReceivePS: ", fmt.Sprintf("%4d", pps), fmt.Sprintf("%2.4f", c.avg_time),
		" Pending Requests: ", c.backlog,
		" Error:", c.totalerrors,
		"|", fmt.Sprintf("%2.2f%s", (float64(c.totalerrors)*100.0/float64(c.totalerrors+c.count)), "%"),
		" Slow Ratio: ", fmt.Sprintf("%2.2f%s", (float64(c.totalslowresp)*100.0/float64(c.totalerrors+c.count)), "%"))
}

// routine to return status
func (c *Counter) stats(res http.ResponseWriter, req *http.Request) {
	res.Header().Set(
		"Content-Type", "text/plain",
	)
	io.WriteString(
		res,
		fmt.Sprintf("Total Request: %d\nTotal Error: %d\n==========\n%s",
			c.count, c.totalerrors, string((*c.profile).Print())),
	)
}

func index(res http.ResponseWriter, req *http.Request) {
	res.Header().Set(
		"Content-Type",
		"text/html",
	)
	io.WriteString(
		res,
		`test`,
	)
}

func (c *Counter) run_once(pps time.Duration) {
	_interval := 1000000000.0 / pps
	/*
			_send_per_tick := 1
		  	if pps > 400 {
				_send_per_tick = 5
				_interval = 1000000000.0 * 5 / pps
				log.Println("dount the per tick sending...")
		  	}
	*/
	log.Println(_interval)
	c.throttle = time.Tick(_interval * time.Nanosecond)

	// fmt.println _users

	go func() {
		for {
			<-c.throttle // rate limit our Service.Method RPCs
			go c.hammer()
			/*
			   if _send_per_tick > 1 {
			     // send two per tick for very high RPS to be more accurate
			     go c.hammer()
			     go c.hammer()
			     go c.hammer()
			     go c.hammer()
			   }
			*/
		}
	}()
}

func (c *Counter) findPPS(_p int64) {
	var _rps time.Duration

	_rps = time.Duration(_p)
	// already a gorouting, we just do a infinity loop to find the best RPS
	for {
		c.run_once(_rps)
		log.Println("Run RPS -> ", int(_rps))
		_result := <-c.runinfo
		log.Println(_result)

		// now we know pass of failed, we can start adjust _rps
		if _result {
			// first, we want make sure we can exit the run, that is
			// if the good and failed RPS is within 5 RPS (will change
			// to 5% later), we can assume we found what we are looking for
			c.lastGoodRPS = _rps
			if (c.lastGoodRPS*c.lastBadRPS > 0) && (c.lastGoodRPS-c.lastBadRPS < 5) {
				log.Println("found it!", _rps)
				// additional report and then quit the process
			}
		} else {
			c.lastBadRPS = _rps
		}
		// not found, keep running, next RPS will be (good + bad ) / 2
		if c.lastBadRPS == 0 {
			_rps = c.lastGoodRPS * 2
		} else if c.lastGoodRPS == 0 {
			_rps = c.lastGoodRPS / 2
		} else {
			_rps = (c.lastGoodRPS + c.lastBadRPS) / 2
		}
	}
}

// init the program from command line
var initRPS int64
var profileFile string
var slownessLimit int64

func init() {
	flag.Int64Var(&initRPS, "rps", 500, "Set Request Per Second")
	flag.StringVar(&profileFile, "profile", "", "The path to the traffic profile")
	flag.Int64Var(&slownessLimit, "slowness", 200, "Set slowness standard (in millisecond)")
}

func InitFromCode(s *scenario.Scenario) {
	s.AddNewCall(10, "GET", "WWW", func()(string, string){
		return "http://localhost:9000/hello", "{}"
		})
	s.AddNewCall(90, "GET", "WWW", func()(string, string){
		return "http://localhost:9000/hello", "{}"
		})
}

// main func
func main() {

	NCPU := runtime.NumCPU()
	log.Println("# of CPU is ", NCPU)

	runtime.GOMAXPROCS(NCPU + 3)

	flag.Parse()
	log.Println("RPS is", initRPS)
	log.Println("Slowness cap is", slownessLimit, "ms")

	profile := scenario.New()
	profile.InitFromCode = InitFromCode
	if profileFile != "" {
		profile.InitFromFile(profileFile)
	} else {
		profile.InitFromCode(profile)
	}

	rand.Seed(time.Now().UnixNano())

	counter := new(Counter)
	counter._init(profile)
	
	go counter.findPPS(initRPS)

	var input string
	fmt.Scanln(&input)
}
