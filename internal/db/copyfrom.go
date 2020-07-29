package db

import (
	"database/sql"
	"github.com/lib/pq"
	"log"
	"strings"
	"time"
)

type CopyFrom struct {
	db *sql.DB
	tx *sql.Tx
	Stmt *sql.Stmt
	query string
	valueChan chan []interface{}
	commitInterval time.Duration
}

func NewCopyFrom (query string, dbconn *sql.DB, queueSize int, commitInterval time.Duration) (*CopyFrom, error) {
	tx, err := dbconn.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &CopyFrom{
		dbconn,
		tx,
		stmt,
		query,
		make(chan []interface{}, queueSize),
		commitInterval,
	}, nil
}

func (m *CopyFrom) Run() {
	timer := time.NewTicker(m.commitInterval).C

	for {
		select {
		case <-timer:
			err := m.Commit()
			if err != nil {
				log.Fatal(err)
			}
			case data, ok := <- m.valueChan:
				if ! ok {
					m.commit()
					return
				}
				_, err := m.Stmt.Exec(data...)
				if err != nil {
					log.Printf("error: CopyFrom insertion failed: %v\n", m.query)
					if perr, ok := err.(*pq.Error); ok {
						log.Printf("error: Error from Postgres: %v\n", perr.Code.Name())
					} else {
						log.Printf("%#v\n", err)
					}
				}
		}
	}
}

func (m *CopyFrom) Stop() {
	close(m.valueChan)
}

func (m *CopyFrom) Insert(values ...interface{}) {
	m.valueChan <- values
}

func (m *CopyFrom) commit() {
	_, err := m.Stmt.Exec()
	if err != nil {
		log.Printf("Error while flushing \"COPY FROM\": %s\n", err)
	}
	err = m.Stmt.Close()
	if err != nil {
		log.Printf("Error while closing statement in COPY FROM: %s\n", err)
	}
	err = m.tx.Commit()
	if err != nil {
		log.Printf("Error while commiting: %s\n", err)
	}
}

func (m *CopyFrom) Commit() error {

    m.commit()

	retries := 0
	retryAfter := time.Second * 5
	var err error

	for {
		// New database
		m.tx, err = m.db.Begin()
		if err != nil {
			if strings.Contains(err.Error(), "connection refused") {
				retries++
				log.Printf("error: connection to database refused. Retrying in %s. %d retries so far\n", retryAfter, retries)
				time.Sleep(retryAfter)
				continue
			}
		} else {
			break
		}
	}

	m.Stmt, err = m.tx.Prepare(m.query)
	if err != nil {
		return err
	}

	return nil
}

