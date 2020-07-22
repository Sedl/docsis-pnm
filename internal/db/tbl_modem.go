package db

import (
	"database/sql"
	"github.com/sedl/docsis-pnm/internal/types"
	"net"
	"sync"
)

const insertModemQueryString = "INSERT INTO modem (mac, sysdescr, ip, cmts_id, snmp_index, docsis_ver, ds_primary, cmts_ds_idx, cmts_us_idx) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id"
const modemQueryStr = "SELECT id, mac, sysdescr, ip, cmts_id, snmp_index, docsis_ver, ds_primary, cmts_ds_idx, cmts_us_idx FROM modem"
const modemUpdate = "UPDATE modem SET %s WHERE id = $1"

func (db *Postgres) ModemInsert(record *types.ModemRecord) (uint64, error) {

	conn, err := db.GetConn()
	if err != nil {
		return 0, err
	}

	var modemId uint64
	err = conn.QueryRow(
		insertModemQueryString,
		record.Mac.String(),
		record.SysDescr,
		record.IP.String(),
		record.CmtsId,
		record.SnmpIndex,
		record.DocsisVersion,
		record.DSPrimary,
		record.CmtsDsIndex,
		record.CmtsUsIndex,
	).Scan(&modemId)
	if err != nil {
		return 0, err
	}

	record.ID = modemId

//	db.cacheModem.Update(record)
	return modemId, nil
}

type ModemQuery struct {
	db *sql.DB
	rows *sql.Rows
}

func (q *ModemQuery) Close() error {
	if q.rows != nil {
		return q.rows.Close()
	}
	return nil
}

func (q *ModemQuery) Next() (*types.ModemRecord, error) {
	if ! q.rows.Next() {
		return nil, nil
	}

	var mac, sysdescr, ip sql.NullString

	record := &types.ModemRecord{}
	err := q.rows.Scan(
		&record.ID,
		&mac,
		&sysdescr,
		&ip,
		&record.CmtsId,
		&record.SnmpIndex,
		&record.DocsisVersion,
		&record.DSPrimary,
		&record.CmtsDsIndex,
		&record.CmtsUsIndex)
	if err != nil {
		return nil, err
	}

	if mac.Valid {
		if record.Mac, err = net.ParseMAC(mac.String); err != nil {
			return nil, err
		}
	}

	if sysdescr.Valid {
		record.SysDescr = sysdescr.String
	}

	if ip.Valid {
		record.IP = net.ParseIP(ip.String)
	}

	return record, nil
}

func NewModemQuery(conn *sql.DB, where string , args... interface{}) (*ModemQuery, error){
	mq := &ModemQuery{conn, nil}

	var query string
	if where != "" {
		query = modemQueryStr + " WHERE " + where
	} else {
		query = modemQueryStr
	}
	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, err
	}

	mq.rows = rows
	return mq, nil
}

// ModemGetByMac might return data from the cache
func (db *Postgres) ModemGetByMac(mac net.HardwareAddr) (*types.ModemRecord, error) {

	conn, err := db.GetConn()
	if err != nil {
		return nil, err
	}

	rows, err := NewModemQuery(conn, "mac = $1", mac.String())
	if err != nil {
		return nil, err
	}
	defer CloseOrLog(rows)

	row, err := rows.Next()
	return row, err
}

func modemDiff(old, new *types.ModemRecord) map[string]interface{} {
	changes := make(map[string]interface{})

	if new.SysDescr != "" && old.SysDescr != new.SysDescr {
		// log.Printf("sysdescr changed from %q to %q (%s)\n", mdataold.Sysdescr, mdatanew.Sysdescr, mdatanew.Mac.String())
		changes["sysdescr"] = new.SysDescr
	}

	if new.CmtsId > 0 && old.CmtsId != new.CmtsId {
		changes["cmts_id"] = new.CmtsId
	}

	if new.IP.Equal(net.IP{0,0,0,0}) && ! old.IP.Equal(new.IP) {
		changes["ip"] = new.IP.String()
	}

	if new.SnmpIndex > 0 && old.SnmpIndex != new.SnmpIndex {
		changes["snmp_index"] = new.SnmpIndex
	}

	if new.DocsisVersion != old.DocsisVersion {
		changes["docsis_ver"] = new.DocsisVersion
	}

	if new.DSPrimary != old.DSPrimary {
		changes["ds_primary"] = new.DSPrimary
	}

	if new.CmtsDsIndex != old.CmtsDsIndex {
		changes["cmts_ds_idx"] = new.CmtsDsIndex
	}

	if new.CmtsUsIndex != old.CmtsUsIndex {
		changes["cmts_us_idx"] = new.CmtsUsIndex
	}

	return changes
}

// ModemUpdate UPDATEs the modem table
func (db *Postgres) ModemUpdate(record *types.ModemRecord) (int, error) {
	conn, err := db.GetConn()
	if err != nil {
		return 0, err
	}
	have, err := db.ModemGetByMac(record.Mac)
	if err != nil {
		return 0, err
	}
	changelist := modemDiff(have, record)
	changes := len(changelist)
	/*
	if changes > 0 {
		log.Printf("debug: updating modem %s, %d changes\n", record.Mac.String(), changes)
	}
	 */
	err = tableUpdate(conn, modemUpdate, have.ID, changelist)
	if err != nil {
		return 0, err
	}
	return changes, nil
}

type ModemCache struct {
	cache map[string]*types.ModemRecord
	lock sync.RWMutex
}

func NewModemCache() *ModemCache {
	return &ModemCache{
		cache: make(map[string]*types.ModemRecord),
	}
}

func (mcache *ModemCache) Get(mac net.HardwareAddr) *types.ModemRecord {
	mcache.lock.RLock()
	defer mcache.lock.RUnlock()

	if val, ok := mcache.cache[mac.String()]; ok {
		return val
	}
	return nil
}

func (mcache *ModemCache) Update(rec *types.ModemRecord) {
	mcache.lock.Lock()
	mcache.cache[rec.Mac.String()] = rec
	mcache.lock.Unlock()
}