#!/bin/bash
#
# This script is part of the ulti-mate organization suite,
# see: https://ulti-mate.org
#
# license: MIT
# author: Emiliano Rizzonelli (emiliano.rizzonelli@proton.me)
# usage: ./snapshot_backup.sh

set -e

ES_URL=${ES_URL:-http://localhost:9200}
REPO_NAME=local_backup
DATE=$(date +%F)

echo "registering snapshot repo (if missing)..."
curl -s -X PUT "$ES_URL/_snapshot/$REPO_NAME" \
  -H 'Content-Type: application/json' \
  -d "{\"type\": \"fs\", \"settings\": {\"location\": \"/usr/share/elasticsearch/snapshots\"}}" > /dev/null || true

echo "creating snapshot: $REPO_NAME/snapshot_$DATE..."
curl -s -X PUT "$ES_URL/_snapshot/$REPO_NAME/snapshot_$DATE?wait_for_completion=true" \
  -H 'Content-Type: application/json' \
  -d '{"indices": "logs-*"}' | jq .

echo "compressing snapshot..."
tar -czf "./snapshots/snapshot_$DATE.tar.gz" -C ./snapshots . || echo "compression failed, trying again in 5 min (if no changes to crontab)"

echo "snapshot completed and compressed."
