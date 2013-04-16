package counter

import (
	"fmt"
	"runtime"
	"sync/atomic"
)

type Counter struct {
	totalReq     int64 // total # of request
	totalResTime int64 // total response time
	totalErr     int64 // how many error
	totalResSlow int64 // how many slow response
	totalSend    int64

	lastSend int64
	lastReq  int64
	sps      int64
	rps      int64
}

// increase the count and record response time.
func (c *Counter) RecordRes(_time int64, slowThreshold int64, method string) {
	atomic.AddInt64(&c.totalReq, 1)
	atomic.AddInt64(&c.totalResTime, _time)

	// if longer that 200ms, it is a slow response
	if _time > slowThreshold*1000000 {
		atomic.AddInt64(&c.totalResSlow, 1)
		// log.Println("slow response -> ", float64(_time)/1.0e9, method)
	}
}

func (c *Counter) RecordError() {
	atomic.AddInt64(&c.totalErr, 1)
}

func (c *Counter) RecordSend() {
	atomic.AddInt64(&c.totalSend, 1)
}

func (c *Counter) getSendPS() int64 {
	return c.sps
}

func (c *Counter) getReqPs() int64 {
	return c.rps
}

func (c *Counter) GetBacklog() int64 {
	return c.totalSend - c.totalReq - c.totalErr
}

func (c *Counter) GeneralStat() string {
	c.sps = c.totalSend - c.lastSend
	c.rps = c.totalReq - c.lastReq
	avgT := float64(c.totalResTime) / (float64(c.totalReq) * 1.0e9)

	atomic.StoreInt64(&c.lastSend, c.totalSend)
	atomic.StoreInt64(&c.lastReq, c.totalReq)

	return fmt.Sprintf(" total: %s req/s: %s res/s: %s avg: %s pending: %d err: %d|%s slow: %s rg: %d",
		fmt.Sprintf("%4d", c.totalSend),
		fmt.Sprintf("%4d", c.sps),
		fmt.Sprintf("%4d", c.rps),
		fmt.Sprintf("%2.4f", avgT),
		c.GetBacklog(),
		c.totalErr,
		fmt.Sprintf("%2.2f%s", (float64(c.totalErr)*100.0/float64(c.totalErr+c.totalReq)), "%"),
		fmt.Sprintf("%2.2f%s", (float64(c.totalResSlow)*100.0/float64(c.totalReq)), "%"),
		runtime.NumGoroutine())
}
