package service

import (
	"context"
	"testing"

	"github.com/AdamBeresnev/op-rating-app/internal/bracket"
	"github.com/AdamBeresnev/op-rating-app/internal/middleware"
	"github.com/AdamBeresnev/op-rating-app/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdvanceWinner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tournamentStore := store.NewTournamentStore(db)
	bracketService := NewTournamentService(db, tournamentStore)
	matchService := NewMatchService(db, tournamentStore)

	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserIDKey, uuid.MustParse(middleware.SuperUserID))

	// Create a tournament with 4 entries
	entryInputs := []EntryInput{
		{Name: "Entry 1"}, {Name: "Entry 2"}, {Name: "Entry 3"}, {Name: "Entry 4"},
	}
	tournamentID, err := bracketService.CreateTournament(ctx, "Test Tournament", entryInputs)
	require.NoError(t, err)

	matches, err := tournamentStore.GetMatches(ctx, tournamentID.String())
	require.NoError(t, err)
	require.Len(t, matches, 3)

	var match1, match2 *bracket.Match
	var match3 *bracket.Match

	for i := range matches {
		if matches[i].RoundNumber == 1 {
			if matches[i].MatchOrder == 1 {
				match1 = &matches[i]
			} else {
				match2 = &matches[i]
			}
		} else {
			match3 = &matches[i]
		}
	}

	require.NotNil(t, match1)
	require.NotNil(t, match2)
	require.NotNil(t, match3)

	entries, err := tournamentStore.GetEntries(ctx, tournamentID.String())
	require.NoError(t, err)
	require.Len(t, entries, 4)

	entry1 := entries[0] // Seed 1
	entry2 := entries[1] // Seed 2

	_, err = matchService.AdvanceWinner(ctx, match1.ID, entry1.ID)
	require.NoError(t, err)

	updatedMatch1, err := tournamentStore.GetMatch(ctx, match1.ID.String())
	require.NoError(t, err)
	assert.Equal(t, bracket.MatchFinished, updatedMatch1.Status)

	updatedMatch3, err := tournamentStore.GetMatch(ctx, match3.ID.String())
	require.NoError(t, err)

	assert.NotNil(t, updatedMatch3.Entry1ID)
	assert.Equal(t, entry1.ID, *updatedMatch3.Entry1ID)
	assert.Nil(t, updatedMatch3.Entry2ID)

	_, err = matchService.AdvanceWinner(ctx, match2.ID, entry2.ID)
	require.NoError(t, err)

	updatedMatch2, err := tournamentStore.GetMatch(ctx, match2.ID.String())
	require.NoError(t, err)
	assert.Equal(t, bracket.MatchFinished, updatedMatch2.Status)

	updatedMatch3, err = tournamentStore.GetMatch(ctx, match3.ID.String())
	require.NoError(t, err)

	assert.NotNil(t, updatedMatch3.Entry2ID)
	assert.Equal(t, entry2.ID, *updatedMatch3.Entry2ID)
}

func TestAdvanceWinner_EnforceOrder(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tournamentStore := store.NewTournamentStore(db)
	bracketService := NewTournamentService(db, tournamentStore)
	matchService := NewMatchService(db, tournamentStore)

	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserIDKey, uuid.MustParse(middleware.SuperUserID))

	// Create a tournament with 4 entries
	entryInputs := []EntryInput{
		{Name: "Entry 1"}, {Name: "Entry 2"}, {Name: "Entry 3"}, {Name: "Entry 4"},
	}
	tournamentID, err := bracketService.CreateTournament(ctx, "Test Tournament Order", entryInputs)
	require.NoError(t, err)

	matches, err := tournamentStore.GetMatches(ctx, tournamentID.String())
	require.NoError(t, err)
	require.Len(t, matches, 3)

	match1 := matches[0]
	match2 := matches[1]
	match3 := matches[2]

	entries, err := tournamentStore.GetEntries(ctx, tournamentID.String())
	require.NoError(t, err)
	entry1 := entries[0]
	entry3 := entries[2]

	// This should fail due to Match 2 being before Match 1
	_, err = matchService.AdvanceWinner(ctx, match2.ID, entry3.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be decided in order")

	_, err = matchService.AdvanceWinner(ctx, match1.ID, entry1.ID)
	require.NoError(t, err)

	_, err = matchService.AdvanceWinner(ctx, match3.ID, entry1.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be decided in order")

	_, err = matchService.AdvanceWinner(ctx, match2.ID, entry3.ID)
	require.NoError(t, err)

	_, err = matchService.AdvanceWinner(ctx, match3.ID, entry1.ID)
	require.NoError(t, err)
}
