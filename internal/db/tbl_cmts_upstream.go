package db

import (
	"database/sql"
	"github.com/sedl/docsis-pnm/internal/types"
)

const cmtsUpstreamQuery = "SELECT id, cmts_id, snmp_idx, descr, freq, alias, admin_status FROM cmts_upstream"
const cmtsUpstreamQueryBySnmpIndex = cmtsUpstreamQuery + " WHERE cmts_id = $1 AND snmp_idx = $2"
const cmtsUpstreamQueryByDescr = cmtsUpstreamQuery + " WHERE cmts_id = $1 AND descr = $2"

const cmtsUpstreamInsert = "INSERT INTO cmts_upstream (cmts_id, snmp_idx, descr, freq, alias, admin_status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"

func rows2CMTSUpstreamRecord(rows *sql.Rows) (*types.CMTSUpstreamRecord, error) {
	up := &types.CMTSUpstreamRecord{}

	var descr, alias sql.NullString
	err := rows.Scan(
		&up.ID,
		&up.CMTSID,
		&up.SNMPIndex,
		&descr,
		&up.Freq,
		&alias,
		&up.AdminStatus,
		)
	if err != nil {
		return nil, err
	}

	if descr.Valid {
		up.Description = descr.String
	} else {
		up.Description = ""
	}

	if alias.Valid {
		up.Alias = alias.String
	}

	return up, nil
}

func (db *Postgres) InsertCMTSUpstream(record *types.CMTSUpstreamRecord) error {
	conn, err := db.GetConn()
	if err != nil {
		return err
	}

	rows, err := conn.Query(
		cmtsUpstreamInsert,
		record.CMTSID,
		record.SNMPIndex,
		record.Description,
		record.Freq,
		record.Alias,
		record.AdminStatus,
	)
	if err != nil {
		return err
	}
	defer CloseOrLog(rows)

	rows.Next()

	var dbid int

	err = rows.Scan(&dbid)
	if err != nil {
		return err
	}

	record.ID = int32(dbid)

	return nil
}

func (db *Postgres) UpdateCmtsUpstreams (records map[int]*types.CMTSUpstreamRecord) error {

	changes := make(RowChangeList)

	for _, upstr := range records {
		// Description seems to be the most "stable" identifier
		upstrDb , err := db.GetCMTSUpstreamByDescr(upstr.CMTSID, upstr.Description)
		if err != nil {
			return err
		}
		if upstrDb == nil {
			err = db.InsertCMTSUpstream(upstr)
			if err != nil {
				return err
			}
			continue
		}

		if upstr.ID == 0 {
			upstr.ID = upstrDb.ID
		}

		if upstrDb.Alias != upstr.Alias {
			changes["alias"] = upstr.Alias
		}
		if upstrDb.Freq != upstr.Freq {
			changes["freq"] = upstr.Freq
		}
		if upstrDb.AdminStatus != upstr.AdminStatus {
			changes["admin_status"] = upstr.AdminStatus
		}

		if len(changes) > 0 {
			conn, err := db.GetConn()
			_, err = UpdateRow(conn, "cmts_upstream", int(upstrDb.ID), &changes)
			if err != nil {
				return err
			}
			changes = make(RowChangeList)
		}

	}
	return nil
}

func (db *Postgres) GetCMTSUpstreamByDescr(cmtsId int32, descr string) (*types.CMTSUpstreamRecord, error) {

	conn, err := db.GetConn()
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(cmtsUpstreamQueryByDescr, cmtsId, descr)
	if err != nil {
		return nil, err
	}
	defer CloseOrLog(rows)

	// upstream was not found in database
	if ! rows.Next() {
		return nil, nil
	}

	upstr, err := rows2CMTSUpstreamRecord(rows)
	if err != nil {
		return nil, err
	}

	// m.cacheUpstreams.Add(upstr)

	return upstr, nil
}


func (db *Postgres) GetCMTSUpstreamBySnmpIndex(cmtsId, idx int) (*types.CMTSUpstreamRecord, error) {
	/*
	upstr := m.cacheUpstreams.GetByIndex(idx)
	if upstr != nil {
		return upstr, nil
	}
	*/
	conn, err := db.GetConn()
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(cmtsUpstreamQueryBySnmpIndex, cmtsId, idx)
	if err != nil {
		return nil, err
	}
	defer CloseOrLog(rows)

	// upstream was not found in database
	if ! rows.Next() {
		return nil, nil
	}

	upstr, err := rows2CMTSUpstreamRecord(rows)
	if err != nil {
		return nil, err
	}

	// m.cacheUpstreams.Add(upstr)

	return upstr, nil
}
