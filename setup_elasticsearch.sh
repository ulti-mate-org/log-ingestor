#!/bin/bash
#
# This script is part of the ulti-mate organization suite,
# see: https://ulti-mate.org
#
# license: MIT
# author: Emiliano Rizzonelli (emiliano.rizzonelli@proton.me)
# usage: ./setup_elasticsearch.sh

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
CERT_DIR="$ROOT_DIR/certs"
COMPOSE_FILE="docker-compose.secure.yml"
ES_URL="https://localhost:9200"
ENV_FILE="$ROOT_DIR/.env"

echo "starting elasticsearch..."
docker compose -f "$COMPOSE_FILE" up -d elasticsearch

echo "waiting for elasticsearch to be reachable"
for i in {1..60}; do
    if curl  --cacert "$CERT_DIR/elastic-cert.crt" "$ES_URL/_cluster/health"; then
        echo "elasticsearch is up!"
        break
    fi
	echo "miss"
    sleep 2
done

if ! curl --cacert "$CERT_DIR/elastic-cert.crt" "$ES_URL/_cluster/health"; then
    echo "ERROR: elasticsearch did not become ready in time."
    exit 1
fi 

echo "resetting built-in passwords"

ELASTIC_PASS=$(docker exec -i elasticsearch \
    bin/elasticsearch-reset-password -u elastic -b --url "$ES_URL" \
    | awk -F': ' '/New value/ {print $2}' | tr -d '[:space:]')

KIBANA_PASS=$(docker exec -i elasticsearch \
    bin/elasticsearch-reset-password -u kibana_system -b --url "$ES_URL" \
    | awk -F': ' '/New value/ {print $2}' | tr -d '[:space:]')

echo "writing .env file"

echo "# auto-generated on $(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> "$ENV_FILE"
echo "ELASTIC_PASSWORD=$ELASTIC_PASS" >> "$ENV_FILE"
echo "KIBANA_PASSWORD=$KIBANA_PASS" >> "$ENV_FILE"

echo "restarting services so they use the new credentials"
docker compose -f "$COMPOSE_FILE" up -d

echo
echo "elastic password:       $ELASTIC_PASS"
echo "kibana_system password: $KIBANA_PASS"

