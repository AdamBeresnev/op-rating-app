package store

import (
	"context"
	"database/sql"

	"github.com/AdamBeresnev/op-rating-app/internal/bracket"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TournamentStore struct {
	db *sqlx.DB
}

const (
	createTournamentQuery = `INSERT INTO tournaments (id, owner_id, name, status, tournament_type, score_requirement)
        VALUES (:id, :owner_id, :name, :status, :tournament_type, :score_requirement)`
	createEntriesQuery = `INSERT INTO entries (id, tournament_id, name, seed, embed_link)
            VALUES (:id, :tournament_id, :name, :seed, :embed_link)`
	createMatchesQuery = `INSERT INTO matches (id, tournament_id, bracket_side, round_number, match_order, entry_1_id, entry_2_id, status, winner_next_match_id, winner_next_slot, winner_slot, is_bye)
		VALUES (:id, :tournament_id, :bracket_side, :round_number, :match_order, :entry_1_id, :entry_2_id, :status, :winner_next_match_id, :winner_next_slot, :winner_slot, :is_bye)`
	getTournamentQuery        = "SELECT * FROM tournaments WHERE id = ?"
	getTournamentsByUserQuery = "SELECT * FROM tournaments WHERE owner_id = ? ORDER BY created_at DESC"
	getEntriesQuery           = "SELECT * FROM entries WHERE tournament_id = ? ORDER BY seed ASC"
	getEntryQuery             = "SELECT * FROM entries WHERE id = ?"
	getMatchesQuery           = "SELECT * FROM matches WHERE tournament_id = ? ORDER BY round_number ASC, match_order ASC"
	getMatchQuery             = "SELECT * FROM matches WHERE id = ?"
	updateMatchQuery          = `UPDATE matches SET
		tournament_id = :tournament_id,
		bracket_side = :bracket_side,
		round_number = :round_number,
		match_order = :match_order,
		entry_1_id = :entry_1_id,
		entry_2_id = :entry_2_id,
		score_1 = :score_1,
		score_2 = :score_2,
		status = :status,
		winner_next_match_id = :winner_next_match_id,
		winner_next_slot = :winner_next_slot,
		loser_next_match_id = :loser_next_match_id,
		loser_next_slot = :loser_next_slot,
		winner_slot = :winner_slot,
		is_bye = :is_bye
		WHERE id = :id`
	hasPreviousPendingMatchesQuery = `SELECT count(*) FROM matches 
		WHERE tournament_id = ? 
		AND status != 'finished'
		AND (round_number < ? OR (round_number = ? AND match_order < ?))`
	getNextPendingMatchQuery = `SELECT * FROM matches 
		WHERE tournament_id = ? 
		AND status != 'finished' 
		ORDER BY round_number ASC, match_order ASC 
		LIMIT 1`
	updateTournamentStatusQuery = "UPDATE tournaments SET status = ? WHERE id = ?"
)

func NewTournamentStore(db *sqlx.DB) *TournamentStore {
	return &TournamentStore{db: db}
}

func (s *TournamentStore) CreateTournament(ctx context.Context, tx *sqlx.Tx, tournament *bracket.Tournament) error {
	_, err := tx.NamedExecContext(ctx, createTournamentQuery, tournament)
	return err
}

func (s *TournamentStore) CreateEntries(ctx context.Context, tx *sqlx.Tx, entries []bracket.Entry) error {
	if len(entries) == 0 {
		return nil
	}
	_, err := tx.NamedExecContext(ctx, createEntriesQuery, entries)
	return err
}

func (s *TournamentStore) CreateMatches(ctx context.Context, tx *sqlx.Tx, matches []bracket.Match) error {
	if len(matches) == 0 {
		return nil
	}
	_, err := tx.NamedExecContext(ctx, createMatchesQuery, matches)
	return err
}

func (s *TournamentStore) GetTournament(ctx context.Context, id string) (*bracket.Tournament, error) {
	var tournament bracket.Tournament
	err := s.db.GetContext(ctx, &tournament, getTournamentQuery, id)
	return &tournament, err
}

func (s *TournamentStore) GetTournamentsByUserID(ctx context.Context, userID uuid.UUID) ([]bracket.Tournament, error) {
	var tournaments []bracket.Tournament
	err := s.db.SelectContext(ctx, &tournaments, getTournamentsByUserQuery, userID)
	return tournaments, err
}

func (s *TournamentStore) GetEntries(ctx context.Context, tournamentID string) ([]bracket.Entry, error) {
	var entries []bracket.Entry
	err := s.db.SelectContext(ctx, &entries, getEntriesQuery, tournamentID)
	return entries, err
}

func (s *TournamentStore) GetEntry(ctx context.Context, id string) (*bracket.Entry, error) {
	var entry bracket.Entry
	err := s.db.GetContext(ctx, &entry, getEntryQuery, id)
	return &entry, err
}

func (s *TournamentStore) GetMatches(ctx context.Context, tournamentID string) ([]bracket.Match, error) {
	var matches []bracket.Match
	err := s.db.SelectContext(ctx, &matches, getMatchesQuery, tournamentID)
	return matches, err
}

func (s *TournamentStore) GetMatch(ctx context.Context, id string) (*bracket.Match, error) {
	var match bracket.Match
	err := s.db.GetContext(ctx, &match, getMatchQuery, id)
	return &match, err
}

func (s *TournamentStore) GetMatchTx(ctx context.Context, tx *sqlx.Tx, id string) (*bracket.Match, error) {
	var match bracket.Match
	err := tx.GetContext(ctx, &match, getMatchQuery, id)
	return &match, err
}

func (s *TournamentStore) UpdateMatch(ctx context.Context, tx *sqlx.Tx, match *bracket.Match) error {
	_, err := tx.NamedExecContext(ctx, updateMatchQuery, match)
	return err
}

func (s *TournamentStore) HasPreviousPendingMatchesTx(ctx context.Context, tx *sqlx.Tx, tournamentID string, roundNumber int, matchOrder int) (bool, error) {
	var count int
	err := tx.GetContext(ctx, &count, hasPreviousPendingMatchesQuery, tournamentID, roundNumber, roundNumber, matchOrder)
	return count > 0, err
}

func (s *TournamentStore) GetNextPendingMatch(ctx context.Context, tournamentID string) (*bracket.Match, error) {
	var match bracket.Match
	err := s.db.GetContext(ctx, &match, getNextPendingMatchQuery, tournamentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &match, nil
}

func (s *TournamentStore) UpdateTournamentStatusTx(ctx context.Context, tx *sqlx.Tx, tournamentID string, status bracket.TournamentStatus) error {
	_, err := tx.ExecContext(ctx, updateTournamentStatusQuery, status, tournamentID)
	return err
}
