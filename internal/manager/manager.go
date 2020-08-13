package manager

import (
	"github.com/sedl/docsis-pnm/internal/cmts"
	"github.com/sedl/docsis-pnm/internal/config"
	"github.com/sedl/docsis-pnm/internal/db"
	"github.com/sedl/docsis-pnm/internal/pollworker"
	"github.com/sedl/docsis-pnm/internal/pgdbsyncer"
	"github.com/sedl/docsis-pnm/internal/types"
	"log"
	"time"
)

// Manager does all the plumbing between the different components
type Manager struct {
	db *db.Postgres
	modemPoller *pollworker.PollWorker
	dbSyncer *pgdbsyncer.PgDbSyncer
	cmtsList []*cmts.Cmts
	config *config.Config
}


func NewManager(config *config.Config) (*Manager, error){

	// initialize database stuff
	log.Println("debug: connecting to database")
	pg, err := db.NewPostgres(config.Db)
	if err != nil {
		return nil, err
	}
	err = pg.InitDb()
	if err != nil {
		log.Printf("error: database init failed: %v\n", err)
		return nil, err
	}

	log.Println("debug: connecting to database successful")

	log.Println("debug: init database syncer")
	// TODO move syncer into db struct
	dbSyncer := pgdbsyncer.NewPgDbSyncer(pg, time.Duration(config.Db.CommitInterval) * time.Second)

	// start modem poller goroutines
	log.Println("debug: init modem pollers")
	poller := pollworker.NewPollWorker(&config.Snmp, dbSyncer)

	manager := &Manager{
		db: pg,
		modemPoller: poller,
		dbSyncer: dbSyncer,
		cmtsList: make([]*cmts.Cmts, 0),
		config: config,
	}

	return manager, nil
}

func (m *Manager) GetDbInterface () *db.Postgres {
	return m.db
}

func (m *Manager) GetCmtsList() []*cmts.Cmts {
	return m.cmtsList
}

func (m *Manager) AddCMTS(cmtsrec *types.CMTSRecord) (*cmts.Cmts, error) {
    cmtsobj, err := cmts.NewCmts(cmtsrec, m.db, m.modemPoller, m.config, m.dbSyncer)
	if err != nil {
		return nil, err
	}
	m.cmtsList = append(m.cmtsList, cmtsobj)
	return cmtsobj, nil
}

func (m *Manager) AddAllCmtsFromDb() error {
    pg := m.GetDbInterface()
	cmtslist, err := pg.GetCMTSAll()
	if err != nil {
		return err
	}

	for _, cmtsrec := range *cmtslist {
		log.Printf("debug: init CMTS %s", cmtsrec.Hostname)
		if cmtsrec.Disabled {
			continue
		}
		cmtsobj, err := m.AddCMTS(cmtsrec)
		if err != nil {
			log.Printf("error: init of CMTS failed: %v\n", err)
			return err
		}
		err = cmtsobj.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) RemoveCmts(cmtsobj *cmts.Cmts) bool {

	log.Printf("debug: stopping CMTS %s\n", cmtsobj.ValueOfHostname())

	found := -1
	var pos int
	var cmtsL *cmts.Cmts

	for pos, cmtsL = range m.cmtsList {
	    if cmtsL == cmtsobj {
	    	found = pos
	    	break
		}
	}

	if found == -1 {
		return false
	}

	// remove element from slice
	// this doesn't maintain order but we don't care
	m.cmtsList[pos] = m.cmtsList[len(m.cmtsList)-1]
	m.cmtsList = m.cmtsList[:len(m.cmtsList)-1]

	if cmtsL != nil {
		cmtsL.Stop()
	}

	return true
}

func (m *Manager) Run() error {
	// start all goroutines for database
	log.Println("debug: starting database goroutines")
	m.db.Run()

	// run database syncer
	log.Println("debug: starting database syncer")
	err := m.dbSyncer.Run()
	if err != nil {
		return err
	}

	m.modemPoller.Run()
	return nil

}

func (m *Manager) Stop() {
	log.Println("debug: shutting down application")

	for _, cmtss := range m.cmtsList {
		m.RemoveCmts(cmtss)
	}

	m.modemPoller.Stop()
	m.dbSyncer.Stop()

}