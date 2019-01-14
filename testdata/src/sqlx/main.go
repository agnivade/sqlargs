package main

import (
	"github.com/jmoiron/sqlx"
)

func main() {
	var db *sqlx.DB
	var p1, p2 string
	db.Exec(`INSERT INTO t (c1) VALUES ($1::uuid, $2)`, p1, p2) // want `No. of columns \(1\) not equal to no. of values \(2\)`
}
