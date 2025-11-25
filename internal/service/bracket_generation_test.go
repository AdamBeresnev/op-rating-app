package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/AdamBeresnev/op-rating-app/internal/bracket"
	"github.com/AdamBeresnev/op-rating-app/internal/middleware"
	"github.com/AdamBeresnev/op-rating-app/internal/store"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates an in-memory SQLite database and applies migrations
func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	database, err := sqlx.Connect("sqlite3", "file::memory:")
	require.NoError(t, err, "Failed to connect to in-memory DB")

	_, err = database.Exec("PRAGMA foreign_keys = ON;")
	require.NoError(t, err)

	driver, err := sqlite3.WithInstance(database.DB, &sqlite3.Config{})
	require.NoError(t, err, "Failed to create migrate driver instance")

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"sqlite3",
		driver,
	)
	require.NoError(t, err, "Failed to create migrate instance")

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err, "Failed to apply migrations")
	}

	return database
}

func TestGenerateRound1SeedOrder(t *testing.T) {
	testCases := []struct {
		name       string
		numEntries int
		expected   [][2]int
	}{
		{
			name:       "2 entries",
			numEntries: 2,
			expected:   [][2]int{{0, 1}},
		},
		{
			name:       "4 entries",
			numEntries: 4,
			expected:   [][2]int{{0, 3}, {1, 2}},
		},
		{
			name:       "8 entries",
			numEntries: 8,
			expected:   [][2]int{{0, 7}, {3, 4}, {1, 6}, {2, 5}},
		},
		{
			name:       "Non-power of 2 (7 entries)",
			numEntries: 7,
			expected:   [][2]int{{0, 7}, {3, 4}, {1, 6}, {2, 5}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := generateRound1Pairs(tc.numEntries)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestCreateTournament(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tournamentStore := store.NewTournamentStore(db)
	bracketService := NewBracketService(db, tournamentStore)

	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserIDKey, uuid.MustParse(middleware.SuperUserID))

	testCases := []struct {
		name                    string
		tournamentName          string
		entryInputs             []EntryInput
		expectedTournamentCount int
		expectedEntryCount      int
		expectedMatchCount      int
		expectedError           bool
	}{
		{
			name:           "Successful tournament creation with 4 entries",
			tournamentName: "Test Tournament 4",
			entryInputs: []EntryInput{
				{Name: "Entry 1"}, {Name: "Entry 2"}, {Name: "Entry 3"}, {Name: "Entry 4"},
			},
			expectedTournamentCount: 1,
			expectedEntryCount:      4,
			expectedMatchCount:      3,
			expectedError:           false,
		},
		{
			name:           "Successful tournament creation with 5 entries",
			tournamentName: "Test Tournament 5",
			entryInputs: []EntryInput{
				{Name: "Entry 1"}, {Name: "Entry 2"}, {Name: "Entry 3"}, {Name: "Entry 4"}, {Name: "Entry 5"},
			},
			expectedTournamentCount: 1,
			expectedEntryCount:      5,
			expectedMatchCount:      7,
			expectedError:           false,
		},
		{
			name:           "Tournament creation with 1 entry",
			tournamentName: "Test Tournament 1",
			entryInputs: []EntryInput{
				{Name: "Solo Entry"},
			},
			expectedTournamentCount: 1,
			expectedEntryCount:      1,
			expectedMatchCount:      0,
			expectedError:           false,
		},
		{
			name:                    "Tournament creation with 0 entries",
			tournamentName:          "Test Tournament 0",
			entryInputs:             []EntryInput{},
			expectedTournamentCount: 1,
			expectedEntryCount:      0,
			expectedMatchCount:      0,
			expectedError:           false,
		},
	}

	fmt.Println(testCases)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			localExpectedEntryCount := tc.expectedEntryCount
			localExpectedMatchCount := tc.expectedMatchCount

			_, err := bracketService.CreateTournament(ctx, tc.tournamentName, tc.entryInputs)

			if tc.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify tournament creation
			var tournaments []bracket.Tournament
			err = db.Select(&tournaments, "SELECT * FROM tournaments WHERE name = ?", tc.tournamentName)
			require.NoError(t, err)
			assert.Len(t, tournaments, tc.expectedTournamentCount)
			require.GreaterOrEqual(t, len(tournaments), 1, "Expected at least one tournament to be created")
			createdTournament := tournaments[0]

			// Verify entries creation
			var entries []bracket.Entry
			err = db.Select(&entries, "SELECT * FROM entries WHERE tournament_id = ? ORDER BY seed", createdTournament.ID)
			require.NoError(t, err)
			assert.Equal(t, localExpectedEntryCount, len(entries))

			// Verify matches creation
			var matches []bracket.Match
			err = db.Select(&matches, "SELECT * FROM matches WHERE tournament_id = ?", createdTournament.ID)
			require.NoError(t, err)
			assert.Len(t, matches, localExpectedMatchCount)

			// Specific case for 4 entries because I couldn't be bothered writing all of them
			if len(tc.entryInputs) == 4 {
				for _, match := range matches {
					if match.RoundNumber == 1 {
						assert.NotNil(t, match.Entry1ID, "Entry1ID should not be nil for round 1 match")
						assert.NotNil(t, match.Entry2ID, "Entry2ID should not be nil for round 1 match")
					}
				}
			}
		})
	}
}
