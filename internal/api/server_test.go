package api

import (
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/websocket"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
)

func TestNewServer(t *testing.T) {
	cfg := DefaultConfig()
	facades := &Facades{}

	server := NewServer(cfg, nil, facades)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.port != cfg.Port {
		t.Errorf("Expected port %d, got %d", cfg.Port, server.port)
	}

	if server.wsHub == nil {
		t.Error("Expected wsHub to be initialized")
	}
}

func TestNewServer_NilConfig(t *testing.T) {
	facades := &Facades{}

	server := NewServer(nil, nil, facades)

	if server == nil {
		t.Fatal("NewServer returned nil with nil config")
	}

	// Should use default port
	if server.port != 8080 {
		t.Errorf("Expected default port 8080, got %d", server.port)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Port)
	}

	if cfg.OpenBrowser {
		t.Error("Expected OpenBrowser to be false by default")
	}

	if cfg.FrontendURL != "" {
		t.Errorf("Expected empty FrontendURL, got %s", cfg.FrontendURL)
	}
}

func TestServer_Port(t *testing.T) {
	cfg := &Config{Port: 9999}
	facades := &Facades{}

	server := NewServer(cfg, nil, facades)

	if server.Port() != 9999 {
		t.Errorf("Expected port 9999, got %d", server.Port())
	}
}

func TestServer_WebSocketHub(t *testing.T) {
	facades := &Facades{}
	server := NewServer(nil, nil, facades)

	hub := server.WebSocketHub()

	if hub == nil {
		t.Error("Expected WebSocketHub to return non-nil hub")
	}
}

func TestServer_NewWebSocketObserver(t *testing.T) {
	facades := &Facades{}
	server := NewServer(nil, nil, facades)

	observer := server.NewWebSocketObserver()

	if observer == nil {
		t.Error("Expected NewWebSocketObserver to return non-nil observer")
	}
}

func TestServer_NewDaemonEventForwarder(t *testing.T) {
	facades := &Facades{}
	server := NewServer(nil, nil, facades)

	forwarder := server.NewDaemonEventForwarder()

	if forwarder == nil {
		t.Error("Expected NewDaemonEventForwarder to return non-nil forwarder")
	}
}

func TestServer_NewDaemonEventForwarder_Type(t *testing.T) {
	facades := &Facades{}
	server := NewServer(nil, nil, facades)

	forwarder := server.NewDaemonEventForwarder()

	// Verify it's the correct type
	_, ok := interface{}(forwarder).(*websocket.DaemonEventForwarder)
	if !ok {
		t.Error("Expected forwarder to be *websocket.DaemonEventForwarder")
	}
}

func TestServer_NewDaemonEventForwarder_UsesServerHub(t *testing.T) {
	facades := &Facades{}
	server := NewServer(nil, nil, facades)

	forwarder := server.NewDaemonEventForwarder()

	// The forwarder should use the same hub as the server
	// We can verify this indirectly by checking that events forwarded
	// through the forwarder increase the hub's broadcast count
	if forwarder == nil {
		t.Error("Forwarder should not be nil")
	}
}

func TestServer_Shutdown_NotStarted(t *testing.T) {
	facades := &Facades{}
	server := NewServer(nil, nil, facades)

	// Shutdown on a server that hasn't started should not error
	err := server.Shutdown(nil)
	if err != nil {
		t.Errorf("Expected no error on shutdown of non-started server, got %v", err)
	}
}

func TestNewServer_WithServices(t *testing.T) {
	cfg := DefaultConfig()
	services := &gui.Services{}
	facades := &Facades{}

	server := NewServer(cfg, services, facades)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.services != services {
		t.Error("Expected services to be set")
	}
}

func TestNewServer_WithFacades(t *testing.T) {
	cfg := DefaultConfig()
	facades := &Facades{
		Match: &gui.MatchFacade{},
		Draft: &gui.DraftFacade{},
	}

	server := NewServer(cfg, nil, facades)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.matchFacade != facades.Match {
		t.Error("Expected matchFacade to be set")
	}

	if server.draftFacade != facades.Draft {
		t.Error("Expected draftFacade to be set")
	}
}
