package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/AdamBeresnev/op-rating-app/internal/db"
	"github.com/AdamBeresnev/op-rating-app/internal/middleware"
	"github.com/AdamBeresnev/op-rating-app/internal/service"
	"github.com/AdamBeresnev/op-rating-app/internal/store"
	"github.com/AdamBeresnev/op-rating-app/views"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func newRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// Serve static files
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Handle routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		views.Index().Render(r.Context(), w)
	})

	r.Post("/tournaments/entries", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		keys := make([]int, 0, len(r.Form))
		for key := range r.Form {
			if strings.HasPrefix(key, "entry_name_") {
				indexStr := strings.TrimPrefix(key, "entry_name_")
				index, err := strconv.Atoi(indexStr)
				if err == nil {
					keys = append(keys, index)
				}
			}
		}
		newIndex := 0
		if len(keys) > 0 {
			maxKey := keys[0]
			for _, key := range keys {
				if key > maxKey {
					maxKey = key
				}
			}
			newIndex = maxKey + 1
		}

		views.Entry(newIndex).Render(r.Context(), w)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth)

		r.Post("/tournaments", func(w http.ResponseWriter, r *http.Request) {
			dbConn := db.GetDB()
			bracketService := service.NewBracketService(dbConn, store.NewTournamentStore(dbConn))

			r.ParseForm()
			name := r.Form.Get("name")
			var entries []service.EntryInput
			for key, values := range r.Form {
				if strings.HasPrefix(key, "entry_name_") {
					indexStr := strings.TrimPrefix(key, "entry_name_")
					entryName := values[0]
					if entryName != "" {
						embedLinkKey := "entry_embed_link_" + indexStr
						embedLink := r.Form.Get(embedLinkKey)
						entries = append(entries, service.EntryInput{
							Name:      entryName,
							EmbedLink: embedLink,
						})
					}
				}
			}

			if id, err := bracketService.CreateTournament(r.Context(), name, entries); err != nil {
				http.Error(w, fmt.Sprintf("Failed to create tournament: %v", err), http.StatusInternalServerError)
				return
			} else {
				w.Header().Set("HX-Redirect", fmt.Sprintf("/tournaments/%s", id))
				w.WriteHeader(http.StatusOK)
			}
		})
	})

	r.Get("/tournaments/create", func(w http.ResponseWriter, r *http.Request) {
		views.CreateTournamentPage().Render(r.Context(), w)
	})

	r.Get("/tournaments/{id}", func(w http.ResponseWriter, r *http.Request) {
		dbConn := db.GetDB()
		bracketService := service.NewBracketService(dbConn, store.NewTournamentStore(dbConn))
		id := chi.URLParam(r, "id")

		data, err := bracketService.GetTournamentData(r.Context(), id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get tournament: %v", err), http.StatusInternalServerError)
			return
		}

		views.TournamentView(data.Tournament, data.Entries, data.Matches).Render(r.Context(), w)
	})

	return r
}
