package pollworker

import (
	"errors"
	"github.com/sedl/docsis-pnm/internal/logger"
	"github.com/sedl/docsis-pnm/internal/modem"
	"github.com/sedl/docsis-pnm/internal/types"
	"sync"
	"sync/atomic"
)

type PollWorker struct {
	WorkerCount    int
	requestChan    chan *modem.Poller
	ModemDataSink  chan *types.ModemData
	wg             *sync.WaitGroup
	statError      uint64
	statOK         uint64
	config         *types.Snmp
	dbModemUpdater types.ModemUpdaterInterface
}

func NewPollWorker(config *types.Snmp, modemUpdater types.ModemUpdaterInterface) *PollWorker {
	return &PollWorker{
		requestChan:    make(chan *modem.Poller, 10000),
		config:         config,
		dbModemUpdater: modemUpdater,
		WorkerCount:    config.WorkerCount,
		ModemDataSink:  make(chan *types.ModemData, 10000),
		wg:             &sync.WaitGroup{},
	}
}

func (p *PollWorker) GetPollOk() uint64 {
	return atomic.LoadUint64(&p.statOK)
}

func (p *PollWorker) GetPollErr() uint64 {
	return atomic.LoadUint64(&p.statError)
}

func (p *PollWorker) GetRequestQueueLength() int {
	return len(p.requestChan)
}

func (p *PollWorker) Poll(request *modem.Poller) error {
	select {
	case p.requestChan <- request:
	default:
		return errors.New("error: modem poll queue is full")
	}
	return nil
}

func poll(req *modem.Poller) (*types.ModemData, error) {
	err := req.Connect()
	if err != nil {
		return nil, err
	}

	defer func() {
		err := req.Close()
		if err != nil {
			logger.Errorf("Error closing SNMP connection for modem %s: %s", req.Hostname, err)
		}
	}()

	return req.Poll()
}

func (p *PollWorker) collector() {
	p.wg.Add(1)
	defer p.wg.Done()
	for {
		select {

		case request, ok := <-p.requestChan:
			if !ok {
				return
			}
			// logger.Debugf("collecting data from modem %s (%s)\n", request.Mac.String(), request.Hostname)
			// TODO SNMP Community aus der Datenbank holen, ggf. schon in den Request packen
			mdata, err := poll(request)
			// TODO error an mdata struct weitergeben um Fehlerdiagnose zu ermöglichen und um Fehler auswerten zu können
			if err != nil {
				logger.Errorf("Error while collecting data from modem (%s) (%q)", request.Hostname, err)
				atomic.AddUint64(&p.statError, 1)
				if mdata == nil {
					continue
				}
			} else {
				atomic.AddUint64(&p.statOK, 1)
			}

			mdata.Mac = request.Mac
			mdata.SnmpIndex = request.SnmpIndex
			mdata.CmtsDbId = request.CmtsId
			err = p.dbModemUpdater.UpdateModemData(mdata)
			if err != nil {
				logger.Errorf("error: updating modem in database failed: %v", err)
			}
			// logger.Debugf("debug: done collecting data from modem %s (time: %s)", request.Mac.String(), time.Duration(mdata.QueryTime))
		}
	}
}

func (p *PollWorker) Run() {
	logger.Infof("Starting %d modem data collectors", p.WorkerCount)
	for i := 0; i < p.WorkerCount; i++ {
		go p.collector()
	}
}

func (p *PollWorker) Stop() {
	close(p.requestChan)
	p.wg.Wait()
}

func (p *PollWorker) StopCollector() {
	close(p.requestChan)
}
