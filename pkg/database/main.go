package database

import (
	"database/sql"
	"github.com/gperis/forza-bot/pkg/config"
	"log"
	_ "modernc.org/sqlite"
)

type conf struct {
	DbPath string `mapstructure:"path"`
}

var moduleConf conf

func init() {
	config.Load("database", &moduleConf)
}

func OpenDatabase() (db *sql.DB) {
	db, err := sql.Open("sqlite", moduleConf.DbPath)
	if err != nil {
		log.Fatal(err)
	}

	return db
}
