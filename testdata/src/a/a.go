package a

import (
	"database/sql"
	"time"
)

func runDB() {
	var db *sql.DB
	defer db.Close()
	var p1, p2 string

	// Execs
	db.Exec(`INSERT INTO t VALUES ($1, $2)`, p1, p2)

	db.Exec(`INSERT INTO t (c1, c2) VALUES ($1, $2)`, p1, p2)

	db.Exec(`INSERT INTO t (c1, c2) VALUES ($1, $2)`, p1, "const")

	db.Exec(`INSERT INTO t (c1) VALUES ($1::uuid, $2)`, p1, p2) // want `No. of columns \(1\) not equal to no. of values \(2\)`

	db.Exec(`INSERT INTO t (c1, c2, c3, c4, c5) values ('o', $1, $1, 1, '{"duration": "1440h00m00s"}')`, time.Now())

	db.Exec(`INSERT INTO t (c1, c2, c3, c4, c5) values ('o', $1, $1, 1, '{"duration": "1440h00m00s"}')`) // // want `No. of args \(0\) is less than no. of params \(1\)`

	// QueryRow
	db.QueryRow(`INSERT INTO t (c1, c2) VALUES ($1) RETURNING c1`, p1, p2) // want `No. of columns \(2\) not equal to no. of values \(1\)`

	db.QueryRow(`INSERT INTO t (c1, c2, c3, c4) VALUES ('o', $1, 'epoch'::timestamp, $2) RETURNING c1`, p1, p2)

	db.QueryRow(`INSERT INTO t (c1, c2, c3, c4) VALUES ('o', $1, 'epoch'::timestamp, $2) RETURNING c1`, p1) // want `No. of args \(1\) is less than no. of params \(2\)`
}

func runTx() {
	var tx *sql.Tx
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
}
