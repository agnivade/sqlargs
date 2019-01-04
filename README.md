# sqlargs
A vet analyzer which checks sql(only Postgres!) queries for correctness.

### Background

Let's assume you have a query like:

`db.Exec("insert into table (c1, c2, c3, c4) values ($1, $2, $3, $4)", p1, p2, p3, p4)`. And it's the middle of the night and you need to add a new column. You quickly change the query to -

`db.Exec("insert into table (c1, c2, c3, c4, c5) values ($1, $2, $3, $4)", p1, p2, p3, p4, p5)`. Everything compiles fine. Except it's not ! A `$5` is missing !

This is a semantic error which can only get caught during runtime. Ofcourse, if there are tests, then this does not happen. But sometimes I get lazy and don't write tests. :sweat_smile:

Therefore I wrote a vet analyzer which will statically check for semantic errors like these and flag them beforehand.

### Quick start

This is written using the `go/analysis` API. So you can plug this directly into `go vet`, or you can run it as a standalone tool too.

Install:
```
go get github.com/agnivade/sqlargs/cmd/sqlargs
```

And then run it on your repo:
```
go vet -vettool $(which sqlargs) ./... # Has to be >= 1.12
OR
sqlargs ./...
```

### P.S.: This only works for Postgres queries. So if your codebase has queries which do not match with the postgres query parser, it might flag incorrect errors.
