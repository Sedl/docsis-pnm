package manager

import (
	"errors"
	"github.com/sedl/docsis-pnm/internal/cmts"
	"github.com/sedl/docsis-pnm/internal/config"
	"github.com/sedl/docsis-pnm/internal/db"
	"github.com/sedl/docsis-pnm/internal/pgdbsyncer"
	"github.com/sedl/docsis-pnm/internal/pollworker"
	"github.com/sedl/docsis-pnm/internal/tftp"
	"github.com/sedl/docsis-pnm/internal/types"
	"log"
	"net"
	"sync"
	"time"
)

// Manager does all the plumbing between the different components
type Manager struct {
	db          *db.Postgres
	modemPoller *pollworker.PollWorker
	dbSyncer    *pgdbsyncer.PgDbSyncer
	cmtsList    map[int32]*cmts.Cmts
	config      *config.Config
	tftpServer  *tftp.Server
	cmtsMutex   sync.RWMutex
}

func tftpServerInstance(cfg config.Tftp) *tftp.Server {
	if cfg.ExternalAddress == "" {
		log.Println("WARNING! external TFTP address not set, disabling TFTP functionality")
		return nil
	} else {
		ipa := net.ParseIP(cfg.ExternalAddress)
		if ipa == nil {
			log.Fatalf(
				"invalid external TFTP IP address %q. Please correct this in your config and retry",
				cfg.ExternalAddress)
		}
		return tftp.NewServer(ipa)
	}
}

func NewManager(config *config.Config) (*Manager, error) {
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
	dbSyncer := pgdbsyncer.NewPgDbSyncer(pg, time.Duration(config.Db.CommitInterval)*time.Second)

	// start modem poller goroutines
	log.Println("debug: init modem pollers")
	poller := pollworker.NewPollWorker(&config.Snmp, dbSyncer)

	manager := &Manager{
		db:          pg,
		modemPoller: poller,
		dbSyncer:    dbSyncer,
		cmtsList:    make(map[int32]*cmts.Cmts),
		config:      config,
		tftpServer:  tftpServerInstance(config.Tftp),
	}

	return manager, nil
}

func (m *Manager) GetTftpServerInstance() *tftp.Server {
	return m.tftpServer
}

func (m *Manager) GetDbInterface() *db.Postgres {
	return m.db
}

func (m *Manager) GetCmtsModemCommunity(cmtsId int32) string {
	m.cmtsMutex.RLock()
	defer m.cmtsMutex.RUnlock()
	if cmtsobj, ok := m.cmtsList[cmtsId]; ok {
		return cmtsobj.GetModemCommunity()
	} else {
		return ""
	}
}

func (m *Manager) AddCMTS(cmtsrec *types.CMTSRecord) (*cmts.Cmts, error) {
	cmtsobj, err := cmts.NewCmts(cmtsrec, m.db, m.modemPoller, m.config, m.dbSyncer)
	if err != nil {
		return nil, err
	}
	m.cmtsMutex.Lock()
	defer m.cmtsMutex.Unlock()
	if _, ok := m.cmtsList[cmtsrec.Id]; ok {
		return nil, errors.New("a cmts with this id already exists")
	}
	m.cmtsList[cmtsrec.Id] = cmtsobj
	// m.cmtsList = append(m.cmtsList, cmtsobj)
	return cmtsobj, nil
}

var dnsError *net.DNSError

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
			if errors.As(err, &dnsError) {
				log.Printf("error: DNS lookup failed, skipping host: %v\n", err)
				continue
			} else {
				return err
			}
		}
	}
	return nil
}

// RemoveCmts returns true if the CMTS was found and removed, false otherwise
func (m *Manager) RemoveCmts(cmtsobj *cmts.Cmts) bool {

	log.Printf("debug: stopping CMTS %s\n", cmtsobj.ValueOfHostname())

	id := cmtsobj.ValueOfDbId()

	m.cmtsMutex.Lock()
	defer m.cmtsMutex.Unlock()
	if cmtsobj, ok := m.cmtsList[id]; ok {
		cmtsobj.Stop()
		delete(m.cmtsList, id)
		return true
	} else {
		return false
	}
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

	// start TFTP server
	if m.tftpServer != nil {
		go func() {
			log.Fatal(m.tftpServer.ListenAndServe(":69"))
		}()
	}

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
