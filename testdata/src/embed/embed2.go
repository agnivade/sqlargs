package embed

// Test taken from https://github.com/danjac/kanban/blob/master/database/tasks.go
import (
	"database/sql"
)

type Task struct {
	ID     int64  `db:"id" json:"id,string"`
	CardID int64  `db:"card_id" json:"-"`
	Text   string `db:"label" json:"text" binding:"required,max=60"`
}

type TaskDB interface {
	Delete(int64) error
	Create(*Task) error
	Move(int64, int64) error
}

type defaultTaskDB struct {
	*sql.DB
}

func (db *defaultTaskDB) Create(task *Task) {
	db.Exec("insertinto tasks(card_id, label) values (?, ?)", task.CardID, task.Text) // want `Invalid query: syntax error at or near "insertinto"`

	db.Exec("insert into tasks(card_id, label) values ($1, $2)", task.CardID) // // want `No. of args \(1\) is less than no. of params \(2\)`
}

func (db *defaultTaskDB) Move(taskID int64, newCardID int64) {
	db.Exec("update tasks set card_id=$1 where id=$2", newCardID)
}

// Testing non-pointer receiver.
func (db defaultTaskDB) Delete(taskID int64) {
	db.Exec("delete from taskswhere id=$1", taskID) // want `Invalid query: syntax error at or near "="`
	db.Exec("delete from tasks where id=$1", taskID)
}
