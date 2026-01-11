// Package benchmarks provides benchmarks for comparing GC performance.
//
// To run with default GC:
//
//	go test -bench=. -benchmem ./benchmarks/...
//
// To run with greenteagc (Go 1.25+):
//
//	GOEXPERIMENT=greenteagc go test -bench=. -benchmem ./benchmarks/...
//
// To compare results:
//
//	go install golang.org/x/perf/cmd/benchstat@latest
//	go test -bench=. -benchmem -count=5 ./benchmarks/... > default.txt
//	GOEXPERIMENT=greenteagc go test -bench=. -benchmem -count=5 ./benchmarks/... > greenteagc.txt
//	benchstat default.txt greenteagc.txt
package benchmarks

import (
	"encoding/json"
	"fmt"
	"runtime"
	"testing"
)

// Card represents a Magic card for benchmarking.
type Card struct {
	ArenaID     int      `json:"arenaId"`
	Name        string   `json:"name"`
	ManaCost    string   `json:"manaCost"`
	CMC         int      `json:"cmc"`
	Types       []string `json:"types"`
	Subtypes    []string `json:"subtypes"`
	Keywords    []string `json:"keywords"`
	Colors      []string `json:"colors"`
	ColorID     []string `json:"colorIdentity"`
	Rarity      string   `json:"rarity"`
	SetCode     string   `json:"setCode"`
	Power       string   `json:"power"`
	Toughness   string   `json:"toughness"`
	OracleText  string   `json:"oracleText"`
	FlavorText  string   `json:"flavorText"`
	GIHWR       float64  `json:"gihwr"`
	ALSA        float64  `json:"alsa"`
	GIH         int      `json:"gih"`
	GamesPlayed int      `json:"gamesPlayed"`
}

// DraftPick represents a draft pick event.
type DraftPick struct {
	SessionID  string   `json:"sessionId"`
	PackNumber int      `json:"packNumber"`
	PickNumber int      `json:"pickNumber"`
	PackCards  []Card   `json:"packCards"`
	PickedCard Card     `json:"pickedCard"`
	PoolCards  []Card   `json:"poolCards"`
	Timestamp  int64    `json:"timestamp"`
	Ratings    []Rating `json:"ratings"`
}

// Rating represents a card rating.
type Rating struct {
	CardID int     `json:"cardId"`
	Score  float64 `json:"score"`
	Tier   string  `json:"tier"`
	Reason string  `json:"reason"`
}

// Match represents a match result.
type Match struct {
	MatchID       string   `json:"matchId"`
	EventType     string   `json:"eventType"`
	DeckID        string   `json:"deckId"`
	DeckCards     []Card   `json:"deckCards"`
	Result        string   `json:"result"`
	OpponentDeck  []Card   `json:"opponentDeck"`
	GameDurations []int    `json:"gameDurations"`
	MulliganInfo  []int    `json:"mulliganInfo"`
	PlaysRecorded []string `json:"playsRecorded"`
}

func makeCard(id int) Card {
	return Card{
		ArenaID:     id,
		Name:        "Test Card Name That Is Reasonably Long",
		ManaCost:    "{2}{W}{U}",
		CMC:         4,
		Types:       []string{"Creature", "Legendary"},
		Subtypes:    []string{"Human", "Wizard"},
		Keywords:    []string{"Flying", "Vigilance", "Ward 2"},
		Colors:      []string{"W", "U"},
		ColorID:     []string{"W", "U"},
		Rarity:      "rare",
		SetCode:     "TST",
		Power:       "3",
		Toughness:   "4",
		OracleText:  "When this creature enters, draw two cards. Whenever this creature attacks, you may tap target creature an opponent controls.",
		FlavorText:  "In the realm of testing, the benchmarks reign supreme.",
		GIHWR:       0.585,
		ALSA:        4.2,
		GIH:         1500,
		GamesPlayed: 1350,
	}
}

func makeDraftPick(pickNum int, poolSize int) DraftPick {
	packCards := make([]Card, 15)
	for i := range packCards {
		packCards[i] = makeCard(pickNum*100 + i)
	}

	poolCards := make([]Card, poolSize)
	for i := range poolCards {
		poolCards[i] = makeCard(i)
	}

	ratings := make([]Rating, 15)
	for i := range ratings {
		ratings[i] = Rating{
			CardID: pickNum*100 + i,
			Score:  float64(i) / 15.0,
			Tier:   "A",
			Reason: "Strong card with good synergy and high win rate in the format",
		}
	}

	return DraftPick{
		SessionID:  "test-session-id-12345",
		PackNumber: (pickNum / 15) + 1,
		PickNumber: (pickNum % 15) + 1,
		PackCards:  packCards,
		PickedCard: packCards[0],
		PoolCards:  poolCards,
		Timestamp:  1700000000000,
		Ratings:    ratings,
	}
}

func makeMatch(id int, deckSize int) Match {
	deckCards := make([]Card, deckSize)
	for i := range deckCards {
		deckCards[i] = makeCard(i)
	}

	opponentDeck := make([]Card, 20) // Observed cards
	for i := range opponentDeck {
		opponentDeck[i] = makeCard(1000 + i)
	}

	plays := make([]string, 50)
	for i := range plays {
		plays[i] = "Player cast Test Card Name That Is Reasonably Long targeting opponent's creature"
	}

	return Match{
		MatchID:       "match-id-12345",
		EventType:     "PremierDraft",
		DeckID:        "deck-id-67890",
		DeckCards:     deckCards,
		Result:        "win",
		OpponentDeck:  opponentDeck,
		GameDurations: []int{720, 540, 0},
		MulliganInfo:  []int{7, 6},
		PlaysRecorded: plays,
	}
}

// BenchmarkCollectionAllocation simulates loading a large card collection.
// This creates many small objects that stress the GC.
func BenchmarkCollectionAllocation(b *testing.B) {
	sizes := []int{1000, 5000, 10000}

	for _, size := range sizes {
		b.Run(sizeName(size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				cards := make([]Card, size)
				for j := range cards {
					cards[j] = makeCard(j)
				}
				runtime.KeepAlive(cards)
			}
		})
	}
}

// BenchmarkDraftSessionAllocation simulates processing draft picks.
// Each pick involves allocating pack cards, pool cards, and ratings.
func BenchmarkDraftSessionAllocation(b *testing.B) {
	pickCounts := []int{15, 30, 45} // 1 pack, 2 packs, full draft

	for _, picks := range pickCounts {
		b.Run(pickName(picks), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				draftPicks := make([]DraftPick, picks)
				for j := range draftPicks {
					poolSize := j // Pool grows with each pick
					draftPicks[j] = makeDraftPick(j, poolSize)
				}
				runtime.KeepAlive(draftPicks)
			}
		})
	}
}

// BenchmarkMatchHistoryAllocation simulates loading match history.
func BenchmarkMatchHistoryAllocation(b *testing.B) {
	matchCounts := []int{50, 200, 500}

	for _, count := range matchCounts {
		b.Run(matchName(count), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				matches := make([]Match, count)
				for j := range matches {
					matches[j] = makeMatch(j, 60) // 60-card deck
				}
				runtime.KeepAlive(matches)
			}
		})
	}
}

// BenchmarkJSONMarshal benchmarks JSON encoding which creates many temporaries.
func BenchmarkJSONMarshal(b *testing.B) {
	card := makeCard(1)
	pick := makeDraftPick(1, 10)
	match := makeMatch(1, 60)

	b.Run("Card", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			data, _ := json.Marshal(card)
			runtime.KeepAlive(data)
		}
	})

	b.Run("DraftPick", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			data, _ := json.Marshal(pick)
			runtime.KeepAlive(data)
		}
	})

	b.Run("Match", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			data, _ := json.Marshal(match)
			runtime.KeepAlive(data)
		}
	})

	b.Run("CardSlice100", func(b *testing.B) {
		cards := make([]Card, 100)
		for j := range cards {
			cards[j] = makeCard(j)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data, _ := json.Marshal(cards)
			runtime.KeepAlive(data)
		}
	})
}

// BenchmarkJSONUnmarshal benchmarks JSON decoding which creates the target objects.
func BenchmarkJSONUnmarshal(b *testing.B) {
	cardJSON, _ := json.Marshal(makeCard(1))
	pickJSON, _ := json.Marshal(makeDraftPick(1, 10))
	matchJSON, _ := json.Marshal(makeMatch(1, 60))

	cards := make([]Card, 100)
	for j := range cards {
		cards[j] = makeCard(j)
	}
	cardsJSON, _ := json.Marshal(cards)

	b.Run("Card", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var card Card
			_ = json.Unmarshal(cardJSON, &card)
		}
	})

	b.Run("DraftPick", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var pick DraftPick
			_ = json.Unmarshal(pickJSON, &pick)
		}
	})

	b.Run("Match", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var match Match
			_ = json.Unmarshal(matchJSON, &match)
		}
	})

	b.Run("CardSlice100", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var cards []Card
			_ = json.Unmarshal(cardsJSON, &cards)
		}
	})
}

// BenchmarkRatingCalculation simulates calculating ratings for many cards.
// This involves many small allocations for intermediate calculations.
func BenchmarkRatingCalculation(b *testing.B) {
	cardCounts := []int{100, 500, 1000}

	for _, count := range cardCounts {
		b.Run(sizeName(count), func(b *testing.B) {
			cards := make([]Card, count)
			for j := range cards {
				cards[j] = makeCard(j)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ratings := make([]Rating, len(cards))
				for j, card := range cards {
					// Simulate rating calculation with intermediate allocations
					score := calculateScore(card)
					tier := determineTier(score)
					reason := generateReason(card, score)

					ratings[j] = Rating{
						CardID: card.ArenaID,
						Score:  score,
						Tier:   tier,
						Reason: reason,
					}
				}
				runtime.KeepAlive(ratings)
			}
		})
	}
}

// BenchmarkMapOperations benchmarks map-heavy operations common in card lookups.
func BenchmarkMapOperations(b *testing.B) {
	sizes := []int{1000, 5000, 10000}

	for _, size := range sizes {
		b.Run(sizeName(size)+"_build", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m := make(map[int]Card, size)
				for j := 0; j < size; j++ {
					m[j] = makeCard(j)
				}
				runtime.KeepAlive(m)
			}
		})

		// Pre-build map for lookup benchmark
		m := make(map[int]Card, size)
		for j := 0; j < size; j++ {
			m[j] = makeCard(j)
		}

		b.Run(sizeName(size)+"_lookup", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				for j := 0; j < size; j++ {
					card := m[j]
					runtime.KeepAlive(card)
				}
			}
		})
	}
}

// BenchmarkSliceGrowth benchmarks slice append operations.
func BenchmarkSliceGrowth(b *testing.B) {
	sizes := []int{100, 1000, 5000}

	for _, size := range sizes {
		b.Run(sizeName(size)+"_append", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var cards []Card
				for j := 0; j < size; j++ {
					cards = append(cards, makeCard(j))
				}
				runtime.KeepAlive(cards)
			}
		})

		b.Run(sizeName(size)+"_preallocated", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				cards := make([]Card, 0, size)
				for j := 0; j < size; j++ {
					cards = append(cards, makeCard(j))
				}
				runtime.KeepAlive(cards)
			}
		})
	}
}

// BenchmarkConcurrentAllocation tests concurrent allocation patterns.
// Uses different parallelism levels to stress GC under concurrent load.
func BenchmarkConcurrentAllocation(b *testing.B) {
	// SetParallelism sets the number of goroutines to p * GOMAXPROCS.
	// So parallelism=2 with GOMAXPROCS=8 runs 16 goroutines.
	parallelismLevels := []int{1, 2, 4}
	itemsPerGoroutine := 1000

	for _, p := range parallelismLevels {
		b.Run(fmt.Sprintf("parallelism%dx", p), func(b *testing.B) {
			b.ReportAllocs()
			b.SetParallelism(p)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					cards := make([]Card, itemsPerGoroutine)
					for j := range cards {
						cards[j] = makeCard(j)
					}
					runtime.KeepAlive(cards)
				}
			})
		})
	}
}

// Helper functions for rating simulation
func calculateScore(card Card) float64 {
	score := card.GIHWR
	if card.CMC <= 3 {
		score += 0.05
	}
	if len(card.Keywords) > 0 {
		score += 0.02 * float64(len(card.Keywords))
	}
	return score
}

func determineTier(score float64) string {
	switch {
	case score >= 0.60:
		return "A+"
	case score >= 0.55:
		return "A"
	case score >= 0.50:
		return "B"
	case score >= 0.45:
		return "C"
	default:
		return "D"
	}
}

func generateReason(card Card, score float64) string {
	return "Card " + card.Name + " has a win rate of " + formatFloat(score) + " which indicates " + qualityDescription(score)
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.0f%%", f*100)
}

func qualityDescription(score float64) string {
	if score >= 0.55 {
		return "strong performance in the format"
	}
	return "average or below average performance"
}

func sizeName(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%dk", n/1000)
	}
	return fmt.Sprintf("%d", n)
}

func pickName(n int) string {
	return fmt.Sprintf("%dpicks", n)
}

func matchName(n int) string {
	return fmt.Sprintf("%dmatches", n)
}
