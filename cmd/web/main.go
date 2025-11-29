package main

import (
	"log"
	"net/http"
	"time"

	"github.com/AdamBeresnev/op-rating-app/internal/db"
	"github.com/AdamBeresnev/op-rating-app/internal/middleware"
	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	database := db.InitDB()
	defer database.Close()

	if err := db.RunMigrations(database.DB); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	middleware.InitAuth()

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Store = sqlite3store.New(database.DB)

	router := newRouter(sessionManager)

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
