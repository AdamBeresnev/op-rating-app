package store

import (
	"context"
	"testing"
	"time"

	"github.com/AdamBeresnev/op-rating-app/internal/bracket"
	"github.com/AdamBeresnev/op-rating-app/internal/utils"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSuperUserID = "00000000-0000-0000-0000-000000000001"

// setupTestDB creates an in-memory SQLite database and applies migrations
func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	database, err := sqlx.Connect("sqlite3", "file::memory:")
	require.NoError(t, err, "Failed to connect to in-memory DB")

	_, err = database.Exec("PRAGMA foreign_keys = ON;")
	require.NoError(t, err)

	driver, err := sqlite3.WithInstance(database.DB, &sqlite3.Config{})
	require.NoError(t, err, "Failed to create migrate driver instance")

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"sqlite3",
		driver,
	)
	require.NoError(t, err, "Failed to create migrate instance")

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err, "Failed to apply migrations")
	}

	return database
}

func TestCreateTournament(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewTournamentStore(db)

	mockOwnerID := uuid.MustParse(testSuperUserID)

	tournament := &bracket.Tournament{
		ID:               uuid.New(),
		OwnerID:          mockOwnerID,
		Name:             "Test Tournament",
		Status:           bracket.TournamentDraft,
		Type:             bracket.SingleElimination,
		ScoreRequirement: 0,
		CreatedAt:        time.Now().UTC(),
	}

	tx, err := db.BeginTxx(context.Background(), nil)
	require.NoError(t, err)

	err = store.CreateTournament(context.Background(), tx, tournament)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	var fetchedTournament bracket.Tournament
	err = db.Get(&fetchedTournament, "SELECT * FROM tournaments WHERE id=$1", tournament.ID)
	require.NoError(t, err)

	assert.Equal(t, tournament.ID, fetchedTournament.ID)
	assert.Equal(t, tournament.OwnerID, fetchedTournament.OwnerID)
	assert.Equal(t, tournament.Name, fetchedTournament.Name)
	assert.Equal(t, tournament.Status, fetchedTournament.Status)
	assert.Equal(t, tournament.Type, fetchedTournament.Type)
	assert.Equal(t, tournament.ScoreRequirement, fetchedTournament.ScoreRequirement)
	assert.WithinDuration(t, tournament.CreatedAt, fetchedTournament.CreatedAt, time.Second)
}

func TestCreateEntries(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewTournamentStore(db)

	tournamentID := uuid.New()
	mockOwnerID := uuid.MustParse(testSuperUserID)

	tournament := &bracket.Tournament{
		ID:               tournamentID,
		OwnerID:          mockOwnerID,
		Name:             "Test Tournament",
		Status:           bracket.TournamentDraft,
		Type:             bracket.SingleElimination,
		ScoreRequirement: 0,
		CreatedAt:        time.Now().UTC(),
	}
	tx, err := db.BeginTxx(context.Background(), nil)
	require.NoError(t, err)
	err = store.CreateTournament(context.Background(), tx, tournament)
	require.NoError(t, err)
	err = tx.Commit()
	require.NoError(t, err)

	entries := []bracket.Entry{
		{ID: uuid.New(), TournamentID: tournamentID, Name: "Entry 1", Seed: 1, EmbedLink: utils.StringOrNil("link1")},
		{ID: uuid.New(), TournamentID: tournamentID, Name: "Entry 2", Seed: 2, EmbedLink: nil},
	}

	tx, err = db.BeginTxx(context.Background(), nil)
	require.NoError(t, err)

	err = store.CreateEntries(context.Background(), tx, entries)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	var fetchedEntries []bracket.Entry
	err = db.Select(&fetchedEntries, "SELECT * FROM entries WHERE tournament_id=$1 ORDER BY seed", tournamentID)
	require.NoError(t, err)

	require.Len(t, fetchedEntries, 2)
	assert.Equal(t, entries[0].ID, fetchedEntries[0].ID)
	assert.Equal(t, entries[0].Name, fetchedEntries[0].Name)
	assert.Equal(t, entries[0].Seed, fetchedEntries[0].Seed)
	assert.Equal(t, *entries[0].EmbedLink, *fetchedEntries[0].EmbedLink)

	assert.Equal(t, entries[1].ID, fetchedEntries[1].ID)
	assert.Equal(t, entries[1].Name, fetchedEntries[1].Name)
	assert.Equal(t, entries[1].Seed, fetchedEntries[1].Seed)
	assert.Nil(t, fetchedEntries[1].EmbedLink)
}

func TestCreateMatches(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewTournamentStore(db)

	tournamentID := uuid.New()
	mockOwnerID := uuid.MustParse(testSuperUserID)
	tournament := &bracket.Tournament{
		ID:               tournamentID,
		OwnerID:          mockOwnerID,
		Name:             "Test Tournament",
		Status:           bracket.TournamentDraft,
		Type:             bracket.SingleElimination,
		ScoreRequirement: 0,
		CreatedAt:        time.Now().UTC(),
	}
	tx, err := db.BeginTxx(context.Background(), nil)
	require.NoError(t, err)
	err = store.CreateTournament(context.Background(), tx, tournament)
	require.NoError(t, err)
	err = tx.Commit()
	require.NoError(t, err)

	matchID1 := uuid.New()
	matchID2 := uuid.New()
	winnerNextMatchID := uuid.New()
	winnerNextSlot := 1

	matches := []bracket.Match{
		{
			ID:           matchID1,
			TournamentID: tournamentID,
			BracketSide:  bracket.WinnersSide,
			RoundNumber:  1,
			MatchOrder:   1,
			Status:       bracket.MatchPending,
		},
		{
			ID:                matchID2,
			TournamentID:      tournamentID,
			BracketSide:       bracket.WinnersSide,
			RoundNumber:       2,
			MatchOrder:        1,
			Status:            bracket.MatchPending,
			WinnerNextMatchID: &winnerNextMatchID,
			WinnerNextSlot:    &winnerNextSlot,
		},
		{
			ID:           winnerNextMatchID,
			TournamentID: tournamentID,
			BracketSide:  bracket.WinnersSide,
			RoundNumber:  3,
			MatchOrder:   1,
			Status:       bracket.MatchPending,
		},
	}

	tx, err = db.BeginTxx(context.Background(), nil)
	require.NoError(t, err)

	err = store.CreateMatches(context.Background(), tx, matches)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	var fetchedMatches []bracket.Match
	err = db.Select(&fetchedMatches, "SELECT * FROM matches WHERE tournament_id=$1 ORDER BY round_number, match_order", tournamentID)
	require.NoError(t, err)

	require.Len(t, fetchedMatches, 3)
	assert.Equal(t, matches[0].ID, fetchedMatches[0].ID)
	assert.Equal(t, matches[0].RoundNumber, fetchedMatches[0].RoundNumber)
	assert.Equal(t, matches[0].MatchOrder, fetchedMatches[0].MatchOrder)
	assert.Equal(t, matches[0].Status, fetchedMatches[0].Status)
	assert.Nil(t, fetchedMatches[0].WinnerNextMatchID)
	assert.Nil(t, fetchedMatches[0].WinnerNextSlot)

	assert.Equal(t, matches[1].ID, fetchedMatches[1].ID)
	assert.Equal(t, matches[1].RoundNumber, fetchedMatches[1].RoundNumber)
	assert.Equal(t, matches[1].MatchOrder, fetchedMatches[1].MatchOrder)
	assert.Equal(t, matches[1].Status, fetchedMatches[1].Status)
	assert.Equal(t, *matches[1].WinnerNextMatchID, *fetchedMatches[1].WinnerNextMatchID)
	assert.Equal(t, *matches[1].WinnerNextSlot, *fetchedMatches[1].WinnerNextSlot)

	assert.Equal(t, matches[2].ID, fetchedMatches[2].ID)
	assert.Equal(t, matches[2].RoundNumber, fetchedMatches[2].RoundNumber)
	assert.Equal(t, matches[2].MatchOrder, fetchedMatches[2].MatchOrder)
	assert.Equal(t, matches[2].Status, fetchedMatches[2].Status)
	assert.Nil(t, fetchedMatches[2].WinnerNextMatchID)
	assert.Nil(t, fetchedMatches[2].WinnerNextSlot)
}
