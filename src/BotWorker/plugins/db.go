package plugins

import (
	"botworker/internal/store"
	"database/sql"

	"gorm.io/gorm"
)

var GlobalDB *sql.DB
var GlobalGORMDB *gorm.DB
var GlobalStore *store.Store

func SetGlobalDB(db *sql.DB) {
	GlobalDB = db
}

func SetGlobalGORMDB(db *gorm.DB) {
	GlobalGORMDB = db
	GlobalStore = store.NewStore(db)
}
