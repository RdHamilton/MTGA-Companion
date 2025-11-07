# Database Backup and Restore

The MTGA Companion application includes comprehensive backup and restore functionality for the SQLite database.

## Features

- **Manual Backup**: Create backups on demand
- **Automatic Backup Scheduling**: Set up automated backups via cron
- **Backup Verification**: Verify backup integrity automatically
- **Backup Listing**: List all available backups with metadata
- **Restore**: Restore database from any backup
- **Configurable Location**: Set custom backup directory

## Manual Backup

### Command Line

Create a backup:
```bash
mtga-companion backup create
```

Create a backup with a custom name:
```bash
mtga-companion backup create my-backup
```

List all backups:
```bash
mtga-companion backup list
```

Verify a backup:
```bash
mtga-companion backup verify /path/to/backup.db
```

Restore from a backup:
```bash
mtga-companion backup restore /path/to/backup.db
```

### Interactive Console

From the interactive console, type `backup` or `b`:
```
> backup
```

Then select from the menu:
1. Create backup
2. List backups
3. Verify backup

## Automatic Backup Scheduling

Automatic backups can be scheduled using cron (Linux/macOS) or Task Scheduler (Windows).

### Linux/macOS (cron)

Edit your crontab:
```bash
crontab -e
```

Add a line to create a daily backup at 2 AM:
```cron
0 2 * * * /path/to/mtga-companion backup create >> /path/to/backup.log 2>&1
```

Or create a weekly backup every Sunday at 3 AM:
```cron
0 3 * * 0 /path/to/mtga-companion backup create >> /path/to/backup.log 2>&1
```

### Windows (Task Scheduler)

1. Open Task Scheduler
2. Create a new task
3. Set trigger (e.g., daily at 2 AM)
4. Set action to run: `mtga-companion backup create`
5. Save the task

## Configuration

### Environment Variables

- `MTGA_DB_PATH`: Path to the database file (default: `~/.mtga-companion/data.db`)
- `MTGA_BACKUP_DIR`: Directory for backups (default: `~/.mtga-companion/backups`)

### Example

```bash
export MTGA_DB_PATH=/custom/path/data.db
export MTGA_BACKUP_DIR=/custom/backup/path
mtga-companion backup create
```

## Backup Location

By default, backups are stored in:
- Linux/macOS: `~/.mtga-companion/backups/`
- Windows: `%USERPROFILE%\.mtga-companion\backups\`

Backup files are named with timestamps: `backup_YYYYMMDD_HHMMSS.db`

## Backup Verification

All backups are automatically verified after creation. The verification process:
1. Opens the backup as a SQLite database
2. Verifies the connection
3. Queries the database to ensure it's valid

If verification fails, the backup file is automatically removed.

## Restore Process

When restoring from a backup:
1. The backup is verified
2. The current database is backed up (renamed with `.old.TIMESTAMP` suffix)
3. The backup is copied to the database location
4. The restored database is verified

This ensures you can always recover if something goes wrong during restore.

## Backup Management

### Listing Backups

The `list` command shows:
- Backup filename
- Full path
- File size (in MB)
- Modification date
- SHA-256 checksum

### Cleaning Up Old Backups

You can manually delete old backups from the backup directory. The application doesn't automatically clean up old backups, so you may want to set up a cleanup script:

```bash
# Keep only the last 30 days of backups
find ~/.mtga-companion/backups -name "backup_*.db" -mtime +30 -delete
```

## Best Practices

1. **Regular Backups**: Set up automatic daily or weekly backups
2. **Offsite Storage**: Periodically copy backups to external storage or cloud
3. **Test Restores**: Periodically test restoring from backups to ensure they work
4. **Monitor Disk Space**: Ensure the backup directory has sufficient space
5. **Keep Multiple Backups**: Don't rely on a single backup

## Troubleshooting

### Backup Creation Fails

- Ensure the backup directory exists and is writable
- Check disk space availability
- Verify the database file is not locked by another process

### Restore Fails

- Ensure the backup file exists and is readable
- Verify the backup file is not corrupted
- Check that the database directory is writable
- Ensure no other process has the database open

### Verification Fails

- The backup file may be corrupted
- The file may not be a valid SQLite database
- Try creating a new backup

