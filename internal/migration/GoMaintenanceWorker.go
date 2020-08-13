package migration

import (
    "database/sql"
    "log"
    "time"
)

type DbConnectionInterface interface {
    GetConn() (*sql.DB, error)
}

// GoMaintenanceWorker is a background task to automatically create database partition tables
// and should handle other database maintenance stuff like expunge of old records after the expiration time
func GoMaintenanceWorker(db DbConnectionInterface, interval time.Duration) {
    ticker := time.NewTicker(interval)
    for {
        select {
        case <-ticker.C:
            conn, err := db.GetConn()
            if err != nil {
                log.Printf("error in GoMaintenanceWorker, can't connect to database: %s", err)
                continue
            }
            if err := CreateAllCurrentPartitions(conn); err != nil {
                log.Printf("Error in goroutine GoMaintenanceWorker while creating partition tables: %s\n", err)
                continue
            }
        }
    }
}
