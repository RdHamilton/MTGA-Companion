package gui

import (
	"context"

	"github.com/ramonehamilton/MTGA-Companion/internal/daemon"
	"github.com/ramonehamilton/MTGA-Companion/internal/ipc"
	"github.com/ramonehamilton/MTGA-Companion/internal/meta"
	"github.com/ramonehamilton/MTGA-Companion/internal/metrics"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/datasets"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/setcache"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/deckexport"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/deckimport"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/recommendations"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// CardFetcher defines the interface for fetching card metadata from external sources.
// This interface allows for easy mocking in tests.
type CardFetcher interface {
	FetchCardByArenaID(ctx context.Context, arenaID int) (*models.SetCard, error)
	FetchCardByName(ctx context.Context, setCode, cardName, arenaID string) (*models.SetCard, error)
	FetchAndCacheSet(ctx context.Context, mtgaSetCode string) (int, error)
	RefreshSet(ctx context.Context, setCode string) (int, error)
	GetCardByArenaID(ctx context.Context, arenaID string) (*models.SetCard, error)
}

// Services contains all shared services needed by facades.
// This struct is passed to each facade to provide access to common dependencies.
type Services struct {
	// Context for the application
	Context context.Context

	// Storage service for database operations
	Storage *storage.Service

	// Card data services
	CardService      *cards.Service
	SetFetcher       CardFetcher // Interface for fetching card metadata (allows mocking)
	RatingsFetcher   *setcache.RatingsFetcher
	DatasetService   *datasets.Service
	DeckImportParser *deckimport.Parser

	// Deck operations
	DeckExporter         *deckexport.Exporter
	RecommendationEngine recommendations.RecommendationEngine

	// Log monitoring
	Poller *logreader.Poller

	// IPC/Daemon communication (Go daemon on port 9999)
	IPCClient *ipc.Client

	// Performance metrics
	DraftMetrics *metrics.DraftMetrics

	// Meta service
	MetaService *meta.Service

	// Daemon mode flag
	DaemonMode bool
	DaemonPort int

	// Daemon service (when running integrated)
	DaemonService *daemon.Service
}

// AppError represents an application error with a user-friendly message.
type AppError struct {
	Message string `json:"message"`
	Err     error  `json:"-"` // Wrapped error for errors.Is/As chain
}

func (e *AppError) Error() string {
	return e.Message
}

// Unwrap returns the wrapped error for errors.Is/As chain.
func (e *AppError) Unwrap() error {
	return e.Err
}
