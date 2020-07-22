package db

import (
	"database/sql"
	"fmt"
	"github.com/sedl/docsis-pnm/internal/types"
	"net"
)

const modemHistoryQuery = `
SELECT
    d.poll_time,
    d.freq,
    d.power,
    d.snr,
    d.microrefl,
    d.unerroreds,
    d.correcteds,
    d.erroreds,
    d.modulation 
FROM
    modem_downstream AS d 
JOIN
    modem 
        ON (
            modem.id = d.modem_id
        ) 
WHERE
      modem.mac = $1 %s
--     AND d.poll_time >= $1 
--     AND d.poll_time <= $2 
--     AND modem.mac = $3 
ORDER BY
    d.poll_time`



type DownstreamHistoryQuery struct {
	db *sql.DB
	rows *sql.Rows
}

func (q *DownstreamHistoryQuery) Close() error {
	if q.rows != nil {
		return q.rows.Close()
	}
	return nil
}

func (q *DownstreamHistoryQuery) Next() (ts int64, ds *types.DownstreamChannel, err error) {
	if !q.rows.Next() {
		return 0, nil, nil
	}

	// var tstamp int64

	ds = &types.DownstreamChannel{}

	err = q.rows.Scan(
		&ts,
		&ds.Freq,
		&ds.Power,
		&ds.SNR,
		&ds.Microrefl,
		&ds.Unerroreds,
		&ds.Correcteds,
		&ds.Uncorrectables,
		&ds.Modulation,
	)
	if err != nil {
		return 0, nil, err
	}

	return
}

func NewDownstreamQuery(conn *sql.DB, mac *net.HardwareAddr, where string, args... interface{}) (*DownstreamHistoryQuery, error){
	mq := &DownstreamHistoryQuery{conn, nil}

	if where != "" {
		where = " AND " + where
	}

	query := fmt.Sprintf(modemHistoryQuery, where)

	argsn := append([]interface{}{mac.String()}, args...)
	rows, err := conn.Query(query, argsn...)
	if err != nil {
		return nil, err
	}

	mq.rows = rows
	return mq, nil
}

// GetDownstreamHistory returns the modem Downstream history. You can pass it an additional WHERE clause but keep in
// mind that the first placeholder you can use is $2
func (db *Postgres) GetDownstreamHistory(mac net.HardwareAddr, where string, args ...interface{}) ([]*types.DownstreamChannelHistory, error) {

	conn, err := db.GetConn()
	if err != nil {
		return nil, err
	}

	if where != "" {
		where = " AND " + where
	}

	query := fmt.Sprintf(modemHistoryQuery, where)

	argsn := append([]interface{}{mac.String()}, args...)
	rows, err := conn.Query(query, argsn...)
	if err != nil {
		return nil, err
	}

	defer CloseOrLog(rows)


	dsdata := make([]*types.DownstreamChannelHistory, 0)

	var lasttime int64 = 0
	var tstamp int64

	var dschan *types.DownstreamChannelHistory = nil

	for rows.Next() {
		ds := &types.DownstreamChannel{}
		err := rows.Scan(
			&tstamp,
			&ds.Freq,
			&ds.Power,
			&ds.SNR,
			&ds.Microrefl,
			&ds.Unerroreds,
			&ds.Correcteds,
			&ds.Uncorrectables,
			&ds.Modulation,
		)
		if err != nil {
			return nil, err
		}
		if dschan == nil || lasttime != tstamp {
			if dschan != nil {
				dsdata = append(dsdata, dschan)
			}
			dschan = &types.DownstreamChannelHistory{
				Timestamp:            tstamp,
				Downstreams:          make([]*types.DownstreamChannel, 0),
			}
			lasttime = tstamp
		}
		dschan.Downstreams = append(dschan.Downstreams, ds)
	}

	return dsdata, nil
}
