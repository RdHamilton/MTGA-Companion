package gui

import (
	"context"

	"github.com/ramonehamilton/MTGA-Companion/internal/ipc"
	"github.com/ramonehamilton/MTGA-Companion/internal/metrics"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/datasets"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/setcache"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// Services contains all shared services needed by facades.
// This struct is passed to each facade to provide access to common dependencies.
type Services struct {
	// Context for the application
	Context context.Context

	// Storage service for database operations
	Storage *storage.Service

	// Card data services
	SetFetcher     *setcache.Fetcher
	RatingsFetcher *setcache.RatingsFetcher
	DatasetService *datasets.Service

	// Log monitoring
	Poller *logreader.Poller

	// IPC/Daemon communication
	IPCClient *ipc.Client

	// Performance metrics
	DraftMetrics *metrics.DraftMetrics

	// Daemon mode flag
	DaemonMode bool
	DaemonPort int
}

// AppError represents an application error with a user-friendly message.
type AppError struct {
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}
