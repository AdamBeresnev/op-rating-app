package main

import (
	"log"
	"net/http"

	"github.com/AdamBeresnev/op-rating-app/internal/db"
)

func main() {
	database := db.InitDB()
	defer database.Close()

	if err := db.RunMigrations(database.DB); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	router := newRouter()

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
