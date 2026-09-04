#!/bin/bash
#
# This script is part of the ulti-mate organization suite,
# see: https://ulti-mate.org
#
# license: MIT
# author: Emiliano Rizzonelli (emiliano.rizzonelli@proton.me)
# usage: ./setup_elasticsearch.sh
#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
CERT_DIR="$ROOT_DIR/certs"
COMPOSE_FILE="docker-compose.secure.yml"
ES_URL="https://localhost:9200"
ENV_FILE="$ROOT_DIR/.env"
PARENT_ENV_FILE="$ROOT_DIR/../.env"

# Load the orchestrator's .env so variables like ES_URL and the X-Pack
# encryption keys are available to docker compose when it runs from this
# directory. Generated credentials below are still written to ./.env.
if [ -f "$PARENT_ENV_FILE" ]; then
    set -a
    # shellcheck disable=SC1090
    source "$PARENT_ENV_FILE"
    set +a
fi

# Host-side curl/health checks hit the published Elasticsearch port.
# ES_URL (possibly from the orchestrator .env) is passed to docker compose
# for Kibana and used by docker exec inside the Elasticsearch container.
ES_URL_HOST="https://localhost:9200"

echo "starting elasticsearch..."
docker compose -f "$COMPOSE_FILE" up --build -d elasticsearch

echo "waiting for elasticsearch to be reachable"
for i in {1..60}; do
    if curl  --cacert "$CERT_DIR/elastic-cert.crt" "$ES_URL_HOST/_cluster/health"; then
        echo "elasticsearch is up!"
        break
    fi
	echo "miss"
    sleep 2
done

if ! curl --cacert "$CERT_DIR/elastic-cert.crt" "$ES_URL_HOST/_cluster/health"; then
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


echo "creating API key for Go logging service..."
API_KEY_JSON=$(curl -s --cacert "$CERT_DIR/elastic-cert.crt" -u "elastic:$ELASTIC_PASS" -X POST "$ES_URL_HOST/_security/api_key" -H "Content-Type: application/json" -d '{
  "name": "go-log-ingestor",
  "role_descriptors": {
    "log_ingestor_role": {
      "cluster": ["monitor"],
      "index": [{
        "names": ["logs-*"],
		"privileges": ["create_index","create","write","auto_configure"]
      }]
    }
  }
}')

ID=$(echo "$API_KEY_JSON" | jq -r '.id')
API_KEY_SECRET=$(echo "$API_KEY_JSON" | jq -r '.api_key')

ENCODED_KEY=$(printf "%s:%s" "$ID" "$API_KEY_SECRET" | base64 | tr -d '\n')

echo "writing .env file"

echo "# auto-generated on $(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> "$ENV_FILE"
echo "ELASTIC_PASSWORD=$ELASTIC_PASS" >> "$ENV_FILE"
echo "KIBANA_PASSWORD=$KIBANA_PASS" >> "$ENV_FILE"
echo "ELASTIC_API_KEY=$ENCODED_KEY" >> "$ENV_FILE"
echo "VERSION=$(git describe --tags --always)" >> "$ENV_FILE"
echo "COMMIT=$(git rev-parse HEAD)" >> "$ENV_FILE"
echo "BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> "$ENV_FILE"

echo "" >> "$ENV_FILE"

echo "restarting services so they use the new credentials"
docker compose -f "$COMPOSE_FILE" up --build -d log-ingestor kibana

echo
echo "elastic password:       $ELASTIC_PASS"
echo "kibana_system password: $KIBANA_PASS"
echo "elastic API key:        $ENCODED_KEY"
