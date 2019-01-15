package sqlx

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

func runSqlxDB() {
	var db *sqlx.DB
	defer db.Close()
	var p1, p2 string

	// Execs
	db.MustExec(`DELETE FROM t`)

	queryStr := fmt.Sprintf(`INSERT INTO t VALUES ($1, $%d) `, 1) // Should not crash
	db.MustExec(queryStr, p1, p2)

	db.MustExec(`INSERT INTO t VALUES ($1, $2)`, p1, p2)

	const q = `INSERT INTO t(c1 c2) VALUES ($1, $2)`
	db.MustExec(q, p1, p2) // want `Invalid query: syntax error at or near "c2"`

	db.MustExec(`INSERT INTO t (c1, c2) VALUES ($1, $2)`, p1, "const")

	db.MustExec(`INSERT INTO t (c1) VALUES ($1::uuid, $2)`, p1, p2) // want `No. of columns \(1\) not equal to no. of values \(2\)`

	db.MustExec(`INSERT INTO t (c1, c2, c3, c4, c5) values ('o', $1, $1, 1, '{"duration": "1440h00m00s"}')`, time.Now())

	db.MustExec(`INSERT INTO t (c1, c2, c3, c4, c5) values ('o', $1, $1, 1, '{"duration": "1440h00m00s"}')`) // // want `No. of args \(0\) is less than no. of params \(1\)`

	// Queryx
	db.Queryx(`SELECT * FROM students`)
	db.Queryx(`INSERT INTO t (c1, c2) VALUES ($1) RETURNING c1`, p1, p2) // want `No. of columns \(2\) not equal to no. of values \(1\)`

	// QueryRowx
	db.QueryRowx(`INSERT INTO t (c1, c2) VALUES ($1) RETURNING c1`, p1, p2) // want `No. of columns \(2\) not equal to no. of values \(1\)`

	db.QueryRowx(`INSERT INTO t (c1, c2, c3, c4) VALUES ('o', $1, 'epoch'::timestamp, $2) RETURNING c1`, p1, p2)

	db.QueryRowx(`INSERT INTO t (c1, c2, c3, c4) VALUES ('o', $1, 'epoch'::timestamp, $2) RETURNING c1`, p1) // want `No. of args \(1\) is less than no. of params \(2\)`

	// Context
	ctx := context.Background()
	db.MustExecContext(ctx, `INSERT INTO t(c1 c2) VALUES ($1, $2)`, p1, p2)               // want `Invalid query: syntax error at or near "c2"`
	db.QueryxContext(ctx, `INSERT INTO t(c1 c2) VALUES ($1, $2) RETURNING c2`, p1, p2)    // want `Invalid query: syntax error at or near "c2"`
	db.QueryRowxContext(ctx, `INSERT INTO t(c1 c2) VALUES ($1, $2) RETURNING c2`, p1, p2) // want `Invalid query: syntax error at or near "c2"`
}

func runSqlxTx() {
	// Doing a non-pointer check with transactions.
	var tx sqlx.Tx
	defer tx.Commit()
	var p1, p2 string

	// Execs
	tx.MustExec(`INSERT INTO t VALUES ($1, $2)`, p1, p2)

	tx.MustExec(`INSERT INTO t (c1, c2) VALUES ($1, $2)`, p1, p2)

	tx.MustExec(`INSERT INTO t (c1, c2) VALUES ($1, $2)`, p1, "const")

	tx.MustExec(`INSERT INTO t (c1) VALUES ($1::uuid, $2)`, p1, p2) // want `No. of columns \(1\) not equal to no. of values \(2\)`

	tx.MustExec(`INSERT INTO t (c1, c2, c3, c4, c5) values ('o', $1, $1, 1, '{"duration": "1440h00m00s"}')`, time.Now())

	// QueryRow
	tx.QueryRowx(`INSERT INTO t (c1, c2) VALUES ($1) RETURNING c1`, p1, p2) // want `No. of columns \(2\) not equal to no. of values \(1\)`

	tx.QueryRowx(`INSERT INTO t (c1, c2, c3, c4) VALUES ('o', $1, 'epoch'::timestamp, $2) RETURNING c1`, p1, p2)

	tx.QueryRowx(`INSERT INTO t (c1, c2, c3, c4) VALUES ('o', $1, 'epoch'::timestamp, $2) RETURNING c1`, p1) // want `No. of args \(1\) is less than no. of params \(2\)`

	ctx := context.Background()
	tx.MustExecContext(ctx, `INSERT INTO t(c1 c2) VALUES ($1, $2)`, p1, p2)               // want `Invalid query: syntax error at or near "c2"`
	tx.QueryxContext(ctx, `INSERT INTO t(c1 c2) VALUES ($1, $2) RETURNING c2`, p1, p2)    // want `Invalid query: syntax error at or near "c2"`
	tx.QueryRowxContext(ctx, `INSERT INTO t(c1 c2) VALUES ($1, $2) RETURNING c2`, p1, p2) // want `Invalid query: syntax error at or near "c2"`
}
