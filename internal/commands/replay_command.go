package commands

import (
	"context"
	"fmt"
	"log"
)

// IPCClient defines the interface for sending commands to the daemon.
type IPCClient interface {
	Send(message map[string]interface{}) error
	IsConnected() bool
}

// ReplayCommand encapsulates the operation of replaying MTGA logs through the daemon.
// Implements the Command pattern for daemon replay operations.
type ReplayCommand struct {
	BaseCommand
	ipcClient     IPCClient
	filePaths     []string
	speed         float64
	filterType    string
	pauseOnDraft  bool
	clearDataMode bool
}

// NewReplayCommand creates a new command for replaying MTGA logs.
// filePaths: paths to log files to replay (empty for current log replay)
// speed: playback speed multiplier (1.0 = realtime, 10.0 = 10x speed)
// filterType: filter type ("all", "drafts", "matches", etc.)
// pauseOnDraft: whether to pause replay when draft events are encountered
// clearDataMode: if true, clears existing data before replay
func NewReplayCommand(
	ipcClient IPCClient,
	filePaths []string,
	speed float64,
	filterType string,
	pauseOnDraft bool,
	clearDataMode bool,
) *ReplayCommand {
	var desc string
	if clearDataMode {
		desc = fmt.Sprintf("Replay current MTGA logs with data clearing (speed: %.1fx, filter: %s)", speed, filterType)
	} else {
		desc = fmt.Sprintf("Replay %d log file(s) (speed: %.1fx, filter: %s, pause on draft: %v)",
			len(filePaths), speed, filterType, pauseOnDraft)
	}

	return &ReplayCommand{
		BaseCommand: BaseCommand{
			name:        "ReplayLogs",
			description: desc,
		},
		ipcClient:     ipcClient,
		filePaths:     filePaths,
		speed:         speed,
		filterType:    filterType,
		pauseOnDraft:  pauseOnDraft,
		clearDataMode: clearDataMode,
	}
}

// Execute starts the log replay operation.
func (c *ReplayCommand) Execute(ctx context.Context) error {
	log.Printf("[ReplayCommand] Executing: %s", c.description)

	// Validate IPC client connection
	if c.ipcClient == nil || !c.ipcClient.IsConnected() {
		return fmt.Errorf("not connected to daemon - replay requires daemon mode")
	}

	var message map[string]interface{}

	if c.clearDataMode {
		// Replay current logs with data clearing
		message = map[string]interface{}{
			"type":       "replay_logs",
			"clear_data": true,
		}
		log.Printf("[ReplayCommand] Sending replay_logs command with clear_data=true")
	} else {
		// Replay specific log files
		message = map[string]interface{}{
			"type":           "start_replay",
			"file_paths":     c.filePaths,
			"speed":          c.speed,
			"filter":         c.filterType,
			"pause_on_draft": c.pauseOnDraft,
		}
		log.Printf("[ReplayCommand] Sending start_replay command for %d file(s)", len(c.filePaths))
	}

	// Send command to daemon
	if err := c.ipcClient.Send(message); err != nil {
		return fmt.Errorf("failed to send replay command to daemon: %w", err)
	}

	log.Printf("[ReplayCommand] Successfully sent replay command to daemon")
	return nil
}

// PauseReplayCommand pauses an active replay operation.
type PauseReplayCommand struct {
	BaseCommand
	ipcClient IPCClient
}

// NewPauseReplayCommand creates a new command to pause replay.
func NewPauseReplayCommand(ipcClient IPCClient) *PauseReplayCommand {
	return &PauseReplayCommand{
		BaseCommand: BaseCommand{
			name:        "PauseReplay",
			description: "Pause the current log replay operation",
		},
		ipcClient: ipcClient,
	}
}

// Execute pauses the replay.
func (c *PauseReplayCommand) Execute(ctx context.Context) error {
	log.Printf("[PauseReplayCommand] Executing")

	if c.ipcClient == nil || !c.ipcClient.IsConnected() {
		return fmt.Errorf("not connected to daemon")
	}

	message := map[string]interface{}{
		"type": "pause_replay",
	}

	if err := c.ipcClient.Send(message); err != nil {
		return fmt.Errorf("failed to send pause command: %w", err)
	}

	log.Printf("[PauseReplayCommand] Successfully paused replay")
	return nil
}

// ResumeReplayCommand resumes a paused replay operation.
type ResumeReplayCommand struct {
	BaseCommand
	ipcClient IPCClient
}

// NewResumeReplayCommand creates a new command to resume replay.
func NewResumeReplayCommand(ipcClient IPCClient) *ResumeReplayCommand {
	return &ResumeReplayCommand{
		BaseCommand: BaseCommand{
			name:        "ResumeReplay",
			description: "Resume the paused log replay operation",
		},
		ipcClient: ipcClient,
	}
}

// Execute resumes the replay.
func (c *ResumeReplayCommand) Execute(ctx context.Context) error {
	log.Printf("[ResumeReplayCommand] Executing")

	if c.ipcClient == nil || !c.ipcClient.IsConnected() {
		return fmt.Errorf("not connected to daemon")
	}

	message := map[string]interface{}{
		"type": "resume_replay",
	}

	if err := c.ipcClient.Send(message); err != nil {
		return fmt.Errorf("failed to send resume command: %w", err)
	}

	log.Printf("[ResumeReplayCommand] Successfully resumed replay")
	return nil
}

// StopReplayCommand stops an active replay operation.
type StopReplayCommand struct {
	BaseCommand
	ipcClient IPCClient
}

// NewStopReplayCommand creates a new command to stop replay.
func NewStopReplayCommand(ipcClient IPCClient) *StopReplayCommand {
	return &StopReplayCommand{
		BaseCommand: BaseCommand{
			name:        "StopReplay",
			description: "Stop the current log replay operation",
		},
		ipcClient: ipcClient,
	}
}

// Execute stops the replay.
func (c *StopReplayCommand) Execute(ctx context.Context) error {
	log.Printf("[StopReplayCommand] Executing")

	if c.ipcClient == nil || !c.ipcClient.IsConnected() {
		return fmt.Errorf("not connected to daemon")
	}

	message := map[string]interface{}{
		"type": "stop_replay",
	}

	if err := c.ipcClient.Send(message); err != nil {
		return fmt.Errorf("failed to send stop command: %w", err)
	}

	log.Printf("[StopReplayCommand] Successfully stopped replay")
	return nil
}
