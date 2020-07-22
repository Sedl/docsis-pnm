package db

import (
	"database/sql"
	"fmt"
	"strings"
)

type RowChangeList map[string]interface{}

// UpdateRow updates a row from the specified table. It returns the number of affected rows.
func UpdateRow(dbconn *sql.DB, table string, rowId int, changes *RowChangeList) (int64, error) {

	i := 2

	cols := make([]string, 0)
	// add 1 to make room for the id
	values := make([]interface{}, 0)

	values = append(values, rowId)

	for col, value := range *changes {
		cols = append(cols, fmt.Sprintf("%s = $%d", col, i))
		values = append(values, value)
		i++
	}

	colstr := strings.Join(cols, ", ")
	updateStr := fmt.Sprintf("UPDATE %s SET %s WHERE id = $1", table,  colstr)

	result, err := dbconn.Exec(updateStr, values...)
	if err != nil {
		return 0, nil
	}
	rows, err := result.RowsAffected()

	return rows, err
}
