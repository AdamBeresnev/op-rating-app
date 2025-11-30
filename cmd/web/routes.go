package main

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/AdamBeresnev/op-rating-app/internal/bracket"
	"github.com/AdamBeresnev/op-rating-app/internal/db"
	"github.com/AdamBeresnev/op-rating-app/internal/httputil"
	"github.com/AdamBeresnev/op-rating-app/internal/middleware"
	"github.com/AdamBeresnev/op-rating-app/internal/service"
	"github.com/AdamBeresnev/op-rating-app/internal/store"
	"github.com/AdamBeresnev/op-rating-app/views"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

func newRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// Serve static files
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Handle routes
	r.Post("/tournaments/entries", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			httputil.BadRequest(w, "Invalid form data", err)
			return
		}
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

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			dbConn := db.GetDB()
			bracketService := service.NewTournamentService(dbConn, store.NewTournamentStore(dbConn))

			tournaments, err := bracketService.GetTournamentsForUser(r.Context())
			if err != nil {
				httputil.InternalServerError(w, "Failed to get tournaments", err)
				return
			}
			views.Index(tournaments).Render(r.Context(), w)
		})

		r.Post("/tournaments", func(w http.ResponseWriter, r *http.Request) {
			dbConn := db.GetDB()
			bracketService := service.NewTournamentService(dbConn, store.NewTournamentStore(dbConn))

			if err := r.ParseForm(); err != nil {
				httputil.BadRequest(w, "Invalid form data", err)
				return
			}
			name := r.Form.Get("name")
			typeStr := r.Form.Get("type")
			// Default to single elim
			tournamentType := bracket.SingleElimination
			if typeStr == "double" {
				tournamentType = bracket.DoubleElimination
			}

			var entryIndices []int
			for key := range r.Form {
				if strings.HasPrefix(key, "entry_name_") {
					indexStr := strings.TrimPrefix(key, "entry_name_")
					if index, err := strconv.Atoi(indexStr); err == nil {
						entryIndices = append(entryIndices, index)
					}
				}
			}
			sort.Ints(entryIndices)

			var entries []service.EntryInput
			for _, index := range entryIndices {
				indexStr := strconv.Itoa(index)
				entryName := r.Form.Get("entry_name_" + indexStr)
				if entryName != "" {
					embedLink := r.Form.Get("entry_embed_link_" + indexStr)
					entries = append(entries, service.EntryInput{
						Name:      entryName,
						EmbedLink: embedLink,
					})
				}
			}

			if id, err := bracketService.CreateTournament(r.Context(), name, tournamentType, entries); err != nil {
				httputil.InternalServerError(w, "Failed to create tournament", err)
				return
			} else {
				w.Header().Set("HX-Redirect", fmt.Sprintf("/tournaments/%s", id))
				w.WriteHeader(http.StatusOK)
			}
		})

		r.Get("/matches/{id}", func(w http.ResponseWriter, r *http.Request) {
			dbConn := db.GetDB()
			matchService := service.NewMatchService(dbConn, store.NewTournamentStore(dbConn))
			id := chi.URLParam(r, "id")

			data, err := matchService.GetMatchViewData(r.Context(), id)
			if err != nil {
				httputil.InternalServerError(w, "Failed to get match data", err)
				return
			}
			views.MatchView(data.Match, data.Entry1, data.Entry2, data.NextMatchID).Render(r.Context(), w)
		})

		r.Post("/matches/{id}/advance", func(w http.ResponseWriter, r *http.Request) {
			dbConn := db.GetDB()
			matchService := service.NewMatchService(dbConn, store.NewTournamentStore(dbConn))
			idStr := chi.URLParam(r, "id")
			matchID, err := uuid.Parse(idStr)
			if err != nil {
				httputil.BadRequest(w, "Invalid match ID", err)
				return
			}
			if err := r.ParseForm(); err != nil {
				httputil.BadRequest(w, "Invalid form data", err)
				return
			}
			winnerIDStr := r.Form.Get("winner_id")
			winnerID, err := uuid.Parse(winnerIDStr)
			if err != nil {
				httputil.BadRequest(w, "Invalid winner ID", err)
				return
			}
			tournamentID, err := matchService.AdvanceWinner(r.Context(), matchID, winnerID)
			if err != nil {
				httputil.InternalServerError(w, "Failed to advance winner", err)
				return
			}

			data, err := matchService.GetMatchViewData(r.Context(), matchID.String())
			if err != nil {
				httputil.InternalServerError(w, "Failed to get next match info", err)
				return
			}
			winnerSlot := 0
			if data.Match.WinnerSlot != nil {
				winnerSlot = *data.Match.WinnerSlot
			}
			views.MatchVotingResult(data.NextMatchID, tournamentID, winnerSlot).Render(r.Context(), w)
		})
	})

	r.Get("/tournaments/create", func(w http.ResponseWriter, r *http.Request) {
		views.CreateTournamentPage().Render(r.Context(), w)
	})

	r.Get("/tournaments/{id}", func(w http.ResponseWriter, r *http.Request) {
		dbConn := db.GetDB()
		bracketService := service.NewTournamentService(dbConn, store.NewTournamentStore(dbConn))
		id := chi.URLParam(r, "id")

		data, err := bracketService.GetTournamentData(r.Context(), id)
		if err != nil {
			httputil.InternalServerError(w, "Failed to get tournament", err)
			return
		}

		views.TournamentView(data.Tournament, data.Entries, data.Matches, data.NextMatchID).Render(r.Context(), w)
	})

	return r
}
