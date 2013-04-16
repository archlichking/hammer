package counter

import (
	"database/sql"
	"fmt"
	_ "github.com/ziutek/mymysql/godrv"
	"log"
	"sync"
)

const (
	logInsertSQL = `
		INSERT load_log(TOTAL_SEND, TOTAL_REQ, REQ_PS, RES_PS, TOTAL_RES_SLOW, TOTAL_RES_ERR, TOTAL_RES_TIME)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
)

type MysqlC struct {
	table    string
	flushing sync.Mutex
	db       *sql.DB
}

func (m *MysqlC) Init(tn string) {
	m.table = tn
}

func (m *MysqlC) Open(config MysqlConfig) error {
	var err error
	m.db, err = sql.Open("mymysql", fmt.Sprintf("tcp:%s*%s/%s/%s",
		config.Mysql.Host,
		"ll",
		config.Mysql.User,
		config.Mysql.Password))
	if err != nil {
		log.Printf("db transaction initial failed")
		return err
	}
	return nil
}

func (m *MysqlC) Close() {
	m.db.Close()
}

func (m *MysqlC) Flush(c *Counter) error {
	m.flushing.Lock()
	defer m.flushing.Unlock()

	tx, err := m.db.Begin()

	if err != nil {
		log.Printf("db transaction begin failed")
		return err
	}

	_, err = tx.Exec(logInsertSQL, c.totalSend, c.totalReq,
		c.getSendPS(), c.getReqPs(), c.totalResSlow,
		c.totalErr, c.totalResTime)

	if err != nil {
		log.Printf("db insertion failed")
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("db commit failed")
		return err
	}

	return nil
}
