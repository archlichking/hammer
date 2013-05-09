package logg

import (
	"errors"
	"fmt"
	"time"
)

type Logger interface {
	Read() string
	Log([]int64, int)
}

var loggers = make(map[string]func(string) (Logger, error))

func RegisterLogger(name string, logger func(string) (Logger, error)) {
	loggers[name] = logger
}

func NewLogger(loggerName string, logname string) (Logger, error) {
	if logger, ok := loggers[loggerName]; ok {
		return logger(logname)
	}

	return nil, errors.New("logger is not registered")
}

type DefaultLogger struct {
	_stats   [][]int64
	_tmpStat []int64
	/*
		'Request/Sec', 'Response/Sec', 'Errors', 'Slow Response', 'Pending Request', 'Avg Response/Nano Sec'
	*/
}

func (l *DefaultLogger) AbstractLog(intervalStat []int64, logInv int) {
	/*
		intervalStat: c.totalReq,
				c.totalErr,
				c.totalResSlow,
				c.sps,
				c.rps,
				c.totalResTime / (c.totalReq * 1.0e3),
				c.totalSend,

		_tmpStat:c.totalReq,
				c.totalErr,
				c.totalResSlow,
				c.sps,
				c.rps,
				c.totalResTime / (c.totalReq * 1.0e3),
				pending,
				time
	*/
	l._tmpStat[0] = intervalStat[0] - l._tmpStat[0]
	l._tmpStat[1] = intervalStat[1] - l._tmpStat[1]
	l._tmpStat[3] = intervalStat[3]
	l._tmpStat[4] = intervalStat[4]
	l._tmpStat[5] = intervalStat[5]
	l._tmpStat[6] = intervalStat[6] - intervalStat[0] - intervalStat[1]
	l._tmpStat[7] = time.Now().Local().Unix()

	l._stats = append(l._stats, []int64{
		l._tmpStat[0],
		l._tmpStat[1],
		(intervalStat[2] - l._tmpStat[2]) / int64(logInv),
		l._tmpStat[3],
		l._tmpStat[4],
		l._tmpStat[5],
		l._tmpStat[6],
		l._tmpStat[7],
	})

	l._tmpStat[2] = intervalStat[2]
}

func (l *DefaultLogger) Log(intervalStat []int64, logInv int) {
	l.AbstractLog(intervalStat, logInv)
}

func (l *DefaultLogger) AbstractRead() string {
	str := "[['Time', 'Request/Sec', 'Response/Sec', 'Errors', 'Slow Response', 'Pending Request', 'Avg Response/Nano Sec'],"
	for _, v := range l._stats {
		t := time.Unix(v[7], 0)
		str += fmt.Sprintf("['%s', %d, %d, %d, %d, %d, %d],",
			fmt.Sprintf("%d:%d:%d", t.Hour(), t.Minute(), t.Second()),
			v[3],
			v[4],
			v[1],
			v[2],
			v[6],
			v[5])
	}

	str = str[:len(str)-1] + "]"

	return str
}

func (l *DefaultLogger) Read() string {
	return l.AbstractRead()
}

func init() {
	RegisterLogger("default", newDefaultLogger)
}

func newDefaultLogger(logname string) (Logger, error) {
	return &DefaultLogger{
		_stats:   make([][]int64, 0),
		_tmpStat: []int64{0, 0, 0, 0, 0, 0, 0, 0},
	}, nil
}
