#!/bin/bash
find_workspace_root() {
    local dir="$PWD"
    while [ "$dir" != "/" ]; do
        if [ -d "$dir/.agents" ] || [ -f "$dir/AGENTS.md" ]; then
            echo "$dir"
            return 0
        fi
        dir="$(dirname "$dir")"
    done
    echo "$PWD"
}

WORKSPACE_ROOT="$(find_workspace_root)"
DB_PATH="$WORKSPACE_ROOT/.agents/memory.db"
SCHEMA_PATH="/home/noxturne/agents/memory_schema.sql"

echo "Initializing JARVIS SQLite memory at: $DB_PATH"
mkdir -p "$(dirname "$DB_PATH")"

if ! command -v sqlite3 &> /dev/null; then
    echo "Error: sqlite3 command not found. Please install sqlite3."
    exit 1
fi

sqlite3 "$DB_PATH" < "$SCHEMA_PATH"
sqlite3 "$DB_PATH" "PRAGMA journal_mode=WAL;"

echo "SQLite memory database successfully initialized and configured in WAL mode."
