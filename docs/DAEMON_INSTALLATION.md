# MTGA Companion Daemon Installation Guide

This guide covers installing and managing the MTGA Companion daemon service on macOS, Windows, and Linux.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [macOS](#macos)
  - [Windows](#windows)
  - [Linux](#linux)
- [Service Management](#service-management)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)
- [Uninstallation](#uninstallation)

## Overview

The MTGA Companion daemon is a background service that:
- Monitors your MTGA Player.log file for updates
- Parses and stores match data, statistics, and deck information
- Provides real-time updates to the GUI via WebSocket
- Runs continuously in the background (even when GUI is closed)
- Starts automatically on system boot

## Platform Verification Status

> **⚠️ IMPORTANT**: Service installation has only been fully tested and verified on **macOS**.

- ✅ **macOS**: Fully tested and verified (v1.0)
- ⚠️ **Windows**: Implementation complete but not yet verified
- ⚠️ **Linux**: Implementation complete but not yet verified

The service installation uses the cross-platform [kardianos/service](https://github.com/kardianos/service) library which has proven support for all three platforms. The implementation should work on Windows and Linux, but has not been tested on those platforms yet. If you encounter issues on Windows or Linux, please [report them](https://github.com/RdHamilton/MTGA-Companion/issues).

## Prerequisites

- MTGA Companion binary (`mtga-companion` or `mtga-companion.exe`)
- Administrator/root privileges for service installation
- MTGA installed and configured

## Installation

### macOS

**1. Open Terminal**

**2. Navigate to the MTGA Companion directory:**
```bash
cd /path/to/MTGA-Companion
```

**3. Install the service:**
```bash
./mtga-companion service install
```

Expected output:
```
✓ Service installed successfully

Next steps:
  1. Start the service: mtga-companion service start
  2. Verify it's running: mtga-companion service status
  3. View logs:
     tail -f ~/Library/Logs/MTGACompanionDaemon.log
```

**4. Start the service:**
```bash
./mtga-companion service start
```

**5. Verify it's running:**
```bash
./mtga-companion service status
```

Expected output:
```
Service Status:
  Status: ✓ Running

Service Details:
  Name: MTGACompanionDaemon
  Display Name: MTGA Companion Daemon
  Description: Background service that monitors MTGA log files...
```

**Service Location:**
- Binary: System will use the current binary location
- Launch Agent: `~/Library/LaunchAgents/MTGACompanionDaemon.plist`
- Logs: `~/Library/Logs/MTGACompanionDaemon.log`

### Windows

**1. Open PowerShell or Command Prompt as Administrator**

Right-click PowerShell → "Run as Administrator"

**2. Navigate to the MTGA Companion directory:**
```powershell
cd C:\Path\To\MTGA-Companion
```

**3. Install the service:**
```powershell
.\mtga-companion.exe service install
```

Expected output:
```
✓ Service installed successfully

Next steps:
  1. Start the service: mtga-companion service start
  2. Verify it's running: mtga-companion service status
  3. View logs:
     Check Event Viewer or C:\ProgramData\MTGACompanionDaemon\logs
```

**4. Start the service:**
```powershell
.\mtga-companion.exe service start
```

**5. Verify it's running:**
```powershell
.\mtga-companion.exe service status
```

**Service Location:**
- Binary: System will use the current binary location
- Service Name: `MTGACompanionDaemon`
- Logs: Event Viewer → Windows Logs → Application (look for "MTGACompanionDaemon")

**Alternative using Windows Services Manager:**
1. Press `Win + R`
2. Type `services.msc`
3. Look for "MTGA Companion Daemon"
4. Right-click → Properties to configure

### Linux

**1. Open Terminal**

**2. Navigate to the MTGA Companion directory:**
```bash
cd /path/to/MTGA-Companion
```

**3. Install the service (requires sudo):**
```bash
sudo ./mtga-companion service install
```

**4. Start the service:**
```bash
sudo ./mtga-companion service start
```

**5. Verify it's running:**
```bash
./mtga-companion service status
```

**Service Location:**
- Binary: System will use the current binary location
- Systemd Unit: `/etc/systemd/system/MTGACompanionDaemon.service`
- Logs: `journalctl -u MTGACompanionDaemon -f`

**Alternative using systemctl:**
```bash
# Check status
sudo systemctl status MTGACompanionDaemon

# Start
sudo systemctl start MTGACompanionDaemon

# Stop
sudo systemctl stop MTGACompanionDaemon

# Enable auto-start on boot
sudo systemctl enable MTGACompanionDaemon
```

## Service Management

### Check Status

```bash
# All platforms
./mtga-companion service status
```

### Start Service

```bash
# macOS/Linux
./mtga-companion service start

# Windows (as Administrator)
.\mtga-companion.exe service start
```

### Stop Service

```bash
# macOS/Linux
./mtga-companion service stop

# Windows (as Administrator)
.\mtga-companion.exe service stop
```

### Restart Service

```bash
# macOS/Linux
./mtga-companion service restart

# Windows (as Administrator)
.\mtga-companion.exe service restart
```

## Verification

### 1. Check Service Status

```bash
./mtga-companion service status
```

Should show "Status: ✓ Running"

### 2. Test WebSocket Connection

The daemon runs a WebSocket server on port 9999. You can test connectivity:

**Using curl (if available):**
```bash
curl http://localhost:9999/status
```

**Using the GUI:**
1. Launch MTGA Companion GUI
2. Go to Settings → Daemon Connection
3. Status should show "Connected to Daemon" (green)

### 3. Check Logs

**macOS:**
```bash
tail -f ~/Library/Logs/MTGACompanionDaemon.log
```

**Windows:**
- Open Event Viewer
- Navigate to Windows Logs → Application
- Filter for source "MTGACompanionDaemon"

**Linux:**
```bash
journalctl -u MTGACompanionDaemon -f
```

### 4. Play a Match

1. Ensure daemon is running
2. Play an MTGA match
3. Check logs for parsing activity
4. Open GUI and verify statistics updated

## Troubleshooting

### Service Won't Start

**Check logs:**
```bash
# macOS
cat ~/Library/Logs/MTGACompanionDaemon.log

# Linux
journalctl -u MTGACompanionDaemon -n 50

# Windows
# Check Event Viewer
```

**Common issues:**
- Port 9999 already in use → Configure different port
- Permissions error → Run install as Administrator/sudo
- Binary not found → Ensure binary exists at installation path

### GUI Can't Connect to Daemon

1. **Verify daemon is running:**
   ```bash
   ./mtga-companion service status
   ```

2. **Check port:**
   Default is 9999. Verify in Settings → Daemon Connection

3. **Test connectivity:**
   ```bash
   curl http://localhost:9999/status
   ```

4. **Restart daemon:**
   ```bash
   ./mtga-companion service restart
   ```

5. **Check firewall:**
   Ensure port 9999 is not blocked by firewall

### Permission Denied Errors

**macOS/Linux:**
```bash
# Ensure binary is executable
chmod +x ./mtga-companion

# Install with sudo (Linux)
sudo ./mtga-companion service install
```

**Windows:**
```powershell
# Run PowerShell as Administrator
# Right-click → "Run as Administrator"
```

### Service Doesn't Auto-Start on Boot

**macOS:**
```bash
# Check if Launch Agent is loaded
launchctl list | grep MTGACompanionDaemon

# If not loaded:
launchctl load ~/Library/LaunchAgents/MTGACompanionDaemon.plist
```

**Windows:**
```powershell
# Check service startup type
sc.exe qc MTGACompanionDaemon

# Set to automatic:
sc.exe config MTGACompanionDaemon start= auto
```

**Linux:**
```bash
# Enable auto-start
sudo systemctl enable MTGACompanionDaemon

# Verify
sudo systemctl is-enabled MTGACompanionDaemon
```

### Database Lock Errors

If you see "database is locked" errors:

1. **Stop all instances:**
   ```bash
   ./mtga-companion service stop
   killall mtga-companion  # macOS/Linux
   ```

2. **Restart daemon:**
   ```bash
   ./mtga-companion service start
   ```

3. **Avoid running both standalone and daemon:**
   Choose either daemon mode OR standalone mode, not both

## Uninstallation

### Stop and Remove Service

**All Platforms:**
```bash
# Stop the service
./mtga-companion service stop

# Uninstall the service
./mtga-companion service uninstall
```

Expected output:
```
✓ Service uninstalled successfully
```

### Clean Up Files

**macOS:**
```bash
# Remove logs
rm ~/Library/Logs/MTGACompanionDaemon.log

# Launch Agent is automatically removed by uninstall
```

**Windows:**
```powershell
# Logs are in Event Viewer (no files to remove)
```

**Linux:**
```bash
# Remove logs
sudo journalctl --vacuum-time=1s -u MTGACompanionDaemon

# Systemd unit is automatically removed by uninstall
```

### Remove Binary (Optional)

If you want to completely remove MTGA Companion:

```bash
# macOS/Linux
rm ./mtga-companion

# Windows
del mtga-companion.exe
```

## Advanced Configuration

### Custom Port

To run daemon on a different port:

**Option 1: Via command line**
```bash
./mtga-companion daemon --port=8888
```

**Option 2: Configure in service**

Edit the service configuration to include the port flag:
- macOS: Edit `~/Library/LaunchAgents/MTGACompanionDaemon.plist`
- Windows: Modify service with `sc.exe config`
- Linux: Edit `/etc/systemd/system/MTGACompanionDaemon.service`

Then update GUI Settings → Daemon Connection → Port to match.

### Custom Log Path

By default, daemon auto-detects your MTGA log location. To specify manually:

```bash
./mtga-companion daemon --log-path="/path/to/Player.log"
```

### Multiple MTGA Installations

If you have MTGA installed in multiple locations:

```bash
# Instance 1 (default)
./mtga-companion daemon --port=9999

# Instance 2 (custom port)
./mtga-companion daemon --port=9998 --log-path="/path/to/other/Player.log"
```

## Getting Help

If you encounter issues:

1. Check logs (see [Verification](#verification) section)
2. Review [Troubleshooting](#troubleshooting) section
3. Open an issue on GitHub: https://github.com/RdHamilton/MTGA-Companion/issues
4. Include:
   - OS and version
   - Daemon logs
   - Service status output
   - Steps to reproduce

## Next Steps

After installing the daemon:

1. ✅ Daemon installed and running
2. ✅ Launch the GUI
3. ✅ Verify connection in Settings → Daemon Connection
4. ✅ Play MTGA and watch statistics update in real-time!

For GUI usage, see the main [README.md](../README.md).
