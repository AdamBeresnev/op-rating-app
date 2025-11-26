package service

import (
	"context"
	"fmt"
	"math"

	"github.com/AdamBeresnev/op-rating-app/internal/bracket"
	"github.com/AdamBeresnev/op-rating-app/internal/middleware"
	"github.com/AdamBeresnev/op-rating-app/internal/store"
	"github.com/AdamBeresnev/op-rating-app/internal/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TournamentService struct {
	db    *sqlx.DB
	store *store.TournamentStore
}

func NewTournamentService(db *sqlx.DB, store *store.TournamentStore) *TournamentService {
	return &TournamentService{db: db, store: store}
}

type EntryInput struct {
	Name      string
	EmbedLink string
}

type TournamentData struct {
	Tournament  *bracket.Tournament
	Entries     []bracket.Entry
	Matches     []bracket.Match
	NextMatchID *uuid.UUID
}

func (s *TournamentService) GetTournamentData(ctx context.Context, id string) (*TournamentData, error) {
	tournament, err := s.store.GetTournament(ctx, id)
	if err != nil {
		return nil, err
	}

	entries, err := s.store.GetEntries(ctx, id)
	if err != nil {
		return nil, err
	}

	matches, err := s.store.GetMatches(ctx, id)
	if err != nil {
		return nil, err
	}

	var nextMatchID *uuid.UUID
	for _, m := range matches {
		if m.Status != bracket.MatchFinished {
			id := m.ID
			nextMatchID = &id
			break
		}
	}

	return &TournamentData{
		Tournament:  tournament,
		Entries:     entries,
		Matches:     matches,
		NextMatchID: nextMatchID,
	}, nil
}

func (s *TournamentService) GetTournamentsForUser(ctx context.Context) ([]bracket.Tournament, error) {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("user ID not found in the context")
	}
	return s.store.GetTournamentsByUserID(ctx, userID)
}

// Gets the nearest power of 2 while rounding up, so with input 5 it returns 8 and so on
func calcBracketSize(count int) int {
	if count <= 0 {
		return 0
	}

	// Log2 -> Ceil -> 2^^log2 to round up
	log2 := math.Ceil(math.Log2(float64(count)))
	return int(math.Pow(2, log2))
}

func generateRound1Pairs(bracketSize int) [][2]int {
	if bracketSize == 0 {
		return [][2]int{}
	}

	rounds := []int{0}
	for len(rounds) < bracketSize {
		var nextRound []int
		currentCount := len(rounds) * 2

		for _, seed := range rounds {
			nextRound = append(nextRound, seed)
			nextRound = append(nextRound, (currentCount-1)-seed)
		}
		rounds = nextRound
	}

	pairs := make([][2]int, 0, bracketSize/2)
	for i := 0; i < len(rounds); i += 2 {
		matchup := [2]int{rounds[i], rounds[i+1]}
		pairs = append(pairs, matchup)
	}

	return pairs
}

// Generate bracket structure for single elimination
func (s *TournamentService) GenerateSingleElimBracket(tournamentID uuid.UUID, entries []bracket.Entry) []bracket.Match {
	var matches []bracket.Match

	bracketSize := calcBracketSize(len(entries))
	totalRounds := int(math.Log2(float64(bracketSize)))

	nextRoundMatchIDs := make(map[int]uuid.UUID)

	// Significantly easier to start from the last round and work backwards
	for r := totalRounds; r >= 1; r-- {
		matchesInCurrentRound := int(math.Pow(2, float64(totalRounds-r)))
		currentRoundMatchIDs := make(map[int]uuid.UUID)

		for i := 0; i < matchesInCurrentRound; i++ {
			matchID := uuid.New()
			matchOrder := i + 1

			m := bracket.Match{
				ID:           matchID,
				TournamentID: tournamentID,
				BracketSide:  bracket.WinnersSide,
				RoundNumber:  r,
				MatchOrder:   matchOrder,
				Status:       bracket.MatchPending,
			}

			if r < totalRounds {
				parentMatchOrder := (matchOrder + 1) / 2
				parentID := nextRoundMatchIDs[parentMatchOrder]

				m.WinnerNextMatchID = &parentID

				if matchOrder%2 != 0 {
					m.WinnerNextSlot = utils.Ptr(1)
				} else {
					m.WinnerNextSlot = utils.Ptr(2)
				}
			}

			matches = append(matches, m)
			currentRoundMatchIDs[matchOrder] = matchID
		}
		nextRoundMatchIDs = currentRoundMatchIDs
	}

	return matches
}

func (s *TournamentService) CreateTournament(ctx context.Context, name string, entryInputs []EntryInput) (uuid.UUID, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback()

	tournamentID := uuid.New()
	ownerID, _ := middleware.GetUserIDFromContext(ctx)
	tournament := bracket.Tournament{
		ID:               tournamentID,
		OwnerID:          ownerID,
		Name:             name,
		Status:           bracket.TournamentStarted,
		Type:             bracket.SingleElimination,
		ScoreRequirement: 0,
	}

	if err := s.store.CreateTournament(ctx, tx, &tournament); err != nil {
		return uuid.Nil, err
	}

	var entries []bracket.Entry
	for i, input := range entryInputs {
		e := bracket.Entry{
			ID:           uuid.New(),
			TournamentID: tournamentID,
			Name:         input.Name,
			Seed:         i + 1,
			EmbedLink:    utils.StringOrNil(input.EmbedLink),
		}

		entries = append(entries, e)
	}

	if err := s.store.CreateEntries(ctx, tx, entries); err != nil {
		return uuid.Nil, err
	}

	matches := s.GenerateSingleElimBracket(tournamentID, entries)

	if len(entries) > 1 {
		// Create a map for easy lookup to propagate winners
		matchMap := make(map[uuid.UUID]*bracket.Match)
		for i := range matches {
			matchMap[matches[i].ID] = &matches[i]
		}

		round1Matches := make([]*bracket.Match, 0)
		for i := range matches {
			if matches[i].RoundNumber == 1 {
				round1Matches = append(round1Matches, &matches[i])
			}
		}

		pairings := generateRound1Pairs(calcBracketSize(len(entries)))
		for i, pair := range pairings {
			if i >= len(round1Matches) {
				break
			}
			match := round1Matches[i]
			if pair[0] < len(entries) {
				match.Entry1ID = &entries[pair[0]].ID
			}
			if pair[1] < len(entries) {
				match.Entry2ID = &entries[pair[1]].ID
			}

			// Check for byes immediately
			if match.Entry1ID != nil && match.Entry2ID == nil {
				match.Status = bracket.MatchFinished
				slot := 1
				                match.WinnerSlot = &slot
				                match.IsBye = true
				                				// Advance to next match
				                				if match.WinnerNextMatchID != nil {					if nextMatch, ok := matchMap[*match.WinnerNextMatchID]; ok {
						if match.WinnerNextSlot != nil {
							if *match.WinnerNextSlot == 1 {
								nextMatch.Entry1ID = match.Entry1ID
							} else {
								nextMatch.Entry2ID = match.Entry1ID
							}
						}
					}
				}
			} else if match.Entry1ID == nil && match.Entry2ID != nil {
				match.Status = bracket.MatchFinished
				slot := 2
				                match.WinnerSlot = &slot
				                match.IsBye = true
				                				// Advance to next match
				                				if match.WinnerNextMatchID != nil {					if nextMatch, ok := matchMap[*match.WinnerNextMatchID]; ok {
						if match.WinnerNextSlot != nil {
							if *match.WinnerNextSlot == 1 {
								nextMatch.Entry1ID = match.Entry2ID
							} else {
								nextMatch.Entry2ID = match.Entry2ID
							}
						}
					}
				}
			}
		}
	}

	if err := s.store.CreateMatches(ctx, tx, matches); err != nil {
		return uuid.Nil, err
	}

	return tournamentID, tx.Commit()
}
