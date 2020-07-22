package db

import (
	"database/sql"
	"github.com/sedl/docsis-pnm/internal/types"
	"net"
)

const upstreamHistoryQuery = `
SELECT
    u.poll_time,
    u.freq,
    u.modulation,
    u.timing_offset
FROM
    modem_upstream AS u
JOIN
    modem
        ON (
            modem.id = u.modem_id
        )
WHERE
    u.poll_time >= $1
    AND u.poll_time <= $2
    AND modem.mac = $3
ORDER BY
	u.poll_time ASC`

func row2UPstreamChannel(rows *sql.Rows) (*types.UpstreamChannel, int64, error) {
	var tstamp int64 = 0

	us := &types.UpstreamChannel{}
	err := rows.Scan(
		&tstamp,
		&us.Freq,
		&us.ModulationProfile,
		&us.TimingOffset,
	)
	if err != nil {
		return nil, 0, err
	}

	return us, tstamp, nil
}

func (db *Postgres) GetUpstreamHistory(mac net.HardwareAddr, from, to int) ([]*types.UpstreamChannelHistory, error) {

	conn, err := db.GetConn()
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(upstreamHistoryQuery)
	if err != nil {
		return nil, err
	}

	defer CloseOrLog(rows)

	var lasttime int64 = 0

	ushist := make([]*types.UpstreamChannelHistory, 0)

	var ush *types.UpstreamChannelHistory = nil

	for rows.Next() {
		us, tstamp, err := row2UPstreamChannel(rows)
		if err != nil {
			return nil, err
		}

		if ush == nil || lasttime != tstamp {
			if ush != nil {
				ushist = append(ushist, ush)
			}
			ush = &types.UpstreamChannelHistory{
				Timestamp: tstamp,
				Upstreams: make([]*types.UpstreamChannel, 0),
			}
			lasttime = tstamp
		}
		ush.Upstreams = append(ush.Upstreams, us)
	}

	return ushist, nil
}
