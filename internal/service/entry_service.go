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

type EntryService struct {
	db    *sqlx.DB
	store *store.TournamentStore
}

func NewEntryService(db *sqlx.DB, store *store.TournamentStore) *EntryService {
	return &EntryService{db: db, store: store}
}

func (s *EntryService) ParseInput(ctx context.Context, tournamentID string, entryLinks string) ([]bracket.Entry, error) {
	var links []string
	var entries []bracket.Entry

	links = strings.Split(entryLinks, "\n")

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for i, link := range links {
		e, err := createEntryFromLink(tournamentID, i, link)

		if err != nil {
			continue
		}

		entries = append(entries, e)
	}

	err = s.store.CreateEntries(ctx, tx, entries)

	if err != nil {
		return nil, err
	}

	return entries, tx.Commit()
}

func createEntryFromLink(tournamentID string, seed int, entryLink string) (bracket.Entry, error) {
	var entry bracket.Entry

	placeholderName := "Anime"
	placeholderImage := "https://pub-92474f7785774e91a790e086dfa6b2ef.r2.dev/anime/large-cover/PRdGioMCwXIr2ixbBXDC5yh4A70KA7vZ1JtqCJoE.jpg"

	tournamentUUID, err := uuid.Parse(tournamentID) 
	if err != nil {
		return entry, err
	}

	entry = bracket.Entry{
		ID:           uuid.New(),
		TournamentID: tournamentUUID,
		Name:         placeholderName,
		ImageLink:    &placeholderImage,
		Seed:         seed,
		EmbedLink:    utils.StringOrNil(entryLink),
	}

	return entry, err
}
