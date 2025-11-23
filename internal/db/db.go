package db

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB() *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", "op_rating.db?_journal_mode=WAL")
	if err != nil {
		log.Fatalln("Failed to connect to DB:", err)
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database connected.")
	return db
}
