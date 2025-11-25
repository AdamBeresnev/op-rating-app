package service

import (
	"context"
	"strings"

	"github.com/AdamBeresnev/op-rating-app/internal/bracket"
	"github.com/AdamBeresnev/op-rating-app/internal/store"
	"github.com/AdamBeresnev/op-rating-app/internal/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type EntryGeneration struct {
	db    *sqlx.DB
	store *store.TournamentStore
}

func (s EntryGeneration) ParseInput(ctx context.Context, entryLinks string) ([]bracket.Entry, []string, error) {
	var links []string
	var entries []bracket.Entry
	var tournamentID uuid.UUID //TODO
	var entryName string       //TODO
	var maxSeed int            //TODO
	var entryImage string      //TODO

	links = strings.Split(entryLinks, "\n")

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, links, err
	}
	defer tx.Rollback()

	for i, link := range links {
		e := bracket.Entry{
			ID:           uuid.New(),
			TournamentID: tournamentID,
			Name:         entryName,
			Seed:         i + maxSeed,
			EmbedLink:    utils.StringOrNil(link),
			ImageLink:    utils.StringOrNil(entryImage),
		}

		entries = append(entries, e)
	}

	return entries, links, tx.Commit()
}
