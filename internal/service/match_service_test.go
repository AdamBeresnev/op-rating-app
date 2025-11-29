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
	tournamentID, err := bracketService.CreateTournament(ctx, "Test Tournament", bracket.SingleElimination, entryInputs)
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
	tournamentID, err := bracketService.CreateTournament(ctx, "Test Tournament Order", bracket.SingleElimination, entryInputs)
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

func TestDoubleEliminationAdvancement(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tournamentStore := store.NewTournamentStore(db)
	bracketService := NewTournamentService(db, tournamentStore)
	matchService := NewMatchService(db, tournamentStore)

	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserIDKey, uuid.MustParse(middleware.SuperUserID))

	entryInputs := []EntryInput{
		{Name: "1"}, {Name: "2"}, {Name: "3"}, {Name: "4"},
	}
	tID, err := bracketService.CreateTournament(ctx, "Double Elim Test", bracket.DoubleElimination, entryInputs)
	require.NoError(t, err)

	findMatch := func(side bracket.BracketSide, round, order int) *bracket.Match {
		matches, _ := tournamentStore.GetMatches(ctx, tID.String())
		for _, m := range matches {
			if m.BracketSide == side && m.RoundNumber == round && m.MatchOrder == order {
				return &m
			}
		}
		return nil
	}

	// WB Round 1 Match 1: Seed 1 vs Seed 4
	wbR1M1 := findMatch(bracket.WinnersSide, 1, 1)
	require.NotNil(t, wbR1M1)
	require.NotNil(t, wbR1M1.Entry1ID)
	require.NotNil(t, wbR1M1.Entry2ID)

	// Advance Seed 1 (Winner) -> WB R2 Match 1
	// Advance Seed 4 (Loser)  -> LB R1 Match 1
	winnerID := *wbR1M1.Entry1ID
	loserID := *wbR1M1.Entry2ID

	_, err = matchService.AdvanceWinner(ctx, wbR1M1.ID, winnerID)
	require.NoError(t, err)

	// Verify Winner moved to WB R2
	wbR2M1 := findMatch(bracket.WinnersSide, 2, 1)
	require.NotNil(t, wbR2M1.Entry1ID)
	assert.Equal(t, winnerID, *wbR2M1.Entry1ID)

	// Verify Loser moved to LB R1
	lbR1M1 := findMatch(bracket.LosersSide, 1, 1)
	require.NotNil(t, lbR1M1)
	assert.True(t, lbR1M1.Entry1ID != nil || lbR1M1.Entry2ID != nil, "Loser should be in one of the slots")

	hasLoser := (lbR1M1.Entry1ID != nil && *lbR1M1.Entry1ID == loserID) || (lbR1M1.Entry2ID != nil && *lbR1M1.Entry2ID == loserID)
	assert.True(t, hasLoser, "Loser should be in LB R1 Match 1")
}

func TestDoubleElimination_ByeHandling(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tournamentStore := store.NewTournamentStore(db)
	bracketService := NewTournamentService(db, tournamentStore)
	matchService := NewMatchService(db, tournamentStore)

	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserIDKey, uuid.MustParse(middleware.SuperUserID))

	// 5 entries -> 8 slots.
	// Seeds: 1, 2, 3, 4, 5.
	// Pairs: 1v8(Bye), 4v5, 3v6(Bye), 2v7(Bye).
	entryInputs := []EntryInput{
		{Name: "1"}, {Name: "2"}, {Name: "3"}, {Name: "4"}, {Name: "5"},
	}
	tID, err := bracketService.CreateTournament(ctx, "Bye Test", bracket.DoubleElimination, entryInputs)
	require.NoError(t, err)

	entries, err := tournamentStore.GetEntries(ctx, tID.String())
	require.NoError(t, err)

	// Java could never btw
	findMatch := func(side bracket.BracketSide, round, order int) *bracket.Match {
		matches, _ := tournamentStore.GetMatches(ctx, tID.String())
		for _, m := range matches {
			if m.BracketSide == side && m.RoundNumber == round && m.MatchOrder == order {
				return &m
			}
		}
		return nil
	}

	var p5ID uuid.UUID
	for _, e := range entries {
		if e.Seed == 5 {
			p5ID = e.ID
		}
	}

	// WB R1 M2 is P4 vs P5
	wbR1M2 := findMatch(bracket.WinnersSide, 1, 2)
	require.NotNil(t, wbR1M2)

	lbR2M2 := findMatch(bracket.LosersSide, 2, 2)
	require.NotNil(t, lbR2M2)
	assert.True(t, lbR2M2.IsBye, "LB R2 M2 should be marked as Bye due to double-bye propagation")

	winnerID := *wbR1M2.Entry1ID
	if winnerID == p5ID {
		winnerID = *wbR1M2.Entry2ID
	} // ensure P4 wins

	_, err = matchService.AdvanceWinner(ctx, wbR1M2.ID, winnerID)
	require.NoError(t, err)

	// Verify P5 is in LB R2 M1
	lbR2M1 := findMatch(bracket.LosersSide, 2, 1)
	require.NotNil(t, lbR2M1)

	hasP5 := (lbR2M1.Entry1ID != nil && *lbR2M1.Entry1ID == p5ID) || (lbR2M1.Entry2ID != nil && *lbR2M1.Entry2ID == p5ID)
	assert.True(t, hasP5, "P5 should have auto-advanced to LB R2 M1")
}
