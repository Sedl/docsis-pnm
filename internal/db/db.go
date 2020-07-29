package db

import (
	"database/sql"
	"fmt"
	"github.com/sedl/docsis-pnm/internal/config"
	"github.com/sedl/docsis-pnm/internal/types"
	"io"
	"log"
	"strings"
	"sync"
	"time"
)

type ModemUpstreamCMTSRecord struct {
	ID int
	PollTime int64
	UpstreamID int32
	PowerRx int32
	Status int32
	Unerroreds uint64
	Correcteds uint64
	Erroreds uint64
}

type ModemUpstreamRecord struct {
	ModemID int64
	PollTime int64
	Freq int32
	Modulation uint32
	TimingOffset uint32
}


type Postgres struct {
	conn    *sql.DB
	timeout time.Duration
	config *config.Db

	cacheCMTS *CMTSCache
	modemUpstreamChan chan *ModemUpstreamRecord
	modemDataChan chan *types.ModemData
	connlock   sync.RWMutex
}


// NewPostgres creates a new Postgres struct
func NewPostgres(cfg config.Db) (*Postgres, error) {

	db := &Postgres{
		timeout:           time.Second * 10,
		cacheCMTS:         NewCMTSCache(),
		modemUpstreamChan: make(chan *ModemUpstreamRecord, 100),
		modemDataChan:     make(chan *types.ModemData, 10),
		config:            &cfg,
		// cacheModem:        NewModemCache(),
	}
	return db, nil
}

// Run starts all necessary goroutines
func (db *Postgres) Run() {
	// TODO make this configurable
	go db.GoMaintenanceWorker(time.Minute * 30)
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
			log.Printf("Connection error. Retrying in %s: %s\n", db.timeout, err)
			time.Sleep(db.timeout)
			db.conn = nil
			continue
		}
		db.conn.SetMaxIdleConns(10)
		return db.conn, nil
	}
	return nil, nil
}

func CloseOrLog(intf io.Closer) {

	switch intf.(type) {
	case *sql.DB:
		err := (intf).(*sql.DB).Close()
		if err != nil {
			log.Println(err)
		}
	case *sql.Rows:
		err := (intf).(*sql.Rows).Close()
		if err != nil {
			log.Println(err)
		}
	default:
		err := intf.Close()
		if err != nil {
			log.Println(err)
		}
	}
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
