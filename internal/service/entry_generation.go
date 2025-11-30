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
	var returnLinks []string
	var entries []bracket.Entry
	var tounamentID uuid.UUID //TODO

	links = strings.Split(entryLinks, "\n")

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, links, err
	}
	defer tx.Rollback()

	for i, link := range links {
		e, err := createEntryFromLink(tounamentID, i, link)

		if err != nil {
			returnLinks = append(returnLinks, link)
			continue
		}

		entries = append(entries, e)
	}

	return entries, returnLinks, tx.Commit()
}

func createEntryFromLink(tournamentID uuid.UUID, seed int, entryLink string) (bracket.Entry, error) {
	var err error

	placeholderName := "Anime"
	placeholderImage := "https://pub-92474f7785774e91a790e086dfa6b2ef.r2.dev/anime/large-cover/PRdGioMCwXIr2ixbBXDC5yh4A70KA7vZ1JtqCJoE.jpg"

	entry := bracket.Entry{
		ID:           uuid.New(),
		TournamentID: tournamentID,
		Name:         placeholderName,
		ImageLink:    &placeholderImage,
		Seed:         seed,
		EmbedLink:    utils.StringOrNil(entryLink),
	}

	return entry, err
}
