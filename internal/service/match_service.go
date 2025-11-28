package service

import (
	"context"
	"fmt"

	"github.com/AdamBeresnev/op-rating-app/internal/bracket"
	"github.com/AdamBeresnev/op-rating-app/internal/store"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type MatchService struct {
	db    *sqlx.DB
	store *store.TournamentStore
}

func NewMatchService(db *sqlx.DB, store *store.TournamentStore) *MatchService {
	return &MatchService{db: db, store: store}
}

type MatchData struct {
	Match       *bracket.Match
	Entry1      *bracket.Entry
	Entry2      *bracket.Entry
	NextMatchID *uuid.UUID
}

func (s *MatchService) GetMatchViewData(ctx context.Context, matchIDStr string) (*MatchData, error) {
	match, err := s.store.GetMatch(ctx, matchIDStr)
	if err != nil {
		return nil, err
	}

	var entry1, entry2 *bracket.Entry
	if match.Entry1ID != nil {
		e, err := s.store.GetEntry(ctx, match.Entry1ID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get entry 1: %w", err)
		}
		entry1 = e
	}
	if match.Entry2ID != nil {
		e, err := s.store.GetEntry(ctx, match.Entry2ID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get entry 2: %w", err)
		}
		entry2 = e
	}

	nextMatch, err := s.store.GetNextPendingMatch(ctx, match.TournamentID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get next match: %w", err)
	}

	var nextMatchID *uuid.UUID
	if nextMatch != nil {
		id := nextMatch.ID
		nextMatchID = &id
	}

	return &MatchData{
		Match:       match,
		Entry1:      entry1,
		Entry2:      entry2,
		NextMatchID: nextMatchID,
	}, nil
}

func (s *MatchService) AdvanceWinner(ctx context.Context, matchID uuid.UUID, winnerEntryID uuid.UUID) (uuid.UUID, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback()

	match, err := s.store.GetMatchTx(ctx, tx, matchID.String())
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get match: %w", err)
	}

	// Verify order
	hasPending, err := s.store.HasPreviousPendingMatchesTx(ctx, tx, match.TournamentID.String(), match.RoundNumber, match.MatchOrder)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to check match order: %w", err)
	}
	if hasPending {
		return uuid.Nil, fmt.Errorf("matches must be decided in order")
	}

	// Verify winner is in the match
	if match.Entry1ID != nil && *match.Entry1ID == winnerEntryID {
		slot := 1
		match.WinnerSlot = &slot
	} else if match.Entry2ID != nil && *match.Entry2ID == winnerEntryID {
		slot := 2
		match.WinnerSlot = &slot
	} else {
		return uuid.Nil, fmt.Errorf("winner is not part of this match")
	}

	match.Status = bracket.MatchFinished

	if err := s.store.UpdateMatch(ctx, tx, match); err != nil {
		return uuid.Nil, fmt.Errorf("failed to update match: %w", err)
	}

	if match.WinnerNextMatchID != nil && match.WinnerNextSlot != nil {
		nextMatch, err := s.store.GetMatchTx(ctx, tx, match.WinnerNextMatchID.String())
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to get next match: %w", err)
		}

		switch *match.WinnerNextSlot {
		case 1:
			nextMatch.Entry1ID = &winnerEntryID
		case 2:
			nextMatch.Entry2ID = &winnerEntryID
		}

		if err := s.store.UpdateMatch(ctx, tx, nextMatch); err != nil {
			return uuid.Nil, fmt.Errorf("failed to update next match: %w", err)
		}
	} else {
		// If there is no next match, update tournament status to finished
		if err := s.store.UpdateTournamentStatusTx(ctx, tx, match.TournamentID.String(), bracket.TournamentCompleted); err != nil {
			return uuid.Nil, fmt.Errorf("failed to update tournament status: %w", err)
		}
	}

	return match.TournamentID, tx.Commit()
}
