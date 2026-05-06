#!/usr/bin/env bash
# Script to update HSReplay data
# Run via cron: 0 */6 * * * /path/to/update-data.sh

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
OUTPUT_FILE="$PROJECT_DIR/data/scraped_comps.json"

echo "[$(date)] Starting HSReplay data fetch..."

cd "$PROJECT_DIR"

# Run the fetcher
./bin/fetch -o "$OUTPUT_FILE"

if [ $? -eq 0 ]; then
    echo "[$(date)] Successfully updated $OUTPUT_FILE"
else
    echo "[$(date)] Failed to fetch data"
    exit 1
fi
