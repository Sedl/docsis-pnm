package db

import "github.com/sedl/docsis-pnm/internal/types"

func minfo2Record (minfo *types.ModemInfo) types.ModemRecord {
    return types.ModemRecord{
        Mac:           minfo.MAC,
        IP:            minfo.IP,
        CmtsId:        minfo.CmtsDbId,
        SnmpIndex:     minfo.Index,
        CmtsDsIndex:   minfo.DownIfIndex,
        CmtsUsIndex:   minfo.UpIfIndex,
    }
}

func (db *Postgres) UpdateModemFromModemInfo(minfo *types.ModemInfo) error {

    current , err := db.ModemGetByMac(minfo.MAC)
    if err != nil {
        return err
    }

    // Does not exist -> insert
    if current == nil {
        rec := minfo2Record(minfo)
        _, err := db.ModemInsert(&rec)
        if err != nil {
            return err
        }
        return nil
    }

    changes := make(RowChangeList)

    if current.CmtsUsIndex != minfo.UpIfIndex {
        changes["cmts_us_idx"] = minfo.UpIfIndex
    }

    if current.CmtsDsIndex != minfo.DownIfIndex {
        changes["cmts_ds_idx"] = minfo.DownIfIndex
    }

    // No need to update, nothing changed
    if len(changes) == 0 {
        return nil
    }

    conn, err := db.GetConn()
    if err != nil {
        return err
    }
    _, err = UpdateRow(conn, "modem", int(current.ID), &changes)

    return nil
}
