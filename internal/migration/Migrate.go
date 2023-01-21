package migration

import (
	"database/sql"
	"github.com/sedl/docsis-pnm/internal/logger"
)

const (
	// PartitionInterval defines the interval in seconds by which the performace data tables are getting partitioned
	PartitionInterval = 86400

	// PartitionRetentionCount is the number of partitions to keep, older partitions will get deleted
	PartitionRetentionCount = 14
)

// Migrate checks the current database schema revision and updates the schema appropriately
func Migrate(conn *sql.DB) error {

	version, err := GetDbRevision(conn)
	if err != nil {
		return err
	}

	maxRevision := len(Revisions) - 1
	logger.Infof("current db revision %d, upgrading to %d", version, maxRevision)

	for i := version + 1; i <= maxRevision; i++ {
		logger.Infof("Upgrading database schema to version %d", i)
		err = Revisions[i](conn)
		if err != nil {
			return err
		}
	}

	return nil
}
