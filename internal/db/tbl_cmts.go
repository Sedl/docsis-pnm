package db

import (
	"github.com/sedl/docsis-pnm/internal/misc"
	"github.com/sedl/docsis-pnm/internal/types"
)

const cmtsQueryStr = "SELECT id, hostname, snmp_community, snmp_community_modem, disabled, poll_interval, snmp_max_repetitions FROM cmts"
const cmtsInsertStr = "INSERT INTO cmts (hostname, snmp_community, snmp_community_modem, disabled, poll_interval, snmp_max_repetitions) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"

// InsertCMTS inserts a CMTSRecord and fills in the CMTSRecord.Id with the newly created database ID
func (db *Postgres) InsertCMTS(cmts *types.CMTSRecord) (err error) {
	conn, err := db.GetConn()
	if err != nil {
		return
	}

	rows, err := conn.Query(
		cmtsInsertStr,
		cmts.Hostname,
		cmts.SNMPCommunity,
		cmts.SNMPModemCommunity,
		cmts.Disabled,
		cmts.PollInterval,
		cmts.MaxRepetitions)
	if err != nil {
		return
	}
	defer misc.CloseOrLog(rows)
	rows.Next()

	var dbId int32

	err = rows.Scan(&dbId)
	if err != nil {
		return
	}

	cmts.Id = dbId
	return
}

// GetCMTSByHostname retrieves a CMTSRecord from the cache. If the CMTS is not in the cache, it gets it from the
// database. Returns nil if there is no CMTS with the given hostname
func (db *Postgres) GetCMTSByHostname(hostname string) (*types.CMTSRecord, error) {
	conn, err := db.GetConn()
	if err != nil {
		return nil, err
	}

	q, err := NewCMTSQuery(conn, "hostname = $1", hostname)
	if err != nil {
		return nil, err
	}
	defer misc.CloseOrLog(q)

	return q.Next()
}

func (db *Postgres) GetCMTSById(id uint32) (*types.CMTSRecord, error) {

	conn, err := db.GetConn()
	if err != nil {
		return nil, err
	}

	rows, err := NewCMTSQuery(conn, "id = $1", id)
	if err != nil {
		return nil, err
	}
	defer misc.CloseOrLog(rows)

	return rows.Next()
}

func (db *Postgres) GetCMTSAll() (*[]*types.CMTSRecord, error) {
	conn, err := db.GetConn()
	if err != nil {
		return nil, err
	}

	rows, err := NewCMTSQuery(conn, "")
	if err != nil {
		return nil, err
	}
	records := make([]*types.CMTSRecord, 0)
	for {
		row, err := rows.Next()
		if err != nil {
			return nil, err
		}

		if row == nil {
			break
		}
		records = append(records, row)
	}

	return &records, nil
}
