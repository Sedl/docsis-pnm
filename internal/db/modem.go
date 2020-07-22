package db

import (
	"github.com/sedl/docsis-pnm/internal/types"
	"net"
)

// UpdateFromModemData updates the modems table and creates an entry if the modem does not exist
func (db *Postgres) UpdateFromModemData(data *types.ModemData) error {

	// log.Printf("debug: updating modem data for %v", data.Mac)
	oldmdata, err := db.ModemGetByMac(data.Mac)
	if err != nil {
		return err
	}

	rec := &types.ModemRecord{
		ID:        data.DbModemId,
		Mac:       data.Mac,
		SysDescr:  data.Sysdescr,
		IP:        net.ParseIP(data.Hostname),
		CmtsId:    data.CmtsDbId,
		SnmpIndex: data.SnmpIndex,
		DocsisVersion: data.DocsisVersion,
		DSPrimary: data.DSPrimary,
	}

	if oldmdata == nil {
		// no entries in cache or database -> insert into database
		id, err := db.ModemInsert(rec)
		if err != nil {
			panic(err.Error())
		}
		data.DbModemId = id
		return nil
	} else {
		rec.ID = oldmdata.ID
		// update modem database ID in source
		data.DbModemId = oldmdata.ID
	}

	_, err = db.ModemUpdate(rec)
	return nil
}
