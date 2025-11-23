package bracket

import (
	"time"

	"github.com/google/uuid"
)

type TournamentStatus string

const (
	TournamentDraft     TournamentStatus = "draft"
	TournamentStarted   TournamentStatus = "started"
	TournamentCompleted TournamentStatus = "completed"
)

type TournamentType string

const (
	SingleElimination TournamentType = "single"
	DoubleElimination TournamentType = "double"
)

type Tournament struct {
	ID               uuid.UUID        `db:"id"`
	OwnerID          uuid.UUID        `db:"owner_id"`
	Name             string           `db:"name" json:"name"`
	Status           TournamentStatus `db:"status"`
	Type             TournamentType   `db:"tournament_type"`
	ScoreRequirement int              `db:"score_requirement"`
	CreatedAt        time.Time        `db:"created_at"`
}
