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

var ErrNotEnoughEntries = fmt.Errorf("tournament must have at least 2 entries")

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

	nextMatch, err := s.store.GetNextPendingMatch(ctx, id)
	if err != nil {
		return nil, err
	}

	var nextMatchID *uuid.UUID
	if nextMatch != nil {
		id := nextMatch.ID
		nextMatchID = &id
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
	if bracketSize < 2 {
		return matches
	}

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

// This sucked
func (s *TournamentService) GenerateDoubleElimBracket(tournamentID uuid.UUID, entries []bracket.Entry) []bracket.Match {
	var matches []bracket.Match

	bracketSize := calcBracketSize(len(entries))
	if bracketSize < 2 {
		return matches
	}

	totalWBRounds := int(math.Log2(float64(bracketSize)))

	wbMap := make(map[int]map[int]*bracket.Match)
	lbMap := make(map[int]map[int]*bracket.Match)

	for r := 1; r <= totalWBRounds; r++ {
		wbMap[r] = make(map[int]*bracket.Match)
		matchesInRound := int(math.Pow(2, float64(totalWBRounds-r)))

		for i := 1; i <= matchesInRound; i++ {
			m := &bracket.Match{
				ID:           uuid.New(),
				TournamentID: tournamentID,
				BracketSide:  bracket.WinnersSide,
				RoundNumber:  r,
				MatchOrder:   i,
				Status:       bracket.MatchPending,
			}
			wbMap[r][i] = m
		}
	}

	totalLBRounds := 0
	if totalWBRounds > 1 {
		totalLBRounds = 2 * (totalWBRounds - 1)
	}

	for r := 1; r <= totalLBRounds; r++ {
		lbMap[r] = make(map[int]*bracket.Match)
		effectiveK := (r + 1) / 2
		matchesInRound := bracketSize / int(math.Pow(2, float64(effectiveK+1)))

		for i := 1; i <= matchesInRound; i++ {
			m := &bracket.Match{
				ID:           uuid.New(),
				TournamentID: tournamentID,
				BracketSide:  bracket.LosersSide,
				RoundNumber:  r,
				MatchOrder:   i,
				Status:       bracket.MatchPending,
			}
			lbMap[r][i] = m
		}
	}

	// grandfinals
	gf := &bracket.Match{
		ID:           uuid.New(),
		TournamentID: tournamentID,
		BracketSide:  bracket.FinalsSide,
		RoundNumber:  1,
		MatchOrder:   1,
		Status:       bracket.MatchPending,
	}

	for r := 1; r <= totalWBRounds; r++ {
		for i, m := range wbMap[r] {
			if r < totalWBRounds {
				parentOrder := (i + 1) / 2
				if parentMatch, ok := wbMap[r+1][parentOrder]; ok {
					m.WinnerNextMatchID = &parentMatch.ID
					if i%2 != 0 {
						m.WinnerNextSlot = utils.Ptr(1)
					} else {
						m.WinnerNextSlot = utils.Ptr(2)
					}
				}
			} else {
				m.WinnerNextMatchID = &gf.ID
				m.WinnerNextSlot = utils.Ptr(1)
			}
		}
	}

	for r := 1; r <= totalLBRounds; r++ {
		for i, m := range lbMap[r] {
			if r < totalLBRounds {
				var nextMatch *bracket.Match
				var nextSlot int

				if r%2 != 0 {
					// Odd Round: Winner goes to Next Round (Even), same match index
					nextMatch = lbMap[r+1][i]
					nextSlot = 1
				} else {
					// Even Round: Winner goes to Next Round (Odd), matches halve
					nextOrder := (i + 1) / 2
					nextMatch = lbMap[r+1][nextOrder]
					if i%2 != 0 {
						nextSlot = 1
					} else {
						nextSlot = 2
					}
				}

				if nextMatch != nil {
					m.WinnerNextMatchID = &nextMatch.ID
					m.WinnerNextSlot = &nextSlot
				}
			} else {
				// LB Final -> Grand Final Slot 2
				m.WinnerNextMatchID = &gf.ID
				m.WinnerNextSlot = utils.Ptr(2)
			}
		}
	}

	if totalLBRounds > 0 {
		// WB R1 Losers -> LB R1
		for i, m := range wbMap[1] {
			targetOrder := (i + 1) / 2
			if targetMatch, ok := lbMap[1][targetOrder]; ok {
				m.LoserNextMatchID = &targetMatch.ID
				if i%2 != 0 {
					m.LoserNextSlot = utils.Ptr(1)
				} else {
					m.LoserNextSlot = utils.Ptr(2)
				}
			}
		}

		for r := 2; r <= totalWBRounds; r++ {
			lbTargetRound := 2 * (r - 1)

			if lbTargetRound > totalLBRounds {
				continue
			}

			matchCount := len(wbMap[r])
			for i, m := range wbMap[r] {
				// Reverse mapping to avoid immediate rematches
				targetOrder := matchCount - i + 1
				if targetMatch, ok := lbMap[lbTargetRound][targetOrder]; ok {
					m.LoserNextMatchID = &targetMatch.ID
					m.LoserNextSlot = utils.Ptr(2) // Slot 1 is taken by LB winner
				}
			}
		}
	} else {
		// Special case if n=2, but who makes 2 person double elim brackets anyways?
		// WB Final Loser -> Grand Final Slot 2
		if totalWBRounds == 1 {
			wbFinal := wbMap[1][1]
			wbFinal.LoserNextMatchID = &gf.ID
			wbFinal.LoserNextSlot = utils.Ptr(2)
		}
	}

	// Flatten everything to a single slice
	for r := 1; r <= totalWBRounds; r++ {
		for i := 1; i <= len(wbMap[r]); i++ {
			matches = append(matches, *wbMap[r][i])
		}
	}
	for r := 1; r <= totalLBRounds; r++ {
		for i := 1; i <= len(lbMap[r]); i++ {
			matches = append(matches, *lbMap[r][i])
		}
	}
	matches = append(matches, *gf)

	return matches
}

func (s *TournamentService) CreateTournament(ctx context.Context, name string, tournamentType bracket.TournamentType, entryInputs []EntryInput) (uuid.UUID, error) {
	if len(entryInputs) < 2 {
		return uuid.Nil, ErrNotEnoughEntries
	}

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
		Type:             tournamentType,
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

	var matches []bracket.Match
	if tournament.Type == bracket.DoubleElimination {
		matches = s.GenerateDoubleElimBracket(tournamentID, entries)
	} else {
		matches = s.GenerateSingleElimBracket(tournamentID, entries)
	}

	if len(entries) > 1 {
		// Maps are surprisingly convenient in Go
		matchMap := make(map[uuid.UUID]*bracket.Match)
		for i := range matches {
			matchMap[matches[i].ID] = &matches[i]
		}

		var propagateByeInMemory func(*uuid.UUID)
		propagateByeInMemory = func(matchID *uuid.UUID) {
			if matchID == nil {
				return
			}
			if m, ok := matchMap[*matchID]; ok {
				if m.IsBye {
					// Bye -> match resolved
					m.Status = bracket.MatchFinished
					propagateByeInMemory(m.WinnerNextMatchID)
				} else {
					m.IsBye = true
				}
			}
		}

		round1Matches := make([]*bracket.Match, 0)
		for i := range matches {
			if matches[i].RoundNumber == 1 && matches[i].BracketSide == bracket.WinnersSide {
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
				if match.WinnerNextMatchID != nil {
					if nextMatch, ok := matchMap[*match.WinnerNextMatchID]; ok {
						if match.WinnerNextSlot != nil {
							if *match.WinnerNextSlot == 1 {
								nextMatch.Entry1ID = match.Entry1ID
							} else {
								nextMatch.Entry2ID = match.Entry1ID
							}
						}
					}
				}
				// Propagate BYE to Loser Bracket
				propagateByeInMemory(match.LoserNextMatchID)

			} else if match.Entry1ID == nil && match.Entry2ID != nil {
				match.Status = bracket.MatchFinished
				slot := 2
				match.WinnerSlot = &slot
				match.IsBye = true
				// Advance to next match
				if match.WinnerNextMatchID != nil {
					if nextMatch, ok := matchMap[*match.WinnerNextMatchID]; ok {
						if match.WinnerNextSlot != nil {
							if *match.WinnerNextSlot == 1 {
								nextMatch.Entry1ID = match.Entry2ID
							} else {
								nextMatch.Entry2ID = match.Entry2ID
							}
						}
					}
				}
				// Propagate BYE to Loser Bracket
				propagateByeInMemory(match.LoserNextMatchID)
			}
		}
	}

	if err := s.store.CreateMatches(ctx, tx, matches); err != nil {
		return uuid.Nil, err
	}

	return tournamentID, tx.Commit()
}
