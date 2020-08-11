package pgdbsyncer

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/sedl/docsis-pnm/internal/db"
	"github.com/sedl/docsis-pnm/internal/types"
	"log"
	"strings"
	"sync"
	"time"
)

var ModemDataQueueFull = fmt.Errorf("error: Can't insert modem data. Queue is full")

type nothing struct{}

type PgDbSyncer struct {
	backend           *db.Postgres
	copyUpstreams     *db.CopyFrom // modem_upstream table
	copyDownstreams   *db.CopyFrom // modem_downstream table
	copyModemdata     *db.CopyFrom // modem_data table
	copyModemUpstream *db.CopyFrom // cmts_upstream_history_modem
	mdataChan         chan *types.ModemData
	mdataWg           *sync.WaitGroup
	commitInterval    time.Duration
	copyFromWg        *sync.WaitGroup
}

func NewPgDbSyncer(postgres *db.Postgres, commitInterval time.Duration) *PgDbSyncer {
	return &PgDbSyncer{
		backend:        postgres,
		mdataChan:      make(chan *types.ModemData, 100),
		commitInterval: commitInterval,
		copyFromWg:     &sync.WaitGroup{},
		mdataWg:        &sync.WaitGroup{},
	}
}

func (m *PgDbSyncer) Run() error {
	// var conn *sql.DB
	conn, err := m.backend.GetConn()
	if err != nil {
		return err
	}

	log.Println("debug: starting modem_upstream copy goroutine")
	upstreamCopy := pq.CopyIn("modem_upstream",
		"modem_id", "poll_time", "freq", "timing_offset", "tx_power")
	m.copyUpstreams, err = db.NewCopyFrom(upstreamCopy, conn, 100, m.commitInterval)
	if err != nil {
		return err
	}

	go func() {
		m.copyFromWg.Add(1)
		defer m.copyFromWg.Done()
		m.copyUpstreams.Run()
	}()

	downstreamCopy := pq.CopyIn("modem_downstream",
		"modem_id", "poll_time", "freq", "power", "snr", "microrefl", "unerroreds",
		"correcteds", "erroreds", "modulation")
	m.copyDownstreams, err = db.NewCopyFrom(downstreamCopy, conn, 100, m.commitInterval)
	if err != nil {
		return err
	}
	go func() {
		m.copyFromWg.Add(1)
		defer m.copyFromWg.Done()
		m.copyDownstreams.Run()
	}()

	mdataCopy := pq.CopyIn("modem_data",
		"modem_id", "poll_time", "error_timeout")
	m.copyModemdata, err = db.NewCopyFrom(mdataCopy, conn, 100, m.commitInterval)
	if err != nil {
		return err
	}
	go func() {
		m.copyFromWg.Add(1)
		defer m.copyFromWg.Done()
		m.copyModemdata.Run()
	}()

	mdmCmtsUpstream := pq.CopyIn(
		"cmts_upstream_history_modem",
		"modem_id", "us_id", "poll_time", "power_rx", "snr",
		"microrefl", "unerroreds", "correcteds", "erroreds")
	m.copyModemUpstream, err = db.NewCopyFrom(mdmCmtsUpstream, conn, 10000, m.commitInterval)
	if err != nil {
		return err
	}
	go func() {
		m.copyFromWg.Add(1)
		defer m.copyFromWg.Done()
		m.copyModemUpstream.Run()
	}()

	go func() {
		m.mdataWg.Add(1)
		defer m.mdataWg.Done()
		m.updateModemData()
	}()
	return nil
}

func (m *PgDbSyncer) Stop() {

	close(m.mdataChan)
	m.mdataWg.Wait()

	m.copyUpstreams.Stop()
	m.copyDownstreams.Stop()
	m.copyModemdata.Stop()
	m.copyModemUpstream.Stop()
	m.copyFromWg.Wait()
}

// insertUpstreamData inserts records into the cmts_upstream_history table
func (m *PgDbSyncer) insertUpstreamData(mdata *types.ModemData) {
	usFreqList := make(map[int64]nothing)

	for _, us := range mdata.UpStreams {
		if us.Freq == 0 {
			// some modems do report the upstream frequency wrong
			continue
		}
		if _, ok := usFreqList[int64(us.Freq)]; ok {
			// and some report the same frequency twice
			// this check prevents a duplicate key entry in the database
			continue
		}
		m.copyUpstreams.Insert(mdata.DbModemId, mdata.Timestamp,
			us.Freq, us.TimingOffset, us.TxPower)
		usFreqList[int64(us.Freq)] = nothing{}
	}
}

func (m *PgDbSyncer) insertDownstreamData(mdata *types.ModemData) {
	dsFreqList := make(map[int]nothing)
	for _, ds := range mdata.DownStreams {
		if ds.Freq == 0 {
			continue
		}
		if _, ok := dsFreqList[int(ds.Freq)]; ok {
			// On some modems the same downstream frequency is assigned to two channels
			// This prevents a duplicate key violation
			continue
		}
		m.copyDownstreams.Insert(mdata.DbModemId, mdata.Timestamp, ds.Freq, ds.Power, ds.SNR, ds.Microrefl, ds.Unerroreds, ds.Correcteds,
			ds.Uncorrectables, ds.Modulation)
		dsFreqList[int(ds.Freq)] = nothing{}
	}
}

func (m *PgDbSyncer) insertModemData(mdata *types.ModemData) {
	errTimeout := false

	if mdata.Err != nil && strings.Contains(mdata.Err.Error(), "Request timeout") {
		errTimeout = true
	}
	m.copyModemdata.Insert(mdata.DbModemId, mdata.Timestamp, errTimeout)
}

func (m *PgDbSyncer) updateModemData() {

	for {
		select {
		case mdata, ok := <-m.mdataChan:
			if !ok {
				return
			}
			err := m.backend.UpdateFromModemData(mdata)
			if err != nil {
				log.Printf("error: can't get modem data for updating: %v\n", err)
				continue
			}

			m.insertUpstreamData(mdata)
			m.insertDownstreamData(mdata)
			m.insertDocsis31Downstreams(mdata)
			m.insertModemData(mdata)
		}
	}
}

// TODO implement insertDocsis31Downstreams
func (m *PgDbSyncer) insertDocsis31Downstreams(mdata *types.ModemData) {
	if mdata.OfdmDownstreams == nil {
		return
	}

}

// UpdateModemData updates the information in the database and inserts performance data
func (m *PgDbSyncer) UpdateModemData(mdata *types.ModemData) error {
	select {
	case m.mdataChan <- mdata:
		return nil
	default:
		return ModemDataQueueFull
	}
}
