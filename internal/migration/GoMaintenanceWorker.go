package migration

import (
	"database/sql"
	"github.com/sedl/docsis-pnm/internal/logger"
	"time"
)

type DbConnectionInterface interface {
	GetConn() (*sql.DB, error)
}

// GoMaintenanceWorker is a background task to automatically create database partition tables
// and deletion of old history data
func GoMaintenanceWorker(db DbConnectionInterface, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			conn, err := db.GetConn()
			if err != nil {
				logger.Errorf("error in GoMaintenanceWorker, can't connect to database: %s", err)
				continue
			}
			if err := CreateAllCurrentPartitions(conn); err != nil {
				logger.Errorf("Error in goroutine GoMaintenanceWorker while creating partition tables: %s", err)
			}
			if err := DropOldPartitions(conn); err != nil {
				logger.Errorf("Error in goroutine GoMaintenanceWorker while dropping old partition tables: %s", err)
			}
		}
	}
}
