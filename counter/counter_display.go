package counter

import (
	"time"
)

type CounterDisplay struct {
	ID             int
	TOTAL_SEND     int
	TOTAL_REQ      int
	REQ_PS         int
	RES_PS         int
	TOTAL_RES_SLOW int
	TOTAL_RES_ERR  int
	TOTAL_RES_TIME int
	TIME_CREATED   time.Time
}
