package counter

import (
	"database/sql"
	"fmt"
	_ "github.com/ziutek/mymysql/godrv"
	"log"
	"sync"
)

type MysqlC struct {
	tableName string
	flushing  sync.Mutex
	db        *sql.DB
}

func (m *MysqlC) Init(tn string) {
	m.tableName = tn
}

func (m *MysqlC) Connect(config MysqlConfig) error {
	var err error
	m.db, err = sql.Open("mymysql", fmt.Sprintf("%s/%s/%s",
		// config.Mysql.Host,
		"ll",
		config.Mysql.User,
		config.Mysql.Password))
	if err != nil {
		log.Printf("db transaction initial failed", err)
		return err
	}
	return nil
}

func (m *MysqlC) CreateEmptyTable() error {
	_, err := m.db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(
		ID       int(32) NOT NULL AUTO_INCREMENT,
		TOTAL_SEND int(32) NOT NULL,
		TOTAL_REQ int(32) NOT NULL,
		REQ_PS int(32) NOT NULL,
		RES_PS int(32) NOT NULL,
		TOTAL_RES_SLOW int(32) NOT NULL,
		TOTAL_RES_ERR      int(32) NOT NULL,
		TOTAL_RES_TIME int(32) NOT NULL,
		TIME_CREATED TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (ID)
		) ENGINE=MyISAM DEFAULT CHARSET=utf8;
		`, m.tableName))
	if err != nil {
		log.Printf("db table create failed", err)
		return err
	}

	_, err = m.db.Exec(fmt.Sprintf(`DELETE FROM %s`, m.tableName))
	if err != nil {
		log.Printf("db table clear failed", err)
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
	log.Println(fmt.Sprintf(`INSERT %s(TOTAL_SEND, TOTAL_REQ, REQ_PS, RES_PS, TOTAL_RES_SLOW, TOTAL_RES_ERR, TOTAL_RES_TIME)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, m.tableName))

	_, err = tx.Exec(fmt.Sprintf(`INSERT %s(TOTAL_SEND, TOTAL_REQ, REQ_PS, RES_PS, TOTAL_RES_SLOW, TOTAL_RES_ERR, TOTAL_RES_TIME)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, m.tableName),
		c.totalSend,
		c.totalReq,
		c.getSendPS(),
		c.getReqPs(),
		c.totalResSlow,
		c.totalErr,
		c.totalResTime)

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
