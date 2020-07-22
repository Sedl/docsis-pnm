package db

import (
	"database/sql"
	"fmt"
	"github.com/sedl/docsis-pnm/internal/types"
)

const modemQueryString = `SELECT
    MAX(modem_data.poll_time) AS mtime,
    modem_data.modem_id,
    modem_data.error_timeout,
    modem.sysdescr,
    modem.ip,
    modem.cmts_id
FROM
    modem_data,
    modem 
WHERE
    modem.mac = $1
    AND modem.id = modem_data.modem_id 
group by
    modem_data.modem_id,
    modem_data.error_timeout,
    modem.sysdescr,
    modem.ip,
    modem.cmts_id,
    modem.snmp_index`

const modemDownstreams = `select
    freq,
    power,
    snr,
    microrefl,
    unerroreds,
    correcteds,
    erroreds,
    modulation 
from
    modem_downstream 
where
    modem_id = $1 
    and poll_time = $2`

const modemUpstreams = `select
    freq,
    modulation,
    timing_offset 
from
    modem_upstream 
where
    modem_id = $1
    and poll_time = $2`

func getUpstreams(conn *sql.DB, modemID int64, timestamp int64) ([]types.UpstreamChannel, *types.ApiError) {
	rows, err := conn.Query(modemUpstreams, modemID, timestamp)

	if err != nil {
		return nil, &types.ApiError{
			ErrorStr:       fmt.Sprintf("Error while getting upstream channels: %s", err),
			HttpStatusCode: 500}
	}

	defer CloseOrLog(rows)

	chlist := make([]types.UpstreamChannel, 0)

	for rows.Next() {
		uchan := types.UpstreamChannel{}
		err := rows.Scan(
			&uchan.Freq,
			&uchan.ModulationProfile,
			&uchan.TimingOffset,
		)
		if err != nil {
			return nil, &types.ApiError{
				ErrorStr:       fmt.Sprintf("Error fetching upstream channels: %s", err),
				HttpStatusCode: 500}
		}
		chlist = append(chlist, uchan)
	}

	return chlist, nil
}

func getDownstreams(conn *sql.DB, modemID int64, timestamp int64) ([]types.DownstreamChannel, *types.ApiError) {
	rows, err := conn.Query(modemDownstreams, modemID, timestamp)

	if err != nil {
		return nil, &types.ApiError{
			ErrorStr:       fmt.Sprintf("Error while getting downstream channels: %s", err),
			HttpStatusCode: 500}
	}

	defer CloseOrLog(rows)

	chlist := make([]types.DownstreamChannel, 0)

	for rows.Next() {
		dchan := types.DownstreamChannel{}
		err := rows.Scan(
			&dchan.Freq,
			&dchan.Power,
			&dchan.SNR,
			&dchan.Microrefl,
			&dchan.Unerroreds,
			&dchan.Correcteds,
			&dchan.Uncorrectables,
			&dchan.Modulation,
		)
		if err != nil {
			return nil, &types.ApiError{
				ErrorStr:       fmt.Sprintf("Error fetching downstream channels: %s", err),
				HttpStatusCode: 500}
		}
		chlist = append(chlist, dchan)
	}
	return chlist, nil
}
