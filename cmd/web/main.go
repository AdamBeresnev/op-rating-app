package main

import (
	"log"
	"net/http"

	"github.com/AdamBeresnev/op-rating-app/internal/db"
	"github.com/AdamBeresnev/op-rating-app/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	database := db.InitDB()
	defer database.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// TODO: Mock protected routes
	r.Post("/test-endpoint", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="bg-green-500 text-black p-4 rounded-lg">Successful API response!!!</div>`))
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var testUnits []views.TestUnit
		err := database.Select(&testUnits, "SELECT id, name, value, created_at FROM test_unit")
		if err != nil {
			log.Printf("Error querying test_unit: %v", err)
			http.Error(w, "Error querying data", 500)
			return
		}

		views.Render(w, r, views.IndexPage("op rating 2025", testUnits))
	})

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
