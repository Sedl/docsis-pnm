package cmts

import (
	"crypto/sha512"
	"encoding/binary"
	"github.com/sedl/docsis-pnm/internal/constants"
	"github.com/sedl/docsis-pnm/internal/modem"
	"github.com/sedl/docsis-pnm/internal/types"
	"log"
	"sync/atomic"
	"time"
)

func NewModemBucket(size int) [][]*types.ModemInfo {
	bucket := make([][]*types.ModemInfo, size)
	for i := 0; i < size; i++ {
		bucket[i] = make([]*types.ModemInfo, 0)
	}

	return bucket
}

func (cmts *Cmts) updateModemList() error {

	tstart := time.Now()
	interval := cmts.ValueOfModemPollInterval()
	bucket := NewModemBucket(interval)
	mlist, err := cmts.ListModems()
	if err != nil {
		log.Printf("failed: %v\n", err)
		return err
	}
	tsnmp := time.Since(tstart)

	tstart = time.Now()

	var mOnline, mOffline int32

	for _, modem_ := range mlist {
		if modem_.IP.Equal(nullIP) {
			mOffline++
			continue
		}

		// skip offline modems
		if ! (modem_.Status == constants.CmStatusRegistrationComplete || modem_.Status == constants.CmStatusOperational) {
			mOffline++
			continue
		}

		mOnline++
		err = cmts.DBBackend.UpdateModemFromModemInfo(modem_)
		if err != nil {
			return err
		}

		hash := sha512.Sum512(modem_.MAC)
		pos := binary.LittleEndian.Uint64(hash[0:8])
		pos = pos % uint64(interval)
		bucket[pos] = append(bucket[pos], modem_)
	}

	tDb := time.Since(tstart)

	cmts.lockModemBucket.Lock()
	cmts.modemBucket = bucket
	cmts.lockModemBucket.Unlock()

	cmts.modemList.ReplaceMap(mlist)

	atomic.StoreInt32(&cmts.modemsOffline, mOffline)
	atomic.StoreInt32(&cmts.modemsOnline, mOnline)

	log.Printf("debug: fetching modems (%d online, %d offline) from CMTS %s finished. Time: (%s SNMP, %s DB, %s total)", mOnline, mOffline, cmts.ValueOfHostname(), tsnmp, tDb, tsnmp+tDb)
	return nil
}

func (cmts *Cmts) ModemScheduler() {

	log.Printf("debug: starting modem poll timer for CMTS %s\n", cmts.dbRec.Hostname)
	defer log.Printf("debug: modem poll timer for CMTS %s exited\n", cmts.dbRec.Hostname)
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case ticktime := <-ticker.C:
			pos := ticktime.Unix() % int64(len(cmts.modemBucket))
			bucket := cmts.GetModemBucket()
			for i := range bucket[pos] {
				mdm := bucket[pos][i]

				request := &modem.Poller{
					Hostname:  mdm.IP.String(),
					Community: cmts.GetModemCommunity(),
					// Community: config.Configuration.Snmp.Community,
					CmtsId:    cmts.dbRec.Id,
					Mac:       mdm.MAC,
					SnmpIndex: mdm.Index,
				}
				// log.Printf("debug: scheduling modem %s for polling\n", request.Mac.String())
				err := cmts.poller.Poll(request)
				if err != nil {
					log.Printf("Can't schedule modem %s for polling. Modem poll queue is full. Consider increasing number of poll workers", bucket[pos][i].MAC)
				}
			}
		case _, ok := <- cmts.stopChannel:
			if ! ok {
				return
			}

		}
	}
}
