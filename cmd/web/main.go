package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/AdamBeresnev/op-rating-app/internal/db"
)

type TestUnit struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	Value     int       `db:"value"`
	CreatedAt time.Time `db:"created_at"`
}

func main() {
	database := db.InitDB()
	defer database.Close()

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	mux.HandleFunc("/test-endpoint", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="bg-green-500 text-black p-4 rounded-lg">Successful API response!!!</div>`))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles(
			"templates/layout.html",
			"templates/index.html",
		)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var testUnits []TestUnit
		err = database.Select(&testUnits, "SELECT id, name, value, created_at FROM test_unit")
		if err != nil {
			log.Printf("Error querying test_unit: %v", err)
			http.Error(w, "Error querying data", 500)
			return
		}

		data := map[string]interface{}{
			"Title":     "op rating 2025",
			"TestUnits": testUnits,
		}

		tmpl.Execute(w, data)
	})

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
