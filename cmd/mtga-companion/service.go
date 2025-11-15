package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kardianos/service"
)

// daemonProgram implements service.Interface
type daemonProgram struct {
	exit chan struct{}
}

// Start implements service.Interface
func (p *daemonProgram) Start(s service.Service) error {
	log.Println("Starting MTGA Companion daemon service...")
	p.exit = make(chan struct{})
	go p.run()
	return nil
}

// run executes the daemon service
func (p *daemonProgram) run() {
	// Run the daemon command in a goroutine
	// This keeps the service running
	runDaemonCommand()
}

// Stop implements service.Interface
func (p *daemonProgram) Stop(s service.Service) error {
	log.Println("Stopping MTGA Companion daemon service...")
	close(p.exit)
	return nil
}

// getServiceConfig returns the service configuration
func getServiceConfig() *service.Config {
	return &service.Config{
		Name:        "MTGACompanionDaemon",
		DisplayName: "MTGA Companion Daemon",
		Description: "Background service that monitors MTGA log files and provides data to the MTGA Companion GUI",
	}
}

// runServiceCommand handles service management commands
func runServiceCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: mtga-companion service [install|uninstall|start|stop|restart|status]")
		fmt.Println("\nAvailable commands:")
		fmt.Println("  install    - Install the daemon as a system service")
		fmt.Println("  uninstall  - Uninstall the daemon service")
		fmt.Println("  start      - Start the daemon service")
		fmt.Println("  stop       - Stop the daemon service")
		fmt.Println("  restart    - Restart the daemon service")
		fmt.Println("  status     - Show daemon service status")
		os.Exit(1)
	}

	action := os.Args[2]

	prg := &daemonProgram{}
	svcConfig := getServiceConfig()
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}

	switch action {
	case "install":
		err = s.Install()
		if err != nil {
			log.Fatalf("Failed to install service: %v", err)
		}
		fmt.Println("✓ Service installed successfully")
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Start the service: mtga-companion service start")
		fmt.Println("  2. Verify it's running: mtga-companion service status")
		fmt.Println("  3. View logs:")
		if service.Platform() == "darwin" {
			fmt.Println("     tail -f ~/Library/Logs/MTGACompanionDaemon.log")
		} else if service.Platform() == "windows-service" {
			fmt.Println("     Check Event Viewer or C:\\ProgramData\\MTGACompanionDaemon\\logs")
		} else {
			fmt.Println("     journalctl -u MTGACompanionDaemon -f")
		}

	case "uninstall":
		err = s.Uninstall()
		if err != nil {
			log.Fatalf("Failed to uninstall service: %v", err)
		}
		fmt.Println("✓ Service uninstalled successfully")

	case "start":
		err = s.Start()
		if err != nil {
			log.Fatalf("Failed to start service: %v", err)
		}
		fmt.Println("✓ Service started successfully")
		fmt.Println("\nThe daemon is now running in the background.")
		fmt.Println("You can now launch the GUI: ./MTGA-Companion.app")

	case "stop":
		err = s.Stop()
		if err != nil {
			log.Fatalf("Failed to stop service: %v", err)
		}
		fmt.Println("✓ Service stopped successfully")

	case "restart":
		err = s.Restart()
		if err != nil {
			log.Fatalf("Failed to restart service: %v", err)
		}
		fmt.Println("✓ Service restarted successfully")

	case "status":
		status, err := s.Status()
		if err != nil {
			log.Fatalf("Failed to get service status: %v", err)
		}

		fmt.Println("Service Status:")
		switch status {
		case service.StatusRunning:
			fmt.Println("  Status: ✓ Running")
		case service.StatusStopped:
			fmt.Println("  Status: ● Stopped")
		case service.StatusUnknown:
			fmt.Println("  Status: ? Unknown")
		default:
			fmt.Printf("  Status: %v\n", status)
		}

		fmt.Println("\nService Details:")
		fmt.Printf("  Name: %s\n", svcConfig.Name)
		fmt.Printf("  Display Name: %s\n", svcConfig.DisplayName)
		fmt.Printf("  Description: %s\n", svcConfig.Description)

	default:
		fmt.Printf("Unknown service command: %s\n", action)
		fmt.Println("Usage: mtga-companion service [install|uninstall|start|stop|restart|status]")
		os.Exit(1)
	}
}
