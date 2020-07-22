package db

import "github.com/sedl/docsis-pnm/internal/types"

const insertCMTSUpstreamHistory = "INSERT INTO cmts_upstream_history (upstream_id, poll_time, unerroreds, correcteds, erroreds, utilization, pkts_broadcast, pkts_unicast, bytes, mer_db) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)"
func (db *Postgres) InsertCMTSUpstreamHistory(record *types.CMTSUpstreamHistoryRecord) error {
	conn, err := db.GetConn()
	if err != nil {
		return err
	}

	_, err = conn.Exec(insertCMTSUpstreamHistory,
		record.UpstreamId,
		record.PollTime.Unix(),
		record.Unerroreds,
		record.Correcteds,
		record.Uncorrectrables,
		record.Utilization,
		record.PktsBroadcast,
		record.PktsUnicast,
		record.Bytes,
		record.MER)

	return err
}
