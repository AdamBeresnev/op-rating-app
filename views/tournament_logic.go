package views

import (
	"sort"

	"github.com/AdamBeresnev/op-rating-app/internal/bracket"
	"github.com/google/uuid"
)

type BracketData struct {
	WBRounds       map[int][]bracket.Match
	WBRoundNums    []int
	LBRounds       map[int][]bracket.Match
	LBRoundNums    []int
	FinalRounds    map[int][]bracket.Match
	FinalRoundNums []int
	EntryMap       map[uuid.UUID]bracket.Entry
}

func PrepareBracketData(entries []bracket.Entry, matches []bracket.Match) BracketData {
	entryMap := make(map[uuid.UUID]bracket.Entry)
	for _, e := range entries {
		entryMap[e.ID] = e
	}

	wbRounds := make(map[int][]bracket.Match)
	lbRounds := make(map[int][]bracket.Match)
	finalRounds := make(map[int][]bracket.Match)

	var wbRoundNums []int
	var lbRoundNums []int
	var finalRoundNums []int

	for _, m := range matches {
		switch m.BracketSide {
		case bracket.WinnersSide:
			if _, exists := wbRounds[m.RoundNumber]; !exists {
				wbRoundNums = append(wbRoundNums, m.RoundNumber)
			}
			wbRounds[m.RoundNumber] = append(wbRounds[m.RoundNumber], m)
		case bracket.LosersSide:
			if _, exists := lbRounds[m.RoundNumber]; !exists {
				lbRoundNums = append(lbRoundNums, m.RoundNumber)
			}
			lbRounds[m.RoundNumber] = append(lbRounds[m.RoundNumber], m)
		case bracket.FinalsSide:
			if _, exists := finalRounds[m.RoundNumber]; !exists {
				finalRoundNums = append(finalRoundNums, m.RoundNumber)
			}
			finalRounds[m.RoundNumber] = append(finalRounds[m.RoundNumber], m)
		}
	}

	sort.Ints(wbRoundNums)
	sort.Ints(lbRoundNums)
	sort.Ints(finalRoundNums)

	sortRounds(wbRounds, wbRoundNums)
	sortRounds(lbRounds, lbRoundNums)
	sortRounds(finalRounds, finalRoundNums)

	return BracketData{
		WBRounds:       wbRounds,
		WBRoundNums:    wbRoundNums,
		LBRounds:       lbRounds,
		LBRoundNums:    lbRoundNums,
		FinalRounds:    finalRounds,
		FinalRoundNums: finalRoundNums,
		EntryMap:       entryMap,
	}
}

func sortRounds(rounds map[int][]bracket.Match, roundNums []int) {
	for _, r := range roundNums {
		sort.Slice(rounds[r], func(i, j int) bool {
			return rounds[r][i].MatchOrder < rounds[r][j].MatchOrder
		})
	}
}
