package misc

import (
	"database/sql"
	"github.com/sedl/docsis-pnm/internal/logger"
	"io"
)

func CloseOrLog(intf io.Closer) {

	switch intf.(type) {
	case *sql.DB:
		err := (intf).(*sql.DB).Close()
		if err != nil {
			logger.Error(err.Error())
		}
	case *sql.Rows:
		err := (intf).(*sql.Rows).Close()
		if err != nil {
			logger.Error(err.Error())
		}
	default:
		err := intf.Close()
		if err != nil {
			logger.Error(err.Error())
		}
	}
}
