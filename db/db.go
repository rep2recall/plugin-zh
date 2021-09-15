package db

import (
	"database/sql"
	"log"

	"github.com/patarapolw/atexit"
	// sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Connect() {
	db, err := sql.Open("sqlite3", "file:./assets/zh.db?_journal=WAL")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	DB = db

	atexit.Register(func() {
		DB.Close()
	})
	atexit.Listen()
}

func Close() {
	DB.Close()
}
