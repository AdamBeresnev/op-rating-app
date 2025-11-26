package store

import (
	"context"
	"fmt"

	"github.com/AdamBeresnev/op-rating-app/internal/bracket"
	"github.com/AdamBeresnev/op-rating-app/internal/middleware"
	"github.com/jmoiron/sqlx"
)

type TournamentStore struct {
	db *sqlx.DB
}

func NewTournamentStore(db *sqlx.DB) *TournamentStore {
	return &TournamentStore{db: db}
}

func (s *TournamentStore) CreateTournament(ctx context.Context, tx *sqlx.Tx, tournament *bracket.Tournament) error {
	_, err := tx.NamedExecContext(ctx, `INSERT INTO tournaments (id, owner_id, name, status, tournament_type, score_requirement)
        VALUES (:id, :owner_id, :name, :status, :tournament_type, :score_requirement)`, tournament)
	return err
}

func (s *TournamentStore) CreateEntries(ctx context.Context, tx *sqlx.Tx, entries []bracket.Entry) error {
	if len(entries) == 0 {
		return nil
	}
	_, err := tx.NamedExecContext(ctx, `INSERT INTO entries (id, tournament_id, name, seed, embed_link)
            VALUES (:id, :tournament_id, :name, :seed, :embed_link)`, entries)
	return err
}

func (s *TournamentStore) CreateMatches(ctx context.Context, tx *sqlx.Tx, matches []bracket.Match) error {
	if len(matches) == 0 {
		return nil
	}
	_, err := tx.NamedExecContext(ctx, `INSERT INTO matches (id, tournament_id, bracket_side, round_number, match_order, entry_1_id, entry_2_id, status, winner_next_match_id, winner_next_slot)
		VALUES (:id, :tournament_id, :bracket_side, :round_number, :match_order, :entry_1_id, :entry_2_id, :status, :winner_next_match_id, :winner_next_slot)`, matches)
	return err
}

func (s *TournamentStore) GetTournament(ctx context.Context, id string) (*bracket.Tournament, error) {
	var tournament bracket.Tournament
	err := s.db.GetContext(ctx, &tournament, "SELECT * FROM tournaments WHERE id = ?", id)
	return &tournament, err
}

func (s *TournamentStore) GetTournamentsByUserID(ctx context.Context) ([]bracket.Tournament, error) {
	var tournaments []bracket.Tournament
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("user ID not found in the context")
	}
	err := s.db.SelectContext(ctx, &tournaments, "SELECT * FROM tournaments WHERE owner_id = ? ORDER BY created_at DESC", userID)
	return tournaments, err
}

func (s *TournamentStore) GetEntries(ctx context.Context, tournamentID string) ([]bracket.Entry, error) {
	var entries []bracket.Entry
	err := s.db.SelectContext(ctx, &entries, "SELECT * FROM entries WHERE tournament_id = ? ORDER BY seed ASC", tournamentID)
	return entries, err
}

func (s *TournamentStore) GetMatches(ctx context.Context, tournamentID string) ([]bracket.Match, error) {
	var matches []bracket.Match
	err := s.db.SelectContext(ctx, &matches, "SELECT * FROM matches WHERE tournament_id = ? ORDER BY round_number ASC, match_order ASC", tournamentID)
	return matches, err
}
