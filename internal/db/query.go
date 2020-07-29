package db

import (
    "database/sql"
    "fmt"
    "log"
    "strings"
)

type ScanCallback func(rows *sql.Rows) (interface{}, error)

type where struct {
    column   string
    operator string
    value interface{}
}

type Query struct {
    db *sql.DB
    rows *sql.Rows
    limit int
    conditions []where
    scanCallback ScanCallback
    queryString string
    groupBy string
    orderBy string
}

func (q *Query) OrderBy(orderBy string) *Query {
    q.orderBy = orderBy
    return q
}

func (q *Query) GroupBy(groupBy string) *Query {
    q.groupBy = groupBy
    return q
}

func (q *Query) Limit(limit int) *Query {
    q.limit = limit
    return q
}

func (q *Query) Close() error {
    if q.rows != nil {
        return q.rows.Close()
    }
    return nil
}

func (q *Query) Where(column, operator string, value interface{}) *Query {
    q.conditions = append(q.conditions, where{
        column:   column,
        operator: operator,
        value:    value,
    })
    return q
}

func concatWhere(wh []where) (string, []interface{}) {
    // whereStr := ""

    whereList := make([]string, 0)

    values := make([]interface{}, 0)

    for i, w := range wh {
        whereList = append(whereList, fmt.Sprintf("%s %s $%d", w.column, w.operator, i + 1))
        values = append(values, w.value)
    }

    return strings.Join(whereList, " AND "), values
}

func (q *Query) Exec() error {
    query := q.queryString

    where, values := concatWhere(q.conditions)

    // log.Printf("%#v", values)

    if where == "" {
        where = "true"
    }

    var groupBy, orderBy, limit string
    if q.groupBy != "" {
        groupBy = fmt.Sprintf(" GROUP BY %s ", q.groupBy)
    }

    if q.limit > 0 {
        limit = fmt.Sprintf("LIMIT %d", q.limit)
    }

    query = fmt.Sprintf("%s WHERE %s %s %s %s", query, where, groupBy, orderBy, limit)

    log.Printf("debug: %s\n", query)
    rows, err := q.db.Query(query, values...)
    if err != nil {
        return err
    }
    q.rows = rows

    return nil
}

func (q *Query) Next() (interface{}, error) {
    return q.scanCallback(q.rows)
}

func NewQuery(conn *sql.DB, cb ScanCallback, query string) *Query {
    return &Query{
        db:        conn,
        rows:      nil,
        limit:     0,
        conditions: make([]where, 0),
        scanCallback: cb,
        queryString: query,
    }
}