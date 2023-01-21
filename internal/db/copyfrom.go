package db

import (
	"database/sql"
	"github.com/lib/pq"
	"github.com/sedl/docsis-pnm/internal/logger"
	"strings"
	"time"
)

type CopyFrom struct {
	db             *sql.DB
	tx             *sql.Tx
	Stmt           *sql.Stmt
	query          string
	valueChan      chan []interface{}
	commitInterval time.Duration
}

func NewCopyFrom(query string, dbconn *sql.DB, queueSize int, commitInterval time.Duration) (*CopyFrom, error) {
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
				logger.Fatal(err.Error())
			}
		case data, ok := <-m.valueChan:
			if !ok {
				m.commit()
				return
			}
			_, err := m.Stmt.Exec(data...)
			if err != nil {
				logger.Errorf("CopyFrom insertion failed: %v", m.query)
				if perr, ok := err.(*pq.Error); ok {
					logger.Errorf("Error from Postgres: %v", perr.Code.Name())
				} else {
					logger.Errorf("%#v", err)
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
		logger.Errorf("Error while flushing \"COPY FROM\": %s\n", err)
	}
	err = m.Stmt.Close()
	if err != nil {
		logger.Errorf("Error while closing statement in COPY FROM: %s\n", err)
	}
	err = m.tx.Commit()
	if err != nil {
		logger.Errorf("Error while commiting: %s\n", err)
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
				logger.Errorf("connection to database refused. Retrying in %s. %d retries so far\n", retryAfter, retries)
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
