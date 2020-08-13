package cmts

import (
	"github.com/sedl/docsis-pnm/internal/config"
	"github.com/sedl/docsis-pnm/internal/db"
	"github.com/sedl/docsis-pnm/internal/modem"
	"github.com/sedl/docsis-pnm/internal/pgdbsyncer"
	"github.com/sedl/docsis-pnm/internal/snmp"
	"github.com/sedl/docsis-pnm/internal/types"
	"github.com/soniah/gosnmp"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type ModemPollWorkerInterface interface {
	Poll(request *modem.Poller) error
}

type Cmts struct {
	Snmp              *gosnmp.GoSNMP
	modemBucket       [][]*types.ModemInfo
	CmtsModemInfoSink chan *types.ModemInfo
	lockModemBucket   sync.RWMutex
	modemList         *types.ModemList

	DBBackend     types.DbInterface
	upstreamCache *db.CMTSUpstreamCache
	poller        ModemPollWorkerInterface
	dbRec         types.CMTSRecord
	stopChannel   chan struct{}
	modemsOnline  int32
	modemsOffline int32
	config        *config.Config
	stopWg        *sync.WaitGroup
	dbSyncer	  *pgdbsyncer.PgDbSyncer
}

func (cmts *Cmts) ValueOfDbId() int32 {
	return cmts.dbRec.Id
}

func (cmts *Cmts) ValueOfModemsOnline() int32 {
	return atomic.LoadInt32(&cmts.modemsOnline)
}

func (cmts *Cmts) ValueOfModemsOffline() int32 {
	return atomic.LoadInt32(&cmts.modemsOffline)
}

func NewCmts(
		dbRec *types.CMTSRecord,
		dbInterface types.DbInterface,
		modemPoller ModemPollWorkerInterface,
		config *config.Config,
		dbSyncer *pgdbsyncer.PgDbSyncer,
		) (*Cmts, error) {
	cmts := &Cmts{
		upstreamCache:     db.NewCMTSUpstreamCache(),
		DBBackend:         dbInterface,
		poller:            modemPoller,
		config:            config,
		CmtsModemInfoSink: make(chan *types.ModemInfo, 100000),
		dbRec:             *dbRec,
		stopChannel:       make(chan struct{}),
		stopWg:            &sync.WaitGroup{},
		dbSyncer:          dbSyncer,
		modemList: 		   types.NewModemList(),
	}

	cmts.Snmp = cmts.NewGoSNMP()

	return cmts, nil
}

func (cmts *Cmts) NewGoSNMP() *gosnmp.GoSNMP {
	return &gosnmp.GoSNMP{
		Target:             cmts.dbRec.Hostname,
		Community:          cmts.dbRec.SNMPCommunity,
		Version:            gosnmp.Version2c,
		Timeout:            time.Duration(cmts.config.Snmp.Timeout) * time.Second,
		Port:               161,
		MaxOids:            1,
		Retries:            cmts.config.Snmp.Retries,
		MaxRepetitions:     30,
		ExponentialTimeout: false,
	}
}

func (cmts *Cmts) ValueOfModemPollInterval() int {
	interval := cmts.config.Snmp.ModemPollInterval

	if interval < 60 {
		log.Println("warning: 'snmp.modempollinterval' shorter than 60 seconds. Setting it to 60 seconds.")
		interval = 60
	}

	return interval
}

func (cmts *Cmts) ValueOfCmtsPollInterval() int {
	interval := cmts.dbRec.PollInterval

	if interval < 60 {
		interval = 60
	}

	return int(interval)
}

func (cmts *Cmts) ValueOfHostname() string {
	return cmts.dbRec.Hostname
}

func (cmts *Cmts) Run() error {
	err := cmts.Snmp.Connect()
	if err != nil {
		return err
	}

	if cmts.modemBucket == nil {
		cmts.modemBucket = NewModemBucket(cmts.ValueOfModemPollInterval())
	}

	go func() {
		cmts.stopWg.Add(1)
		defer cmts.stopWg.Done()
		cmts.ModemScheduler()
	}()

	go func() {

		cmts.GoCMTSPoller()
		defer func() {
			err := cmts.Snmp.Conn.Close()
			if err != nil {
				log.Printf("error: can't close SNMP connection for CMTS %s: %v\n", cmts.dbRec.Hostname, err)
			}
		}()
	}()

	return nil
}

func (cmts *Cmts) GetUpstreamByDescr(descr string) (*types.CMTSUpstreamRecord, error) {
	upstr := cmts.upstreamCache.GetByDescr(descr)

	if upstr != nil {
		return upstr, nil
	}

	return cmts.DBBackend.GetCMTSUpstreamByDescr(cmts.dbRec.Id, descr)
}

func (cmts *Cmts) GetModemBucket() [][]*types.ModemInfo {
	cmts.lockModemBucket.RLock()
	bucket := cmts.modemBucket
	cmts.lockModemBucket.RUnlock()
	return bucket
}

var modemOids = []string{
	// DOCS-IF-MIB::docsIfCmtsCmStatusMacAddress
	".1.3.6.1.2.1.10.127.1.3.3.1.2",

	// DOCS-IF-MIB::docsIfCmtsCmStatusIpAddress
	".1.3.6.1.2.1.10.127.1.3.3.1.3",

	// DOCS-IF-MIB::docsIfCmtsCmStatusDownChannelIfIndex
	".1.3.6.1.2.1.10.127.1.3.3.1.4",

	// DOCS-IF-MIB::docsIfCmtsCmStatusUpChannelIfIndex
	".1.3.6.1.2.1.10.127.1.3.3.1.5",

	// DOCS-IF-MIB::docsIfCmtsCmStatusRxPower
	".1.3.6.1.2.1.10.127.1.3.3.1.6",

	// DOCS-IF-MIB::docsIfCmtsCmStatusTimingOffset
	".1.3.6.1.2.1.10.127.1.3.3.1.7",

	// DOCS-IF-MIB::docsIfCmtsCmStatusValue
	".1.3.6.1.2.1.10.127.1.3.3.1.9",

	// DOCS-IF-MIB::docsIfCmtsCmStatusExtUnerroreds
	".1.3.6.1.2.1.10.127.1.3.3.1.15",

	// DOCS-IF-MIB::docsIfCmtsCmStatusExtCorrecteds
	".1.3.6.1.2.1.10.127.1.3.3.1.16",

	// DOCS-IF-MIB::docsIfCmtsCmStatusExtUncorrectables
	".1.3.6.1.2.1.10.127.1.3.3.1.17",
}

var nullIP = net.IP{0, 0, 0, 0}

func (cmts *Cmts) Stop() {
	close(cmts.stopChannel)
	log.Printf("shutting down CMTS process \"%s\"\n", cmts.ValueOfHostname())
	cmts.stopWg.Wait()
	log.Printf("shutdown of CMTS process \"%s\" complete\n", cmts.ValueOfHostname())
}

func (cmts *Cmts) ListModems() (modemlist map[int]*types.ModemInfo, err error) {

	var modem_ *types.ModemInfo
	var ok bool
	var results []gosnmp.SnmpPDU

	modemlist = make(map[int]*types.ModemInfo)

	log.Printf("Fetching modem list for %s\n", cmts.dbRec.Hostname)
	start := time.Now()

	for _, oid := range modemOids {

		res, err := cmts.Snmp.BulkWalkAll(oid)

		results = append(results, res...)
		if err != nil {
			log.Printf("error: BulkWalkAll %s failed for %s ; error %v\n", oid, cmts.dbRec.Hostname, err)
			return nil, err
		}

	}
	for _, result := range results {
		oid, idx := snmp.SliceOID(result.Name)

		if modem_, ok = modemlist[idx]; !ok {
			modem_ = &types.ModemInfo{
				Index:     int32(idx),
				Timestamp: start.Unix(),
				CmtsDbId:  cmts.dbRec.Id,
			}
			modemlist[idx] = modem_
		}

		switch oid {
		case ".1.3.6.1.2.1.10.127.1.3.3.1.2":
			// DOCS-IF-MIB::docsIfCmtsCmStatusMacAddress
			modem_.MAC = result.Value.([]byte)

		case ".1.3.6.1.2.1.10.127.1.3.3.1.3":
			// DOCS-IF-MIB::docsIfCmtsCmStatusIpAddress
			modem_.IP = net.ParseIP(result.Value.(string))

		case ".1.3.6.1.2.1.10.127.1.3.3.1.4":
			// DOCS-IF-MIB::docsIfCmtsCmStatusDownChannelIfIndex
			modem_.DownIfIndex, _ = snmp.ToInt32(&result)

		case ".1.3.6.1.2.1.10.127.1.3.3.1.5":
			// DOCS-IF-MIB::docsIfCmtsCmStatusUpChannelIfIndex
			modem_.UpIfIndex, _ = snmp.ToInt32(&result)

		case ".1.3.6.1.2.1.10.127.1.3.3.1.6":
			// DOCS-IF-MIB::docsIfCmtsCmStatusRxPower
			modem_.PowerRx, _ = snmp.ToInt(&result)

		case ".1.3.6.1.2.1.10.127.1.3.3.1.7":
			// DOCS-IF-MIB::docsIfCmtsCmStatusTimingOffset
			offset, _ := snmp.ToUint(&result, cmts.Snmp)
			modem_.TimingOffset = offset

		case ".1.3.6.1.2.1.10.127.1.3.3.1.9":
			// DOCS-IF-MIB::docsIfCmtsCmStatusValue
			mstatus, _ := snmp.ToInt(&result)
			modem_.Status = (types.CmStatus)(mstatus)

		case ".1.3.6.1.2.1.10.127.1.3.3.1.15":
			// DOCS-IF-MIB::docsIfCmtsCmStatusExtUnerroreds
			modem_.Unerroreds, _ = snmp.ToUint64(&result)

		case ".1.3.6.1.2.1.10.127.1.3.3.1.16":
			// DOCS-IF-MIB::docsIfCmtsCmStatusExtCorrecteds
			modem_.Correcteds, _ = snmp.ToUint64(&result)

		case ".1.3.6.1.2.1.10.127.1.3.3.1.17":
			// DOCS-IF-MIB::docsIfCmtsCmStatusExtUncorrectables
			modem_.Uncorrectables, _ = snmp.ToUint64(&result)
		}
	}

	log.Printf("Found %d modems for %s. Time: %s\n", len(modemlist), cmts.dbRec.Hostname, time.Since(start))
	return
}
