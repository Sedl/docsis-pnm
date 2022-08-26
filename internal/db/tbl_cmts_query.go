package db

import (
	"database/sql"
	"github.com/sedl/docsis-pnm/internal/types"
)

type CMTSQuery struct {
	db   *sql.DB
	rows *sql.Rows
}

func (q *CMTSQuery) Close() error {
	if q.rows != nil {
		return q.rows.Close()
	}
	return nil
}

func NewCMTSQuery(conn *sql.DB, where string, args ...interface{}) (*CMTSQuery, error) {
	mq := &CMTSQuery{conn, nil}

	var query string
	if where != "" {
		query = cmtsQueryStr + " WHERE " + where
	} else {
		query = cmtsQueryStr
	}
	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, err
	}

	mq.rows = rows
	return mq, nil
}

func (q *CMTSQuery) Next() (*types.CMTSRecord, error) {
	if !q.rows.Next() {
		return nil, nil
	}

	var community, commModem sql.NullString

	record := &types.CMTSRecord{}
	err := q.rows.Scan(
		&record.Id,
		&record.Hostname,
		&community,
		&commModem,
		&record.Disabled,
		&record.PollInterval,
		&record.MaxRepetitions)
	if err != nil {
		return nil, err
	}

	if community.Valid {
		record.SNMPCommunity = community.String
	}

	if commModem.Valid {
		record.SNMPModemCommunity = commModem.String
	}

	return record, nil
}
