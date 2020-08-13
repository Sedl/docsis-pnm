package misc

import (
    "database/sql"
    "io"
    "log"
)

func CloseOrLog(intf io.Closer) {

    switch intf.(type) {
    case *sql.DB:
        err := (intf).(*sql.DB).Close()
        if err != nil {
            log.Println(err)
        }
    case *sql.Rows:
        err := (intf).(*sql.Rows).Close()
        if err != nil {
            log.Println(err)
        }
    default:
        err := intf.Close()
        if err != nil {
            log.Println(err)
        }
    }
}


