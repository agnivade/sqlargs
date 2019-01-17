// Test taken from https://github.com/croll/arkeogis-server/blob/master/model/database.go
package sqlx

import (
	"bytes"

	"github.com/jmoiron/sqlx"
)

const Database_handle_InsertStr = "\"database_id\", \"import_id\", \"identifier\", \"url\", \"declared_creation_date\", \"created_at\""
const Database_handle_InsertValuesStr = "$1, $2, $3, $4, $5, now()"

type Database struct {
	Id               int    `db:"id" json:"id"`
	Name             string `db:"name" json:"name" min:"1" max:"255" error:"DATABASE.FIELD_NAME.T_CHECK_MANDATORY"`
	Owner            int    `db:"owner" json:"owner"` // User.Id
	Editor           string `db:"editor" json:"editor"`
	Contributor      string `db:"contributor" json:"contributor"`
	Default_language string `db:"default_language" json:"default_language"` // Lang.Isocode
	State            string `db:"state" json:"state" enum:"undefined,in-progress,finished" error:"DATABASE.FIELD_STATE.T_CHECK_INCORRECT"`
	License_id       int    `db:"license_id" json:"license_id"` // License.Id
}

// AnotherExistsWithSameName checks if database already exists with same name and owned by another user
func (d *Database) AnotherExistsWithSameName(tx *sqlx.Tx) (exists bool, err error) {
	tx.QueryRowx("SELECT id FROM \"database\" WHERE name = $1 AND owner != $2", d.Name, d.Owner).Scan(&d.Id)
	return true, nil
}

// Get retrieves informations about a database stored in the main table
func (d *Database) Get(tx *sqlx.Tx) (err error) {
	stmt, err := tx.PrepareNamed("SELECT * from \"database\" WHERE id=$1")
	defer stmt.Close()
	return stmt.Get(d, d)
}

// AddHandle links a handle  to a database
func (d *Database) AddHandle(tx *sqlx.Tx) (id int, err error) {
	stmt, err := tx.PrepareNamed("INSERT INTO \"database_handle\" (" + Database_handle_InsertStr + ") VALUES (" + Database_handle_InsertValuesStr + ") RETURNING id")
	tx.PrepareNamed("INSERT INTOdatabase_handle\" (" + Database_handle_InsertStr + ") VALUES (" + Database_handle_InsertValuesStr + ") RETURNING id") // want `Invalid query: syntax error at or near "INTOdatabase_handle"`
	defer stmt.Close()
	return
}

// DeleteHandles unlinks handles
func (d *Database) DeleteSpecificHandle(tx *sqlx.Tx, id int) error {
	_, err := tx.Exec("DELETE FROM \"database_handle\" WHERE identifier = $1", id)
	return err
}

// SetContexts links users as contexts to a database
func (d *Database) SetContexts(tx *sqlx.Tx, contexts []string) error {
	for _, cname := range contexts {
		tx.Exec("INSERT INTO \"database_context\" (database_id) VALUES ($1, $2)", d.Id, cname)      // want `No. of columns \(1\) not equal to no. of values \(2\)`
		tx.Exec("INSERT INTO \"database_context\" (database_id, context) VALUES ($1)", d.Id, cname) // // want `No. of columns \(2\) not equal to no. of values \(1\)`
	}
	return nil
}

// DeleteContexts deletes the context linked to a database
func (d *Database) DeleteContexts(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"\"database_context\" WHERE database_id=$1", d) // want `Invalid query: zero-length delimited identifier at or near """"`
	return err
}

func (d *Database) SetTranslations(tx *sqlx.Tx, field string, translations []struct {
	Lang_Isocode string
	Text         string
}) (err error) {
	var transID int
	for _, tr := range translations {
		err = tx.QueryRow("SELECT count(database_id) FROM database_tr WHERE database_id = $1 AND lang_isocode = $2", d.Id, tr.Lang_Isocode).Scan(&transID)
		if transID == 0 {
			_, err = tx.Exec("INSERT INTO database_tr (database_id, lang_isocode, description, geographical_limit, bibliography, context_description, source_description, source_relation, copyright, subject) VALUES ($1, $2, '', '', '', '', '', '', '', '')", d.Id, tr.Lang_Isocode)
			_, err = tx.Exec("INSERT INTO database_tr (database_id, lang_isocode, description, geographical_limit, bibliography, context_description, source_description, source_relation, copyright, subject) VALUES ($1, $2, '', '', '', '', '', '', '', '')", d.Id) // want `No. of args \(1\) is less than no. of params \(2\)`
		}
	}
	return
}

// UpdateFields updates "database" fields (crazy isn't it ?)
func (d *Database) UpdateFields(tx *sqlx.Tx, params interface{}, fields ...string) (err error) {
	var upd string
	query := "UPDATE \"database\" SET " + upd + " WHERE id = :id"
	_, err = tx.NamedExec(query, params)
	return
}

// CacheGeom get database sites extend and cache enveloppe
func (d *Database) CacheGeom(tx *sqlx.Tx) (err error) {
	var c int
	err = tx.Get(&c, "SELECT COUNT(*) FROM (SELECT DISTINCT geom FROM site WHERE database_id = $1) AS temp", d.Id)
	// Envelope
	if c > 2 {
		_, err = tx.NamedExec("UPDATE database SET geographical_extent_geom = (SELECT (ST_Envelope((SELECT ST_Multi(ST_Collect(f.geom)) as singlegeom FROM (SELECT (ST_Dump(geom::::geometry)).geom As geom FROM site WHERE database_id = $1) As f)))) WHERE id =$1", d) // want `Invalid query: syntax error at or near "::"`
	} else {
		_, err = tx.NamedExec("UPDATE database SET geographical_extent_geom = (SELECT ST_Buffer((SELECT geom FROM site WHERE database_id = $1 AND geom IS NOT NULL LIMIT 1), 1)) WHERE id = $1", d)
	}
	return
}

// CacheDates get database sites extend and cache enveloppe
func (d *Database) CacheDates(tx *sqlx.Tx) (err error) {
	_, err = tx.NamedExec("UPDATE database SET start_date = (SELECT COALESCE(min(start_date1),-2147483648) FROM site_range WHERE site_id IN (SELECT id FROM site where database_id = $1) AND start_date1 != -2147483648), end_date = (SELECT COALESCE(max(end_date2),2147483647) FROM site_range WHERE site_id IN (SELECT id FROM site where database_id = $1) AND end_date2 != 2147483647) WHERE id = $1", d)
	return
}

// LinkToUserProject links database to project
func (d *Database) LinkToUserProject(tx *sqlx.Tx, project_ID int) (err error) {
	_, err = tx.Exec("INSERT INTO project__database (project_id, database_id) VALUES ($1, $2)", project_ID, d.Id)
	return
}

// ExportCSV exports database and sites as as csv file
func (d *Database) ExportCSV(tx *sqlx.Tx, siteIDs ...[]int) (outp string, err error) {
	var buff bytes.Buffer
	const q = "WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.name::TEXT AS path FROM charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = $1 AND ca.parent_id = 0 UNION ALL SELECT ca.id, (p.path || ';' cat.name) FROM nodes_cte AS p, charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = $1 AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC"

	tx.Query(q, d.Default_language) // want `Invalid query: syntax error at or near "cat"`

	tx.Query(`SELECT s.code, s.name, s.city_name, s.city_geonameid, ST_X(s.geom::geometry) as longitude, ST_Y(s.geom::geometry) as latitude, ST_X(s.geom_3d::geometry) as longitude_3d, ST_Y(s.geom_3d::geometry) as latitude3d, ST_Z(s.geom_3d::geometry) as altitude, s.centroid, s.occupation, sr.start_date1, sr.start_date2, sr.end_date1, sr.end_date2, src.exceptional, src.knowledge_type, srctr.bibliography, srctr.comment, c.id as charac_id FROM site s LEFT JOIN site_range sr ON s.id = sr.site_id LEFT JOIN site_tr str ON s.id = str.site_id LEFT JOIN site_range__charac src ON sr.id = src.site_range_id LEFT JOIN site_range__charac_tr srctr ON src.id = srctr.site_range__charac_id LEFT JOIN charac c ON src.charac_id = c.id WHERE s.database_id = $1 AND str.lang_isocode IS NULL OR str.lang_isocode = $2 ORDER BY s.id, sr.id`, d.Id, d.Default_language)

	return buff.String(), nil
}
