package main

import (
	"log"
	"net/http"

	"github.com/AdamBeresnev/op-rating-app/internal/db"
)

func main() {
	database := db.InitDB()
	defer database.Close()

	router := newRouter()

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
