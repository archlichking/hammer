package main

import (
	"client"
	"counter"
	"flag"
	"fmt"
	"html/template"
	"log"
	"logg"
	"math/rand"
	"net/http"
	"runtime"
	"scenario"
	"strconv"
	"time"
)

// to reduce size of thread, speed up
const SizePerThread = 10000000

// Counter will be an atomic, to count the number of request handled
// which will be used to print PPS, etc.
type Hammer struct {
	counter *counter.Counter

	client  client.ClientInterface
	monitor *time.Ticker
	// ideally error should be organized by type TODO
	throttle <-chan time.Time
	// 0 for constant, 1 for flexible
	mode       int
	modeAdjInv int
}

// init
func (c *Hammer) Init(clientType string) {
	switch mode {
	case "constant":
		c.mode = 0
	case "flexible":
		c.mode = 1
	default:
		c.mode = 0
	}
	c.modeAdjInv = 5

	c.counter = new(counter.Counter)
	c.client, _ = client.New(clientType, proxy)
}

// main goroutine to drive traffic
func (c *Hammer) hammer(rg *rand.Rand) {
	// before send out, update send count
	c.counter.RecordSend()
	call, session, cur, err := profile.NextCall(rg)

	if err != nil {
		log.Println("next call error: ", err)
		return
	}
	response_time, err := c.client.Do(call, debug)

	if session != nil {
		// session type so we need to lock for next step
		defer session.LockNext(cur)
	}

	if err != nil {
		if response_time != -1 {
			// only document successful request
			c.counter.RecordError()
		}
		log.Println(err)
	} else {
		c.counter.RecordRes(response_time, slowThreshold)
	}
}

func (c *Hammer) monitorHammer() {
	log.Println(c.counter.GeneralStat(), profile.CustomizedReport())
}

func (c *Hammer) launch(rps int, warmup int) {

	_p := time.Duration(rps)
	_interval := time.Second / _p
	c.throttle = time.Tick(_interval * time.Nanosecond)

	switch c.mode {
	case 0:
		// constant mode, enable warmup
		if warmup != 0 {
			t := 1
			i := int(warmup / 4)
			w := time.Tick(time.Second * time.Duration(i))

			c.throttle = time.Tick(_interval * time.Duration(4) * time.Nanosecond)

			go func() {
				for t < 4 {
					// 4, 3, 2, 1 times _interval for each warmup/4
					<-w
					t += 1
					c.throttle = time.Tick(_interval * time.Duration(int(6/t)) * time.Nanosecond)
				}
			}()
		}
		break
	case 1:
		// no warm up in this mode
		w := time.Tick(time.Second * time.Duration(5))
		var j1 int64 = 0
		go func() {
			for {
				<-w
				j2 := c.counter.GetAllStat()[5]
				if j1 == 0 {
					j1 = j2
				}
				switch {
				case j1 < j2:
					// getting slower
					if j2-j1 > int64(float64(j1)*0.1) {
						_interval = time.Duration(int(float64(_interval.Nanoseconds()) * 1.1))
						c.throttle = time.Tick(_interval)
					}
					break
				case j1 > j2:
					// getting faster
					if j1-j2 > int64(float64(j1)*0.1) {
						_interval = time.Duration(int(float64(_interval.Nanoseconds()) * 0.9))
						c.throttle = time.Tick(_interval)
					}
					break
				default:
					break
				}

				j1 = j2
			}
		}()
		break
	}

	go func() {
		for {
			i := rand.Intn(len(rands))
			<-c.throttle

			go c.hammer(rands[i])
		}
	}()

	c.monitor = time.NewTicker(time.Second)
	go func() {
		for {
			// rate limit for monitor routine
			<-c.monitor.C
			go c.monitorHammer()
		}
	}()

	// do log here
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
			"Performance Log [rps:%d]",
			rps),
		Data: logger.Read(),
	}
	t, _ := template.ParseFiles("log.tpl")
	t.Execute(rw, p)
}

// init the program from command line
var (
	rps           int
	profileFile   string
	slowThreshold int64
	debug         bool
	proxy         string
	mode          string
	duration      int
	warmup        int

	logIntv int
	logType string

	// profile
	profile *scenario.Profile
	logger  logg.Logger

	// rands
	rands []*rand.Rand
)

func init() {
	flag.IntVar(&rps, "r", 500, "Request # per second")
	flag.StringVar(&profileFile, "p", "", "Profile json file path (required)")
	flag.Int64Var(&slowThreshold, "t", 200, "Threshold for slow response in ms")
	flag.BoolVar(&debug, "D", false, "Debug flag (true|false)")
	flag.StringVar(&proxy, "P", "nil", "Http proxy (e.g. http://127.0.0.1:8888)")
	flag.IntVar(&logIntv, "i", 6, "Log interval")
	flag.StringVar(&mode, "m", "constant", "Load generate mode (constant|flexible)")
	flag.StringVar(&logType, "l", "default", "Log type (file|db)")
	flag.IntVar(&duration, "d", 0, "Test duration, infinite by default")
	flag.IntVar(&warmup, "w", 0, "Test wrapup duration, infinite by default")

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

	log.Println("cpu #      ->", NCPU)
	log.Println("profile    ->", profileFile)
	log.Println("rps        ->", rps)
	log.Println("slow req   ->", slowThreshold, "ms")
	log.Println("proxy      ->", proxy)
	log.Println("mode       ->", mode)
	log.Println("duration   ->", duration, "s")
	log.Println("wrapup     ->", warmup, "s")

	profile, _ = scenario.New(profileFile)

	logger, _ = logg.NewLogger(logType, fmt.Sprintf("%d_%d", rps, slowThreshold))

	rand.Seed(time.Now().UnixNano())

	hamm := new(Hammer)
	hamm.Init(profile.Client)

	go hamm.launch(rps, warmup)

	if duration != 0 {
		timer := time.NewTimer(time.Second * time.Duration(duration))
		<-timer.C
	} else {
		var input string
		for {
			// block exiting
			fmt.Scanln(&input)
			if input == "exit" {
				break
			}
		}
	}

	// http.HandleFunc("/log", hamm.log)
	// http.HandleFunc("/health", hamm.health)
	// http.ListenAndServe(":9090", nil)
}
