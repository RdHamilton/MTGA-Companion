package storage

// Re-export types from models package for backward compatibility.
import "github.com/ramonehamilton/MTGA-Companion/internal/storage/models"

type (
	Match             = models.Match
	Game              = models.Game
	PlayerStats       = models.PlayerStats
	Deck              = models.Deck
	DeckCard          = models.DeckCard
	CollectionCard    = models.CollectionCard
	CollectionHistory = models.CollectionHistory
	RankHistory       = models.RankHistory
	DraftEvent        = models.DraftEvent
	StatsFilter       = models.StatsFilter
	Statistics        = models.Statistics
)
