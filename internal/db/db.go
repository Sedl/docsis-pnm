package db

import (
	"database/sql"
	"fmt"
	"github.com/sedl/docsis-pnm/internal/logger"
	"github.com/sedl/docsis-pnm/internal/migration"
	"github.com/sedl/docsis-pnm/internal/types"
	"strings"
	"sync"
	"time"
)

type Postgres struct {
	conn          *sql.DB
	timeout       time.Duration
	config        *types.Db
	modemDataChan chan *types.ModemData
	connlock      sync.RWMutex
}

// NewPostgres creates a new Postgres struct
func NewPostgres(cfg types.Db) (*Postgres, error) {

	db := &Postgres{
		timeout:       time.Second * 10,
		modemDataChan: make(chan *types.ModemData, 10),
		config:        &cfg,
	}
	return db, nil
}

// Run starts all necessary goroutines
func (db *Postgres) Run() {
	// TODO make this configurable
	go migration.GoMaintenanceWorker(db, time.Minute*30)
}

func (db *Postgres) GetConn() (dbc *sql.DB, err error) {

	db.connlock.RLock()
	conn := db.conn
	db.connlock.RUnlock()

	if conn != nil {
		return conn, nil
	}
	db.connlock.Lock()
	defer db.connlock.Unlock()

	// TODO DB reconnect: check if we have to do this manually or the driver is handling it already
	for i := 0; i < 20; i++ {
		db.conn, err = sql.Open("postgres", db.config.Connstr)
		if err != nil {
			logger.Errorf("Connection error. Retrying in %s: %s", db.timeout, err)
			time.Sleep(db.timeout)
			db.conn = nil
			continue
		}
		db.conn.SetMaxIdleConns(10)
		return db.conn, nil
	}
	return nil, nil
}

func tableUpdate(conn *sql.DB, updateQuery string, whereValue interface{}, changes map[string]interface{}) error {

	i := 2

	cols := make([]string, 0)
	// add 1 to make room for the id
	values := make([]interface{}, 0)

	values = append(values, whereValue)

	for col, value := range changes {
		cols = append(cols, fmt.Sprintf("%s = $%d", col, i))
		values = append(values, value)
		i++
	}

	if len(changes) == 0 {
		return nil
	}

	colstr := strings.Join(cols, ", ")
	updateStr := fmt.Sprintf(updateQuery, colstr)
	_, err := conn.Exec(updateStr, values...)

	return err
}
