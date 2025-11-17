#!/bin/bash
# Stop any running MTGA Companion processes

# Kill all mtga-companion processes (daemon and others)
pkill -9 -f "mtga-companion" 2>/dev/null || true

# Also kill anything using port 9999
PID=$(lsof -ti :9999 2>/dev/null)
if [ -n "$PID" ]; then
    echo "Killing process $PID using port 9999..."
    kill -9 $PID 2>/dev/null || true
fi

# Wait for port to be released
for i in {1..5}; do
    if ! lsof -i :9999 >/dev/null 2>&1; then
        echo "✓ All MTGA Companion processes stopped"
        exit 0
    fi
    sleep 0.5
done

echo "✓ Cleanup complete"
exit 0
