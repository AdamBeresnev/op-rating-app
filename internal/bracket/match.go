package bracket

import (
	"time"

	"github.com/google/uuid"
)

type MatchStatus string

const (
	MatchPending   MatchStatus = "pending"
	MatchScheduled MatchStatus = "in_progress"
	MatchFinished  MatchStatus = "finished"
)

type BracketSide string

const (
	WinnersSide BracketSide = "winners"
	LosersSide  BracketSide = "losers"
	FinalsSide  BracketSide = "finals"
)

type Match struct {
	ID           uuid.UUID `db:"id"`
	TournamentID uuid.UUID `db:"tournament_id"`

	// Position in the tournament for reconstructing the view
	BracketSide BracketSide `db:"bracket_side"`
	RoundNumber int         `db:"round_number"`
	MatchOrder  int         `db:"match_order"`

	Entry1ID *uuid.UUID `db:"entry_1_id"`
	Entry2ID *uuid.UUID `db:"entry_2_id"`

	Score1 int         `db:"score_1"`
	Score2 int         `db:"score_2"`
	Status MatchStatus `db:"status"`

	WinnerNextMatchID *uuid.UUID `db:"winner_next_match_id"`
	WinnerNextSlot    *int       `db:"winner_next_slot"`

	LoserNextMatchID *uuid.UUID `db:"loser_next_match_id"`
	LoserNextSlot    *int       `db:"loser_next_slot"`

	WinnerSlot *int `db:"winner_slot"`
	IsBye      bool `db:"is_bye"`

	CreatedAt time.Time `db:"created_at"`
}

func (m *Match) IsWinner(slot int) bool {
	return m.Status == MatchFinished && m.WinnerSlot != nil && *m.WinnerSlot == slot
}

func (m *Match) IsLoser(slot int) bool {
	return m.Status == MatchFinished && m.WinnerSlot != nil && *m.WinnerSlot != slot
}
