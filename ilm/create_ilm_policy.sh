#!/bin/bash
#
# This script is part of the ulti-mate organization suite,
# see: https://ulti-mate.org
#
# license: MIT
# author: Emiliano Rizzonelli (emiliano.rizzonelli@proton.me)
# usage: ./create_ilm_policy.sh

set -e

ES_URL=${ES_URL:-https://localhost:9200}
CA_CERT=${CA_CERT:-certs/elastic-cert.crt}
AUTH="-u elastic:${ELASTIC_PASSWORD}"

echo "creating ILM policy..."
curl --cacert "$CA_CERT" $AUTH -X PUT "$ES_URL/_ilm/policy/logs_policy" \
  -H 'Content-Type: application/json' \
  -d '{
    "policy": {
      "phases": {
        "hot": {
          "actions": {
            "rollover": { "max_age": "7d", "max_size": "1gb" }
          }
        },
        "warm": {
          "min_age": "7d",
          "actions": {
            "shrink": { "number_of_shards": 1 }
          }
        },
        "delete": {
          "min_age": "90d",
          "actions": { "delete": {} }
        }
      }
    }
  }'

echo
echo "creating index template..."
curl --cacert "$CA_CERT" $AUTH -X PUT "$ES_URL/_index_template/logs_template" \
  -H 'Content-Type: application/json' \
  -d '{
    "index_patterns": ["logs-*"],
    "template": {
      "settings": {
        "index.lifecycle.name": "logs_policy",
        "index.lifecycle.rollover_alias": "logs"
      },
      "aliases": {
        "logs": { "is_write_index": false }
      }
    }
  }'

echo
echo "creating initial index..."
curl --cacert "$CA_CERT" $AUTH -X PUT "$ES_URL/logs-000001" \
  -H 'Content-Type: application/json' \
  -d '{
    "aliases": {
      "logs": { "is_write_index": true }
    }
  }'

echo "ILM policy + template + initial index created."
