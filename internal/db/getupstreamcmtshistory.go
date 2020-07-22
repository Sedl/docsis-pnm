package db

import (
	"database/sql"
	"github.com/sedl/docsis-pnm/internal/types"
	"net"
)

const upstreamCMTSHistory = `
SELECT
    u.poll_time,
    u.power_rx,
    u.status,
    u.unerroreds,
    u.correcteds,
    u.erroreds 
FROM
    modem_upstream_cmts AS u 
JOIN
    modem 
        ON (
            modem.id = u.modem_id
        ) 
WHERE
    u.poll_time BETWEEN $1 AND $2 
    AND modem.mac = $3 
ORDER BY
    u.poll_time ASC 
`


func row2UPstreamCMTS(rows *sql.Rows) (*types.UpstreamCMTS, error) {
	us := &types.UpstreamCMTS{}
	err := rows.Scan(
		&us.Timestamp,
		&us.PowerRx,
		&us.Status,
		&us.Unerroreds,
		&us.Correcteds,
		&us.Uncorrectables,
	)
	if err != nil {
		return nil, err
	}

	return us, nil
}

func (db *Postgres) GetCMTSModemUpstream(mac net.HardwareAddr, from, to int) (*[]*types.UpstreamCMTS, error) {

	conn, err := db.GetConn()
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(upstreamCMTSHistory)
	if err != nil {
		return nil, err

	}

	defer CloseOrLog(rows)

	ushist := make([]*types.UpstreamCMTS, 0)

	for rows.Next() {
		us, err := row2UPstreamCMTS(rows)
		if err != nil {
			return nil, err
		}

		ushist = append(ushist, us)
	}

	return &ushist, nil
}