# CLI Flag Migration Guide

This document maps old flags to new standardized flags for MTGA-Companion.

## Migration Summary

| Old Flag | New Flag | Shorthand | Notes |
|----------|----------|-----------|-------|
| `-gui` | `-gui-mode` | `-g` | Launch GUI mode |
| `-debug` | `-debug-mode` | `-d` | Enable debug logging |
| `-cache` | `-cache-enabled` | | Enable caching |
| `-poll-interval` | `-log-poll-interval` | | Log polling interval |
| `-enable-metrics` | `-enable-metrics` | | (No change) |
| `-use-file-events` | `-log-use-fsnotify` | | Use fsnotify |
| `-draft-overlay` | `-draft-overlay-mode` | | Launch draft overlay |
| `-set-file` | `-overlay-set-file` | | Path to set file |
| `-log-path` | `-log-file-path` | | Path to log file |
| `-overlay-set` | `-overlay-set-code` | | Set code |
| `-overlay-format` | `-overlay-format` | | (No change) |
| `-overlay-resume` | `-overlay-resume` | | (No change) |
| `-overlay-lookback` | `-overlay-lookback-hours` | | Lookback hours |
| `-cache-ttl` | `-cache-ttl` | | (No change) |
| `-cache-max-size` | `-cache-max-size` | | (No change) |

## Backward Compatibility

Old flags will continue to work but will print deprecation warnings:
```
⚠️  Warning: Flag '-gui' is deprecated. Use '-gui-mode' or '-g' instead.
```

Old flags will be removed in v2.0.0.

## Examples

### Old Syntax (Still Works)
```bash
mtga-companion -debug -gui -cache -poll-interval 5s
```

### New Syntax (Recommended)
```bash
mtga-companion -debug-mode -gui-mode -cache-enabled -log-poll-interval 5s
```

### New Syntax with Shortcuts
```bash
mtga-companion -d -g -cache-enabled -log-poll-interval 5s
```

## Export Command Flags

Export commands will use consistent flags across all export types:

| Flag | Shorthand | Description |
|------|-----------|-------------|
| `-start` | `-s` | Start date (YYYY-MM-DD) |
| `-end` | `-e` | End date (YYYY-MM-DD) |
| `-format` | `-f` | Format filter |
| `-event` | | Event name filter |
| `-output` | `-o` | Output file path |
| `-type` | `-t` | Export type (csv/json) |

