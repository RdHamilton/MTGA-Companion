package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/daemon"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

var (
	// NEW STANDARDIZED FLAGS (v0.2.0+)
	// Application mode flags
	debugMode      = flag.Bool("debug-mode", false, "Enable verbose debug logging")
	debugModeShort = flag.Bool("d", false, "Enable debug logging (shorthand for -debug-mode)")

	// Log file configuration flags
	logFilePath     = flag.String("log-file-path", "", "Path to MTGA Player.log file (auto-detected if not specified)")
	logPollInterval = flag.Duration("log-poll-interval", 2*time.Second, "Interval for polling log file (e.g., 1s, 2s, 5s)")
	logUseFsnotify  = flag.Bool("log-use-fsnotify", true, "Use file system events (fsnotify) for monitoring")

	// Cache configuration flags
	cacheEnabled = flag.Bool("cache-enabled", true, "Enable in-memory caching for card ratings (default: true)")

	// DEPRECATED FLAGS (v0.1.0) - Will be removed in v2.0.0
	// These are kept for backward compatibility
	pollInterval  = flag.Duration("poll-interval", 2*time.Second, "DEPRECATED: Use -log-poll-interval instead")
	useFileEvents = flag.Bool("use-file-events", true, "DEPRECATED: Use -log-use-fsnotify instead")
	logPath       = flag.String("log-path", "", "DEPRECATED: Use -log-file-path instead")
	debug         = flag.Bool("debug", false, "DEPRECATED: Use -debug-mode or -d instead")
	cacheOld      = flag.Bool("cache", true, "DEPRECATED: Use -cache-enabled instead")
)

// deprecatedFlags tracks which deprecated flags were explicitly set by the user
var deprecatedFlags = make(map[string]string)

// flagDeprecationMap maps old flag names to their new equivalents
var flagDeprecationMap = map[string]string{
	"debug":           "debug-mode (or -d)",
	"cache":           "cache-enabled",
	"poll-interval":   "log-poll-interval",
	"use-file-events": "log-use-fsnotify",
	"log-path":        "log-file-path",
}

// checkDeprecatedFlags detects and warns about deprecated flag usage
func checkDeprecatedFlags() {
	visited := make(map[string]bool)

	flag.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
		if newFlag, ok := flagDeprecationMap[f.Name]; ok {
			deprecatedFlags[f.Name] = newFlag
		}
	})

	// Print deprecation warnings
	if len(deprecatedFlags) > 0 {
		fmt.Fprintf(os.Stderr, "\n⚠️  Warning: You are using deprecated flags:\n")
		for oldFlag, newFlag := range deprecatedFlags {
			fmt.Fprintf(os.Stderr, "   - Flag '-%s' is deprecated. Use '-%s' instead.\n", oldFlag, newFlag)
		}
		fmt.Fprintf(os.Stderr, "   Deprecated flags will be removed in v2.0.0.\n")
		fmt.Fprintf(os.Stderr, "   See FLAG_MIGRATION.md for migration guide.\n\n")
	}

	// Map old flag values to new flags for backward compatibility
	if visited["debug"] && !visited["debug-mode"] {
		*debugMode = *debug
	}
	if visited["cache"] && !visited["cache-enabled"] {
		*cacheEnabled = *cacheOld
	}
	if visited["poll-interval"] && !visited["log-poll-interval"] {
		*logPollInterval = *pollInterval
	}
	if visited["use-file-events"] && !visited["log-use-fsnotify"] {
		*logUseFsnotify = *useFileEvents
	}
	if visited["log-path"] && !visited["log-file-path"] {
		*logFilePath = *logPath
	}

	// Handle shorthand flags
	if *debugModeShort {
		*debugMode = true
	}
}

// getDBPath returns the database path from environment variable or default location.
func getDBPath() string {
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}
		dbPath = filepath.Join(home, ".mtga-companion", "mtga.db")
	}
	return dbPath
}

func main() {
	// Parse flags before checking for subcommands
	flag.Parse()

	// Check for deprecated flag usage and apply backward compatibility
	checkDeprecatedFlags()

	// Validate poll interval
	if *logPollInterval < 1*time.Second {
		log.Fatalf("Poll interval must be at least 1 second, got %v", *logPollInterval)
	}
	if *logPollInterval > 1*time.Minute {
		log.Fatalf("Poll interval must be at most 1 minute, got %v", *logPollInterval)
	}

	// Check if this is a migration command
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrationCommand()
		return
	}

	// Check if this is a backup command
	if len(os.Args) > 1 && os.Args[1] == "backup" {
		runBackupCommand()
		return
	}

	// Check if this is a service command
	if len(os.Args) > 1 && os.Args[1] == "service" {
		runServiceCommand()
		return
	}

	// Check if this is a daemon command
	if len(os.Args) > 1 && os.Args[1] == "daemon" {
		runDaemonCommand()
		return
	}

	// Check if this is a replay command
	if len(os.Args) > 1 && os.Args[1] == "replay" {
		runReplayCommand()
		return
	}

	// No command provided - show usage
	printUsage()
}

func printUsage() {
	fmt.Println("MTGA Companion - Daemon")
	fmt.Println("=======================")
	fmt.Println()
	fmt.Println("Usage: mtga-companion <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  daemon     - Run the background daemon service")
	fmt.Println("  service    - Manage daemon as system service (install/start/stop/status/uninstall)")
	fmt.Println("  migrate    - Run database migrations")
	fmt.Println("  backup     - Create database backup")
	fmt.Println("  replay     - Replay historical log files for testing")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  mtga-companion daemon --port 9999")
	fmt.Println("  mtga-companion service install")
	fmt.Println("  mtga-companion service start")
	fmt.Println("  mtga-companion migrate up")
	fmt.Println("  mtga-companion backup create")
	fmt.Println()
	fmt.Println("For more information, see: https://github.com/RdHamilton/MTGA-Companion")
	fmt.Println()
}

func runMigrationCommand() {
	if len(os.Args) < 3 {
		printMigrationUsage()
		os.Exit(1)
	}

	// Get database path from environment or use default
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}
		dbPath = filepath.Join(home, ".mtga-companion", "data.db")
	}

	// Ensure directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		log.Fatalf("Error creating database directory: %v", err)
	}

	// Create migration manager
	mgr, err := storage.NewMigrationManager(dbPath)
	if err != nil {
		log.Fatalf("Error creating migration manager: %v", err)
	}
	defer func() {
		if err := mgr.Close(); err != nil {
			log.Printf("Error closing migration manager: %v", err)
		}
	}()

	command := os.Args[2]

	switch command {
	case "up":
		fmt.Println("Applying all pending migrations...")
		if err := mgr.Up(); err != nil {
			log.Fatalf("Error applying migrations: %v", err)
		}
		version, dirty, err := mgr.Version()
		if err != nil {
			log.Fatalf("Error getting version: %v", err)
		}
		if dirty {
			fmt.Printf("Current version: %d (dirty)\n", version)
		} else {
			fmt.Printf("Current version: %d\n", version)
		}
		fmt.Println("All migrations applied successfully!")

	case "down":
		fmt.Println("Rolling back last migration...")
		if err := mgr.Down(); err != nil {
			log.Fatalf("Error rolling back migration: %v", err)
		}
		version, dirty, err := mgr.Version()
		if err != nil {
			log.Fatalf("Error getting version: %v", err)
		}
		if dirty {
			fmt.Printf("Current version: %d (dirty)\n", version)
		} else {
			fmt.Printf("Current version: %d\n", version)
		}
		fmt.Println("Migration rolled back successfully!")

	case "status", "version":
		version, dirty, err := mgr.Version()
		if err != nil {
			log.Fatalf("Error getting version: %v", err)
		}
		if dirty {
			fmt.Printf("Current version: %d (dirty - migration failed or interrupted)\n", version)
			fmt.Println("Use 'migrate force <version>' to recover")
		} else {
			fmt.Printf("Current version: %d\n", version)
		}

	case "force":
		if len(os.Args) < 4 {
			fmt.Println("Error: force command requires a version number")
			fmt.Println("Usage: mtga-companion migrate force <version>")
			os.Exit(1)
		}
		versionStr := os.Args[3]
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		fmt.Printf("Forcing migration version to %d...\n", version)
		fmt.Println("WARNING: This does not run migrations, only sets the version.")
		if err := mgr.Force(version); err != nil {
			log.Fatalf("Error forcing version: %v", err)
		}
		fmt.Println("Version forced successfully!")

	case "goto":
		if len(os.Args) < 4 {
			fmt.Println("Error: goto command requires a version number")
			fmt.Println("Usage: mtga-companion migrate goto <version>")
			os.Exit(1)
		}
		versionStr := os.Args[3]
		version, err := strconv.ParseUint(versionStr, 10, 32)
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		fmt.Printf("Migrating to version %d...\n", version)
		if err := mgr.Goto(uint(version)); err != nil {
			log.Fatalf("Error migrating to version %d: %v", version, err)
		}
		fmt.Println("Migration successful!")

	default:
		fmt.Printf("Unknown migration command: %s\n\n", command)
		printMigrationUsage()
		os.Exit(1)
	}
}

func printMigrationUsage() {
	fmt.Println("MTGA Companion - Database Migration Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  mtga-companion migrate <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up                Apply all pending migrations")
	fmt.Println("  down              Rollback the last migration")
	fmt.Println("  status            Show current migration version")
	fmt.Println("  version           Show current migration version (alias for status)")
	fmt.Println("  goto <version>    Migrate to a specific version")
	fmt.Println("  force <version>   Force set migration version (use with caution)")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  MTGA_DB_PATH      Override default database path")
	fmt.Println("                    (default: ~/.mtga-companion/data.db)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  mtga-companion migrate up")
	fmt.Println("  mtga-companion migrate status")
	fmt.Println("  mtga-companion migrate goto 1")
	fmt.Println("  MTGA_DB_PATH=/tmp/test.db mtga-companion migrate up")
}

// runBackupCommand handles backup and restore commands.
func runBackupCommand() {
	// Get database path from environment or use default
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}
		dbPath = filepath.Join(home, ".mtga-companion", "data.db")
	}

	// Check if database exists (except for list command which doesn't need it)
	if len(os.Args) >= 3 && os.Args[2] != "list" && os.Args[2] != "ls" {
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			log.Fatalf("Database file does not exist: %s", dbPath)
		}
	}

	// Create backup manager
	backupMgr := storage.NewBackupManager(dbPath)

	if len(os.Args) < 3 {
		printBackupUsage()
		os.Exit(1)
	}

	command := os.Args[2]

	switch command {
	case "create", "backup":
		// Define flags for create command
		createFlags := flag.NewFlagSet("create", flag.ExitOnError)
		backupType := createFlags.String("type", "full", "Backup type: 'full' or 'incremental'")
		backupDir := createFlags.String("dir", os.Getenv("MTGA_BACKUP_DIR"), "Backup directory")
		backupName := createFlags.String("name", "", "Backup name (default: auto-generated timestamp)")
		compress := createFlags.Bool("compress", false, "Compress backup with gzip")
		encrypt := createFlags.Bool("encrypt", false, "Encrypt backup")
		passwordEnv := createFlags.String("password-env", "", "Environment variable containing encryption password")
		verify := createFlags.Bool("verify", true, "Verify backup after creation")

		if err := createFlags.Parse(os.Args[3:]); err != nil {
			log.Fatalf("Error parsing flags: %v", err)
		}

		// Build backup config
		config := storage.DefaultBackupConfig()
		config.BackupDir = *backupDir
		config.BackupName = *backupName
		config.VerifyBackup = *verify
		config.Compress = *compress
		config.Encrypt = *encrypt

		// Set backup type
		switch *backupType {
		case "full":
			config.BackupType = storage.BackupTypeFull
		case "incremental", "incr":
			config.BackupType = storage.BackupTypeIncremental
		default:
			log.Fatalf("Invalid backup type: %s (must be 'full' or 'incremental')", *backupType)
		}

		// Handle encryption password
		if *encrypt {
			if *passwordEnv == "" {
				log.Fatal("Error: --password-env is required when --encrypt is specified")
			}
			password := os.Getenv(*passwordEnv)
			if password == "" {
				log.Fatalf("Error: environment variable %s is not set or empty", *passwordEnv)
			}
			config.EncryptionPassword = password
		}

		// Print configuration
		fmt.Printf("Creating %s backup...\n", *backupType)
		if *compress {
			fmt.Println("  Compression: enabled")
		}
		if *encrypt {
			fmt.Println("  Encryption: enabled")
		}

		backupPath, err := backupMgr.Backup(config)
		if err != nil {
			log.Fatalf("Error creating backup: %v", err)
		}

		fmt.Printf("\n✓ Backup created successfully: %s\n", backupPath)

		// Display backup size
		info, err := os.Stat(backupPath)
		if err == nil {
			sizeMB := float64(info.Size()) / (1024 * 1024)
			fmt.Printf("  Size: %.2f MB\n", sizeMB)
		}

	case "restore":
		// Define flags for restore command
		restoreFlags := flag.NewFlagSet("restore", flag.ExitOnError)
		passwordEnv := restoreFlags.String("password-env", "", "Environment variable containing decryption password")
		noConfirm := restoreFlags.Bool("yes", false, "Skip confirmation prompt")

		if err := restoreFlags.Parse(os.Args[3:]); err != nil {
			log.Fatalf("Error parsing flags: %v", err)
		}

		if restoreFlags.NArg() < 1 {
			fmt.Println("Error: restore command requires a backup file path")
			fmt.Println("Usage: mtga-companion backup restore <backup-file> [flags]")
			fmt.Println("\nFlags:")
			restoreFlags.PrintDefaults()
			os.Exit(1)
		}
		backupPath := restoreFlags.Arg(0)

		// Check if backup file exists
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			log.Fatalf("Backup file does not exist: %s", backupPath)
		}

		// Show warning and get confirmation
		if !*noConfirm {
			fmt.Println("WARNING: This will overwrite the current database!")
			fmt.Printf("Database: %s\n", dbPath)
			fmt.Printf("Backup:   %s\n", backupPath)
			fmt.Print("\nAre you sure you want to continue? (yes/no): ")

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				log.Fatalf("Error reading input: %v", err)
			}

			response = strings.TrimSpace(strings.ToLower(response))
			if response != "yes" && response != "y" {
				fmt.Println("Restore cancelled.")
				return
			}
		}

		fmt.Println("\nRestoring database from backup...")

		// Handle decryption password if needed
		var password string
		if *passwordEnv != "" {
			password = os.Getenv(*passwordEnv)
			if password == "" {
				log.Fatalf("Error: environment variable %s is not set or empty", *passwordEnv)
			}
		}

		// Restore with optional password
		if password != "" {
			if err := backupMgr.Restore(backupPath, password); err != nil {
				log.Fatalf("Error restoring backup: %v", err)
			}
		} else {
			if err := backupMgr.Restore(backupPath); err != nil {
				log.Fatalf("Error restoring backup: %v", err)
			}
		}

		fmt.Println("✓ Database restored successfully!")

	case "list", "ls":
		// Define flags for list command
		listFlags := flag.NewFlagSet("list", flag.ExitOnError)
		format := listFlags.String("format", "table", "Output format: 'table' or 'json'")
		backupDir := listFlags.String("dir", os.Getenv("MTGA_BACKUP_DIR"), "Backup directory")

		if err := listFlags.Parse(os.Args[3:]); err != nil {
			log.Fatalf("Error parsing flags: %v", err)
		}

		if *backupDir == "" {
			*backupDir = backupMgr.GetBackupDir()
		}

		backups, err := backupMgr.ListBackups(*backupDir)
		if err != nil {
			log.Fatalf("Error listing backups: %v", err)
		}

		if len(backups) == 0 {
			fmt.Println("No backups found.")
			return
		}

		// Format output
		switch *format {
		case "json":
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(backups); err != nil {
				log.Fatalf("Error encoding JSON: %v", err)
			}
		case "table":
			fmt.Printf("\nFound %d backup(s) in %s:\n\n", len(backups), *backupDir)
			for i, backup := range backups {
				sizeMB := float64(backup.Size) / (1024 * 1024)
				fmt.Printf("%d. %s\n", i+1, backup.Name)
				fmt.Printf("   Path:     %s\n", backup.Path)
				fmt.Printf("   Size:     %.2f MB\n", sizeMB)
				fmt.Printf("   Modified: %s\n", backup.ModTime.Format("2006-01-02 15:04:05"))
				fmt.Printf("   Checksum: %s\n", backup.Checksum)
				fmt.Println()
			}
		default:
			log.Fatalf("Invalid format: %s (must be 'table' or 'json')", *format)
		}

	case "verify":
		// Define flags for verify command
		verifyFlags := flag.NewFlagSet("verify", flag.ExitOnError)
		if err := verifyFlags.Parse(os.Args[3:]); err != nil {
			log.Fatalf("Error parsing flags: %v", err)
		}

		if verifyFlags.NArg() < 1 {
			fmt.Println("Error: verify command requires a backup file path")
			fmt.Println("Usage: mtga-companion backup verify <backup-file>")
			os.Exit(1)
		}
		backupPath := verifyFlags.Arg(0)

		fmt.Printf("Verifying backup: %s\n", backupPath)
		if err := backupMgr.VerifyBackup(backupPath); err != nil {
			log.Fatalf("Backup verification failed: %v", err)
		}

		fmt.Println("✓ Backup verification successful!")

	case "cleanup":
		// Define flags for cleanup command
		cleanupFlags := flag.NewFlagSet("cleanup", flag.ExitOnError)
		backupDir := cleanupFlags.String("dir", os.Getenv("MTGA_BACKUP_DIR"), "Backup directory")
		olderThan := cleanupFlags.Int("older-than", 0, "Delete backups older than N days (0 = disabled)")
		keepLast := cleanupFlags.Int("keep-last", 0, "Keep only the last N backups (0 = disabled)")
		dryRun := cleanupFlags.Bool("dry-run", false, "Show what would be deleted without actually deleting")

		if err := cleanupFlags.Parse(os.Args[3:]); err != nil {
			log.Fatalf("Error parsing flags: %v", err)
		}

		if *backupDir == "" {
			*backupDir = backupMgr.GetBackupDir()
		}

		if *olderThan == 0 && *keepLast == 0 {
			fmt.Println("Error: either --older-than or --keep-last must be specified")
			fmt.Println("Usage: mtga-companion backup cleanup [flags]")
			fmt.Println("\nFlags:")
			cleanupFlags.PrintDefaults()
			os.Exit(1)
		}

		// List backups first to show what would be deleted
		backups, err := backupMgr.ListBackups(*backupDir)
		if err != nil {
			log.Fatalf("Error listing backups: %v", err)
		}

		if len(backups) == 0 {
			fmt.Println("No backups found.")
			return
		}

		if *dryRun {
			fmt.Printf("DRY RUN: Would clean up backups in %s\n", *backupDir)
			fmt.Printf("Found %d backup(s)\n", len(backups))
			if *olderThan > 0 {
				fmt.Printf("  - Deleting backups older than %d days\n", *olderThan)
			}
			if *keepLast > 0 {
				fmt.Printf("  - Keeping only the last %d backups\n", *keepLast)
			}
			return
		}

		fmt.Printf("Cleaning up backups in %s...\n", *backupDir)
		if err := backupMgr.CleanupBackups(*backupDir, *olderThan, *keepLast); err != nil {
			log.Fatalf("Error cleaning up backups: %v", err)
		}

		// Show how many remain
		remainingBackups, err := backupMgr.ListBackups(*backupDir)
		if err == nil {
			fmt.Printf("✓ Cleanup complete. %d backup(s) remaining.\n", len(remainingBackups))
		}

	case "info":
		// Define flags for info command
		infoFlags := flag.NewFlagSet("info", flag.ExitOnError)
		format := infoFlags.String("format", "table", "Output format: 'table' or 'json'")

		if err := infoFlags.Parse(os.Args[3:]); err != nil {
			log.Fatalf("Error parsing flags: %v", err)
		}

		if infoFlags.NArg() < 1 {
			fmt.Println("Error: info command requires a backup file path")
			fmt.Println("Usage: mtga-companion backup info <backup-file> [flags]")
			fmt.Println("\nFlags:")
			infoFlags.PrintDefaults()
			os.Exit(1)
		}
		backupPath := infoFlags.Arg(0)

		// Check if backup exists
		info, err := os.Stat(backupPath)
		if os.IsNotExist(err) {
			log.Fatalf("Backup file does not exist: %s", backupPath)
		}
		if err != nil {
			log.Fatalf("Error accessing backup file: %v", err)
		}

		// Try to load metadata
		metadata, err := backupMgr.LoadBackupMetadata(backupPath)

		// Format output
		switch *format {
		case "json":
			type BackupDetails struct {
				Path     string                  `json:"path"`
				Size     int64                   `json:"size"`
				ModTime  time.Time               `json:"modified"`
				Metadata *storage.BackupMetadata `json:"metadata,omitempty"`
			}

			details := BackupDetails{
				Path:     backupPath,
				Size:     info.Size(),
				ModTime:  info.ModTime(),
				Metadata: metadata,
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(details); err != nil {
				log.Fatalf("Error encoding JSON: %v", err)
			}
		case "table":
			sizeMB := float64(info.Size()) / (1024 * 1024)
			fmt.Printf("\nBackup Information:\n")
			fmt.Printf("  Path:     %s\n", backupPath)
			fmt.Printf("  Size:     %.2f MB\n", sizeMB)
			fmt.Printf("  Modified: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))

			if metadata != nil {
				fmt.Printf("\n  Type:     %s\n", metadata.BackupType)
				fmt.Printf("  Created:  %s\n", metadata.Timestamp.Format("2006-01-02 15:04:05"))
				if metadata.BaseBackup != "" {
					fmt.Printf("  Base:     %s\n", metadata.BaseBackup)
				}
				if len(metadata.Tables) > 0 {
					fmt.Printf("\n  Tables:   %d\n", len(metadata.Tables))
				}
			} else if err != nil {
				fmt.Printf("\n  Metadata: Not available (%v)\n", err)
			}
		default:
			log.Fatalf("Invalid format: %s (must be 'table' or 'json')", *format)
		}

	default:
		fmt.Printf("Unknown backup command: %s\n\n", command)
		printBackupUsage()
		os.Exit(1)
	}
}

func printBackupUsage() {
	fmt.Println("MTGA Companion - Database Backup Management")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  mtga-companion backup <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create     Create a new database backup")
	fmt.Println("  restore    Restore database from backup")
	fmt.Println("  list, ls   List all available backups")
	fmt.Println("  verify     Verify backup integrity")
	fmt.Println("  cleanup    Delete old backups based on retention policy")
	fmt.Println("  info       Show detailed backup information")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Create full backup")
	fmt.Println("  mtga-companion backup create")
	fmt.Println()
	fmt.Println("  # Create incremental backup with encryption")
	fmt.Println("  export BACKUP_PWD=mypassword")
	fmt.Println("  mtga-companion backup create --type incremental --encrypt --password-env BACKUP_PWD")
	fmt.Println()
	fmt.Println("  # Create compressed backup")
	fmt.Println("  mtga-companion backup create --compress")
	fmt.Println()
	fmt.Println("  # Restore from encrypted backup")
	fmt.Println("  mtga-companion backup restore backup.db --password-env BACKUP_PWD")
	fmt.Println()
	fmt.Println("  # List backups in JSON format")
	fmt.Println("  mtga-companion backup list --format json")
	fmt.Println()
	fmt.Println("  # Clean up old backups (keep last 10)")
	fmt.Println("  mtga-companion backup cleanup --keep-last 10")
	fmt.Println()
	fmt.Println("  # Clean up backups older than 30 days")
	fmt.Println("  mtga-companion backup cleanup --older-than 30")
	fmt.Println()
	fmt.Println("  # Show backup metadata")
	fmt.Println("  mtga-companion backup info backup.db")
	fmt.Println()
	fmt.Println("For command-specific help:")
	fmt.Println("  mtga-companion backup <command> --help")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  MTGA_DB_PATH     Path to database file (default: ~/.mtga-companion/data.db)")
	fmt.Println("  MTGA_BACKUP_DIR  Backup directory (default: ~/.mtga-companion/backups)")
	fmt.Println()
}

// runBackupCommandInteractive handles backup commands from the interactive console.
func runDaemonCommand() {
	fs := flag.NewFlagSet("daemon", flag.ExitOnError)
	port := fs.Int("port", 9999, "WebSocket server port")
	logPath := fs.String("log-path", "", "MTGA log file path (auto-detect if empty)")
	dbPath := fs.String("db-path", "", "Database path (default: ~/.mtga-companion/data.db)")
	pollInterval := fs.Duration("poll-interval", 2*time.Second, "Log polling interval")
	useFSNotify := fs.Bool("use-fsnotify", false, "Use file system events for log watching")
	enableMetrics := fs.Bool("enable-metrics", false, "Enable metrics")

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing daemon flags: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("MTGA Companion - Daemon Mode")
	fmt.Println("=============================")
	fmt.Println()

	// Setup database path
	finalDBPath := *dbPath
	if finalDBPath == "" {
		finalDBPath = getDBPath()
	}

	// Open database
	config := storage.DefaultConfig(finalDBPath)
	config.AutoMigrate = true
	db, err := storage.Open(config)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	service := storage.NewService(db)
	defer func() {
		if err := service.Close(); err != nil {
			log.Printf("Error closing service: %v", err)
		}
	}()

	// Create daemon configuration
	daemonConfig := daemon.DefaultConfig()
	daemonConfig.Port = *port
	daemonConfig.DBPath = finalDBPath
	daemonConfig.LogPath = *logPath
	daemonConfig.PollInterval = *pollInterval
	daemonConfig.UseFSNotify = *useFSNotify
	daemonConfig.EnableMetrics = *enableMetrics

	// Create and start daemon
	daemonService := daemon.New(daemonConfig, service)
	if err := daemonService.Start(); err != nil {
		log.Fatalf("Failed to start daemon: %v", err)
	}

	fmt.Println()
	fmt.Println("Daemon is running. Press Ctrl+C to stop.")
	fmt.Println()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	fmt.Println("Shutting down...")

	// Stop daemon
	if err := daemonService.Stop(); err != nil {
		log.Printf("Error stopping daemon: %v", err)
	}

	fmt.Println("Daemon stopped.")
}

// runReplayCommand handles the replay command for testing with historical logs.
func runReplayCommand() {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	file := fs.String("file", "", "Path to log file to replay (required)")
	speed := fs.Float64("speed", 1.0, "Replay speed multiplier (1.0 = real-time, 2.0 = 2x speed, etc.)")
	filter := fs.String("filter", "all", "Filter entries by type: all, draft, match, event")
	dbPath := fs.String("db-path", "", "Database path (default: ~/.mtga-companion/data.db)")
	port := fs.Int("port", 9999, "Daemon port for replay")

	fs.Usage = func() {
		fmt.Println("Usage: mtga-companion replay --file <log-file> [options]")
		fmt.Println()
		fmt.Println("Replay historical MTGA log files with simulated timing for testing.")
		fmt.Println()
		fmt.Println("Options:")
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Replay at normal speed")
		fmt.Println("  mtga-companion replay --file Player.log")
		fmt.Println()
		fmt.Println("  # Replay at 5x speed")
		fmt.Println("  mtga-companion replay --file Player.log --speed 5")
		fmt.Println()
		fmt.Println("  # Replay only draft events")
		fmt.Println("  mtga-companion replay --file Player.log --filter draft")
		fmt.Println()
		fmt.Println("  # Replay only matches")
		fmt.Println("  mtga-companion replay --file Player.log --filter match")
		fmt.Println()
		fmt.Println("Filter types:")
		fmt.Println("  all    - All log entries (no filtering)")
		fmt.Println("  draft  - Draft picks and status")
		fmt.Println("  match  - Match/game events")
		fmt.Println("  event  - Tournament/event entries")
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing replay flags: %v\n", err)
		os.Exit(1)
	}

	// Validate required arguments
	if *file == "" {
		fmt.Fprintf(os.Stderr, "Error: --file is required\n\n")
		fs.Usage()
		os.Exit(1)
	}

	// Validate file exists
	if _, err := os.Stat(*file); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Log file not found: %s\n", *file)
		os.Exit(1)
	}

	// Validate speed
	if *speed <= 0 || *speed > 100 {
		fmt.Fprintf(os.Stderr, "Error: Speed must be between 0 and 100, got %.2f\n", *speed)
		os.Exit(1)
	}

	// Validate filter
	validFilters := map[string]bool{"all": true, "draft": true, "match": true, "event": true}
	if !validFilters[*filter] {
		fmt.Fprintf(os.Stderr, "Error: Invalid filter '%s'. Must be one of: all, draft, match, event\n", *filter)
		os.Exit(1)
	}

	fmt.Println("MTGA Companion - Log Replay")
	fmt.Println("===========================")
	fmt.Println()
	fmt.Printf("File:   %s\n", *file)
	fmt.Printf("Speed:  %.1fx\n", *speed)
	fmt.Printf("Filter: %s\n", *filter)
	fmt.Println()

	// Setup database
	finalDBPath := *dbPath
	if finalDBPath == "" {
		finalDBPath = getDBPath()
	}

	fmt.Printf("Opening database: %s\n", finalDBPath)

	// Open database
	storageConfig := storage.DefaultConfig(finalDBPath)
	storageConfig.AutoMigrate = true
	db, err := storage.Open(storageConfig)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	stor := storage.NewService(db)
	defer func() {
		if err := stor.Close(); err != nil {
			log.Printf("Error closing service: %v", err)
		}
	}()

	// Create daemon config (for replay engine)
	daemonConfig := daemon.DefaultConfig()
	daemonConfig.Port = *port
	daemonConfig.DBPath = finalDBPath

	// Create daemon service
	daemonService := daemon.New(daemonConfig, stor)

	fmt.Println("Starting daemon for replay...")
	if err := daemonService.Start(); err != nil {
		log.Fatalf("Failed to start daemon: %v", err)
	}

	// Give daemon a moment to start
	time.Sleep(500 * time.Millisecond)

	fmt.Println()
	fmt.Println("Starting replay...")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Start replay (CLI doesn't support auto-pause on draft events yet)
	if err := daemonService.StartReplay([]string{*file}, *speed, *filter, false); err != nil {
		log.Fatalf("Failed to start replay: %v", err)
	}

	// Wait for replay to complete or Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Monitor replay status
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status := daemonService.GetReplayStatus()
			if !status["isActive"].(bool) {
				// Replay completed
				fmt.Println()
				fmt.Println("✓ Replay completed")
				goto cleanup
			}

			// Print progress
			percentComplete := status["percentComplete"].(float64)
			currentEntry := status["currentEntry"].(int)
			totalEntries := status["totalEntries"].(int)
			elapsed := status["elapsed"].(float64)

			fmt.Printf("\rProgress: %.1f%% (%d/%d entries) - Elapsed: %.1fs",
				percentComplete, currentEntry, totalEntries, elapsed)

		case <-sigChan:
			fmt.Println()
			fmt.Println()
			fmt.Println("Stopping replay...")
			if err := daemonService.StopReplay(); err != nil {
				log.Printf("Error stopping replay: %v", err)
			}
			goto cleanup
		}
	}

cleanup:
	fmt.Println()
	fmt.Println("Shutting down...")

	// Stop daemon
	if err := daemonService.Stop(); err != nil {
		log.Printf("Error stopping daemon: %v", err)
	}

	fmt.Println("Replay stopped.")
}
