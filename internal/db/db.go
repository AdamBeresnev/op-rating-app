package db

import (
	"database/sql"
	"errors"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB() *sqlx.DB {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "op_rating.db?_journal_mode=WAL"
	}

	var err error
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		log.Fatalln("Failed to connect to DB:", err)
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database connected.")
	dbInstance = db
	return db
}

var dbInstance *sqlx.DB

func GetDB() *sqlx.DB {
	return dbInstance
}

func RunMigrations(db *sql.DB) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	log.Println("Migrations ran successfully")
	return nil
}
