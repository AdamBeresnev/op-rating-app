package bracket

import "github.com/google/uuid"

type EntryType string

const (
	YoutubeLink     EntryType = "youtube"
	AnimeThemesLink EntryType = "animethemes"
)

type Entry struct {
	ID           uuid.UUID `db:"id"`
	TournamentID uuid.UUID `db:"tournament_id"`
	Name         string    `db:"name"`
	Seed         int       `db:"seed"`
	EmbedLink    *string   `db:"embed_link"`
	ImageLink    *string   `db:"image_link"`
	Type         EntryType `db:"entry_type"`
}
