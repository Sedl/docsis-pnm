package db

import (
	"github.com/sedl/docsis-pnm/internal/migration"
)

// InitDb creates the database structure and partition tables for the current timespan
func (db *Postgres) InitDb() error {
	conn, err := db.GetConn()
	if err != nil {
		return err
	}

	if err := migration.Migrate(conn); err != nil {
		return err
	}
	if err := migration.CreateAllCurrentPartitions(conn); err != nil {
		return err
	}

	if err := migration.DropOldPartitions(conn); err != nil {
		return err
	}

	return nil
}

