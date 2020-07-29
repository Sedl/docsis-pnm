package modem

import (
	"errors"
	"github.com/sedl/docsis-pnm/internal/config"
	"github.com/sedl/docsis-pnm/internal/types"
	"log"
	"sync"
	"sync/atomic"
)

type Poller struct {
	WorkerCount    int
	requestChan    chan *types.ModemPollRequest
	ModemDataSink  chan *types.ModemData
	wg             *sync.WaitGroup
	statError      uint64
	statOK         uint64
	config         *config.Snmp
	dbModemUpdater types.ModemUpdaterInterface
}

func NewPoller(config *config.Snmp, modemUpdater types.ModemUpdaterInterface) *Poller {
	return &Poller{
		requestChan:    make(chan *types.ModemPollRequest, 10000),
		config:         config,
		dbModemUpdater: modemUpdater,
		WorkerCount:    config.WorkerCount,
		ModemDataSink:  make(chan *types.ModemData, 10000),
		wg: &sync.WaitGroup{},
	}
}

func (p *Poller) GetPollOk() uint64 {
	return atomic.LoadUint64(&p.statOK)
}

func (p *Poller) GetPollErr() uint64 {
	return atomic.LoadUint64(&p.statError)
}

func (p *Poller) GetRequestQueueLength() int {
	return len(p.requestChan)
}

func (p *Poller) Poll(request *types.ModemPollRequest) error {
	select {
	case p.requestChan <- request:
	default:
		return errors.New("error: modem poll queue is full")
	}
	return nil
}

func (p *Poller) collector() {
	p.wg.Add(1)
	defer p.wg.Done()
	for {
		select {

		case request, ok := <-p.requestChan:
			if ! ok {
				return
			}
			// log.Printf("debug: collecting data from modem %s (%s)\n", request.Mac.String(), request.Hostname)
			// TODO SNMP Community aus der Datenbank holen, ggf. schon in den Request packen
			mdata, err := Poll(request.Hostname, request.Mac, request.Community)
			// TODO error an mdata struct weitergeben um Fehlerdiagnose zu ermöglichen und um Fehler auswerten zu können
			if err != nil {
				log.Printf("Error while collecting data from modem (%s) (%q)", request.Hostname, err)
				atomic.AddUint64(&p.statError, 1)
				if mdata == nil {
					continue
				}
				// log.Printf("%#v\n", mdata)
			} else {
				atomic.AddUint64(&p.statOK, 1)
			}

			mdata.Mac = request.Mac
			mdata.SnmpIndex = request.SnmpIndex
			mdata.CmtsDbId = request.CmtsId
			//			if mdata.Err != nil {
			//				log.Printf("Error while collecting data from modem (%s) (%q): %s", request.Hostname, mdata.Sysdescr, mdata.Err)
			//				atomic.AddUint64(&p.statError, 1)
			//			} else {

			// insertModemData(mdata, *p.ModemDataSink)
			err = p.dbModemUpdater.UpdateModemData(mdata)
			if err != nil {
				log.Printf("error: updating modem in database failed: %v\n", err)
			}
			// log.Printf("debug: done collecting data from modem %s\n", request.Mac.String())
		}
	}
}

func (p *Poller) Run() {
	log.Printf("Starting %d modem data collectors\n", p.WorkerCount)
	for i := 0; i < p.WorkerCount; i++ {
		go p.collector()
	}
}

func (p *Poller) Stop() {
	close(p.requestChan)
	p.wg.Wait()
}

func (p *Poller) StopCollector() {
	close(p.requestChan)
}
