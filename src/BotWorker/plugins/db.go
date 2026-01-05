package plugins

import (
	"database/sql"
)

var GlobalDB *sql.DB

func SetGlobalDB(db *sql.DB) {
	GlobalDB = db
}
