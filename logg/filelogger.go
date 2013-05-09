package logg

import (
	"fmt"
	"os"
	"time"
)

type FileLogger struct {
	DefaultLogger
	_statsFile *os.File
	_statsName string
}

func (l *FileLogger) Log(intervalStat []int64, logInv int) {

	if l._statsFile == nil {
		l._stats = make([][]int64, 0)
		l._tmpStat = []int64{0, 0, 0, 0, 0, 0, 0, 0}
		l._statsFile, _ = os.Create(l._statsName)
	}
	l.AbstractLog(intervalStat, logInv)

	t := time.Unix(l._stats[len(l._stats)-1][7], 0)
	str := fmt.Sprintf("['%s', %d, %d, %d, %d, %d, %d],",
		fmt.Sprintf("%d:%d:%d", t.Hour(), t.Minute(), t.Second()),
		l._stats[len(l._stats)-1][3],
		l._stats[len(l._stats)-1][4],
		l._stats[len(l._stats)-1][1],
		l._stats[len(l._stats)-1][2],
		l._stats[len(l._stats)-1][6],
		l._stats[len(l._stats)-1][5])
	l._statsFile.Write([]byte(str))
}

func (l *FileLogger) Read() string {
	return l.AbstractRead()
}

func init() {
	RegisterLogger("file", newFileLogger)
}

func newFileLogger(logname string) (Logger, error) {
	return &FileLogger{
		_statsFile: nil,
		_statsName: logname,
	}, nil
}
