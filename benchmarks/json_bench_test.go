//go:build goexperiment.jsonv2

// Package benchmarks provides JSON v1 vs v2 benchmarks.
//
// These benchmarks require Go 1.25+ with the jsonv2 experiment enabled.
//
// To run:
//
//	GOEXPERIMENT=jsonv2 go test -bench=BenchmarkJSON -benchmem ./benchmarks/...
//
// To compare v1 vs v2:
//
//	./benchmarks/run_json_comparison.sh
package benchmarks

import (
	"bytes"
	"encoding/json"
	jsonv2 "encoding/json/v2"
	"runtime"
	"testing"
)

// jsonTestCard is a card structure for JSON benchmarking.
type jsonTestCard struct {
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
	Power       string   `json:"power,omitempty"`
	Toughness   string   `json:"toughness,omitempty"`
	OracleText  string   `json:"oracleText"`
	FlavorText  string   `json:"flavorText,omitempty"`
	GIHWR       float64  `json:"gihwr"`
	ALSA        float64  `json:"alsa"`
	GIH         int      `json:"gih"`
	GamesPlayed int      `json:"gamesPlayed"`
}

// jsonTestDeck is a deck structure for JSON benchmarking.
type jsonTestDeck struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Format      string         `json:"format"`
	MainDeck    []jsonTestCard `json:"mainDeck"`
	Sideboard   []jsonTestCard `json:"sideboard"`
	Colors      []string       `json:"colors"`
	CreatedAt   int64          `json:"createdAt"`
	LastPlayed  int64          `json:"lastPlayed"`
	WinRate     float64        `json:"winRate"`
	MatchCount  int            `json:"matchCount"`
	Description string         `json:"description"`
}

// jsonTestDraftSession is a draft session for JSON benchmarking.
type jsonTestDraftSession struct {
	SessionID    string                `json:"sessionId"`
	EventType    string                `json:"eventType"`
	SetCode      string                `json:"setCode"`
	Picks        []jsonTestDraftPick   `json:"picks"`
	FinalPool    []jsonTestCard        `json:"finalPool"`
	DeckBuilt    []jsonTestCard        `json:"deckBuilt"`
	MatchResults []jsonTestMatchResult `json:"matchResults"`
	StartedAt    int64                 `json:"startedAt"`
	CompletedAt  int64                 `json:"completedAt"`
}

type jsonTestDraftPick struct {
	PackNumber int            `json:"packNumber"`
	PickNumber int            `json:"pickNumber"`
	PackCards  []jsonTestCard `json:"packCards"`
	PickedCard jsonTestCard   `json:"pickedCard"`
	Timestamp  int64          `json:"timestamp"`
}

type jsonTestMatchResult struct {
	MatchID    string `json:"matchId"`
	Result     string `json:"result"`
	GameWins   int    `json:"gameWins"`
	GameLosses int    `json:"gameLosses"`
}

func makeJSONTestCard(id int) jsonTestCard {
	return jsonTestCard{
		ArenaID:     id,
		Name:        "Test Card With A Reasonably Long Name",
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
		OracleText:  "When this creature enters, draw two cards. Whenever this creature attacks, you may tap target creature.",
		FlavorText:  "Testing the limits of serialization.",
		GIHWR:       0.585,
		ALSA:        4.2,
		GIH:         1500,
		GamesPlayed: 1350,
	}
}

func makeJSONTestDeck(size int) jsonTestDeck {
	mainDeck := make([]jsonTestCard, size)
	for i := range mainDeck {
		mainDeck[i] = makeJSONTestCard(i)
	}

	sideboard := make([]jsonTestCard, 15)
	for i := range sideboard {
		sideboard[i] = makeJSONTestCard(1000 + i)
	}

	return jsonTestDeck{
		ID:          "deck-12345-67890",
		Name:        "Test Deck With Description",
		Format:      "Standard",
		MainDeck:    mainDeck,
		Sideboard:   sideboard,
		Colors:      []string{"W", "U", "B"},
		CreatedAt:   1700000000000,
		LastPlayed:  1700100000000,
		WinRate:     0.58,
		MatchCount:  50,
		Description: "A control deck featuring card draw and removal spells.",
	}
}

func makeJSONTestDraftSession(pickCount int) jsonTestDraftSession {
	picks := make([]jsonTestDraftPick, pickCount)
	for i := range picks {
		packCards := make([]jsonTestCard, 15-i%15)
		for j := range packCards {
			packCards[j] = makeJSONTestCard(i*100 + j)
		}
		picks[i] = jsonTestDraftPick{
			PackNumber: (i / 15) + 1,
			PickNumber: (i % 15) + 1,
			PackCards:  packCards,
			PickedCard: packCards[0],
			Timestamp:  1700000000000 + int64(i*60000),
		}
	}

	finalPool := make([]jsonTestCard, pickCount)
	for i := range finalPool {
		finalPool[i] = makeJSONTestCard(i)
	}

	deckBuilt := make([]jsonTestCard, 40)
	for i := range deckBuilt {
		deckBuilt[i] = makeJSONTestCard(i)
	}

	return jsonTestDraftSession{
		SessionID: "draft-session-12345",
		EventType: "PremierDraft",
		SetCode:   "TST",
		Picks:     picks,
		FinalPool: finalPool,
		DeckBuilt: deckBuilt,
		MatchResults: []jsonTestMatchResult{
			{MatchID: "m1", Result: "win", GameWins: 2, GameLosses: 0},
			{MatchID: "m2", Result: "win", GameWins: 2, GameLosses: 1},
			{MatchID: "m3", Result: "loss", GameWins: 1, GameLosses: 2},
		},
		StartedAt:   1700000000000,
		CompletedAt: 1700003600000,
	}
}

// BenchmarkJSONMarshalV1 benchmarks encoding/json (v1) Marshal.
func BenchmarkJSONMarshalV1(b *testing.B) {
	card := makeJSONTestCard(1)
	deck := makeJSONTestDeck(60)
	draftSession := makeJSONTestDraftSession(45)

	b.Run("Card", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			data, _ := json.Marshal(card)
			runtime.KeepAlive(data)
		}
	})

	b.Run("Deck60", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			data, _ := json.Marshal(deck)
			runtime.KeepAlive(data)
		}
	})

	b.Run("DraftSession", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			data, _ := json.Marshal(draftSession)
			runtime.KeepAlive(data)
		}
	})

	cards := make([]jsonTestCard, 100)
	for i := range cards {
		cards[i] = makeJSONTestCard(i)
	}

	b.Run("Cards100", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			data, _ := json.Marshal(cards)
			runtime.KeepAlive(data)
		}
	})
}

// BenchmarkJSONMarshalV2 benchmarks encoding/json/v2 Marshal.
func BenchmarkJSONMarshalV2(b *testing.B) {
	card := makeJSONTestCard(1)
	deck := makeJSONTestDeck(60)
	draftSession := makeJSONTestDraftSession(45)

	b.Run("Card", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			data, _ := jsonv2.Marshal(card)
			runtime.KeepAlive(data)
		}
	})

	b.Run("Deck60", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			data, _ := jsonv2.Marshal(deck)
			runtime.KeepAlive(data)
		}
	})

	b.Run("DraftSession", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			data, _ := jsonv2.Marshal(draftSession)
			runtime.KeepAlive(data)
		}
	})

	cards := make([]jsonTestCard, 100)
	for i := range cards {
		cards[i] = makeJSONTestCard(i)
	}

	b.Run("Cards100", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			data, _ := jsonv2.Marshal(cards)
			runtime.KeepAlive(data)
		}
	})
}

// BenchmarkJSONUnmarshalV1 benchmarks encoding/json (v1) Unmarshal.
func BenchmarkJSONUnmarshalV1(b *testing.B) {
	cardJSON, _ := json.Marshal(makeJSONTestCard(1))
	deckJSON, _ := json.Marshal(makeJSONTestDeck(60))
	draftJSON, _ := json.Marshal(makeJSONTestDraftSession(45))

	cards := make([]jsonTestCard, 100)
	for i := range cards {
		cards[i] = makeJSONTestCard(i)
	}
	cardsJSON, _ := json.Marshal(cards)

	b.Run("Card", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var card jsonTestCard
			_ = json.Unmarshal(cardJSON, &card)
		}
	})

	b.Run("Deck60", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var deck jsonTestDeck
			_ = json.Unmarshal(deckJSON, &deck)
		}
	})

	b.Run("DraftSession", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var draft jsonTestDraftSession
			_ = json.Unmarshal(draftJSON, &draft)
		}
	})

	b.Run("Cards100", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var cards []jsonTestCard
			_ = json.Unmarshal(cardsJSON, &cards)
		}
	})
}

// BenchmarkJSONUnmarshalV2 benchmarks encoding/json/v2 Unmarshal.
func BenchmarkJSONUnmarshalV2(b *testing.B) {
	cardJSON, _ := json.Marshal(makeJSONTestCard(1))
	deckJSON, _ := json.Marshal(makeJSONTestDeck(60))
	draftJSON, _ := json.Marshal(makeJSONTestDraftSession(45))

	cards := make([]jsonTestCard, 100)
	for i := range cards {
		cards[i] = makeJSONTestCard(i)
	}
	cardsJSON, _ := json.Marshal(cards)

	b.Run("Card", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var card jsonTestCard
			_ = jsonv2.Unmarshal(cardJSON, &card)
		}
	})

	b.Run("Deck60", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var deck jsonTestDeck
			_ = jsonv2.Unmarshal(deckJSON, &deck)
		}
	})

	b.Run("DraftSession", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var draft jsonTestDraftSession
			_ = jsonv2.Unmarshal(draftJSON, &draft)
		}
	})

	b.Run("Cards100", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var cards []jsonTestCard
			_ = jsonv2.Unmarshal(cardsJSON, &cards)
		}
	})
}

// BenchmarkJSONStreamV1 benchmarks streaming JSON encoding/decoding with v1.
func BenchmarkJSONStreamV1(b *testing.B) {
	cards := make([]jsonTestCard, 50)
	for i := range cards {
		cards[i] = makeJSONTestCard(i)
	}

	b.Run("Encode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			for _, card := range cards {
				_ = enc.Encode(card)
			}
			runtime.KeepAlive(buf.Bytes())
		}
	})

	// Prepare data for decode benchmark
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, card := range cards {
		_ = enc.Encode(card)
	}
	data := buf.Bytes()

	b.Run("Decode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			reader := bytes.NewReader(data)
			dec := json.NewDecoder(reader)
			for j := 0; j < 50; j++ {
				var card jsonTestCard
				if err := dec.Decode(&card); err != nil {
					break
				}
			}
		}
	})
}

// Note: BenchmarkJSONStreamV2 is not included because json/v2 uses a different
// streaming API (jsontext.Encoder/Decoder) which is not directly comparable.
// The Marshal/Unmarshal benchmarks above provide the main comparison points.
