package sqlx

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

func runDB() {
	var db *sqlx.DB
	defer db.Close()
	var p1, p2 string

	// Execs
	db.Exec(`DELETE FROM t`)

	queryStr := fmt.Sprintf(`INSERT INTO t VALUES ($1, $%d) `, 1) // Should not crash
	db.Exec(queryStr, p1, p2)

	db.Exec(`INSERT INTO t VALUES ($1, $2)`, p1, p2)

	const q = `INSERT INTO t(c1 c2) VALUES ($1, $2)`
	db.Exec(q, p1, p2) // want `Invalid query: syntax error at or near "c2"`

	db.Exec(`INSERT INTO t (c1, c2) VALUES ($1, $2)`, p1, "const")

	db.Exec(`INSERT INTO t (c1) VALUES ($1::uuid, $2)`, p1, p2) // want `No. of columns \(1\) not equal to no. of values \(2\)`

	db.Exec(`INSERT INTO t (c1, c2, c3, c4, c5) values ('o', $1, $1, 1, '{"duration": "1440h00m00s"}')`, time.Now())

	db.Exec(`INSERT INTO t (c1, c2, c3, c4, c5) values ('o', $1, $1, 1, '{"duration": "1440h00m00s"}')`) // // want `No. of args \(0\) is less than no. of params \(1\)`

	// QueryRow
	db.QueryRow(`INSERT INTO t (c1, c2) VALUES ($1) RETURNING c1`, p1, p2) // want `No. of columns \(2\) not equal to no. of values \(1\)`

	db.QueryRow(`INSERT INTO t (c1, c2, c3, c4) VALUES ('o', $1, 'epoch'::timestamp, $2) RETURNING c1`, p1, p2)

	db.QueryRow(`INSERT INTO t (c1, c2, c3, c4) VALUES ('o', $1, 'epoch'::timestamp, $2) RETURNING c1`, p1) // want `No. of args \(1\) is less than no. of params \(2\)`

	ctx := context.Background()
	db.ExecContext(ctx, `INSERT INTO t(c1 c2) VALUES ($1, $2)`, p1, p2) // want `Invalid query: syntax error at or near "c2"`

	db.QueryRowContext(ctx, `INSERT INTO t(c1 c2) VALUES ($1, $2) RETURNING c2`, p1, p2) // want `Invalid query: syntax error at or near "c2"`
}

func runTx() {
	// Doing a non-pointer check with transactions.
	var tx sqlx.Tx
	defer tx.Commit()
	var p1, p2 string

	// Execs
	tx.Exec(`INSERT INTO t VALUES ($1, $2)`, p1, p2)

	tx.Exec(`INSERT INTO t (c1, c2) VALUES ($1, $2)`, p1, p2)

	tx.Exec(`INSERT INTO t (c1, c2) VALUES ($1, $2)`, p1, "const")

	tx.Exec(`INSERT INTO t (c1) VALUES ($1::uuid, $2)`, p1, p2) // want `No. of columns \(1\) not equal to no. of values \(2\)`

	tx.Exec(`INSERT INTO t (c1, c2, c3, c4, c5) values ('o', $1, $1, 1, '{"duration": "1440h00m00s"}')`, time.Now())

	// QueryRow
	tx.QueryRow(`INSERT INTO t (c1, c2) VALUES ($1) RETURNING c1`, p1, p2) // want `No. of columns \(2\) not equal to no. of values \(1\)`

	tx.QueryRow(`INSERT INTO t (c1, c2, c3, c4) VALUES ('o', $1, 'epoch'::timestamp, $2) RETURNING c1`, p1, p2)

	tx.QueryRow(`INSERT INTO t (c1, c2, c3, c4) VALUES ('o', $1, 'epoch'::timestamp, $2) RETURNING c1`, p1) // want `No. of args \(1\) is less than no. of params \(2\)`

	ctx := context.Background()
	tx.ExecContext(ctx, `INSERT INTO t(c1 c2) VALUES ($1, $2)`, p1, p2) // want `Invalid query: syntax error at or near "c2"`

	tx.QueryRowContext(ctx, `INSERT INTO t(c1 c2) VALUES ($1, $2) RETURNING c2`, p1, p2) // want `Invalid query: syntax error at or near "c2"`
}
