#!/bin/bash
#
# This script is part of the ulti-mate organization suite,
# see: https://ulti-mate.org
#
# license: MIT
# author: Emiliano Rizzonelli (emiliano.rizzonelli@proton.me)
# usage: ./create_ilm_policy.sh

set -e

ES_URL=${ES_URL:-http://localhost:9200}

echo "creating ILM policy..."
curl -X PUT "$ES_URL/_ilm/policy/logs_policy" \
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
            "shrink": { "number_of_shards": 1 },
            "forcemerge": { "max_num_segments": 1 }
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
echo "creating index template for logs..."
curl -X PUT "$ES_URL/_index_template/logs_template" \
  -H 'Content-Type: application/json' \
  -d '{
    "index_patterns": ["logs-*"],
    "template": {
      "settings": {
        "index.lifecycle.name": "logs_policy",
        "index.lifecycle.rollover_alias": "logs"
      }
    }
  }'

echo
echo "creating initial index alias..."
curl -X PUT "$ES_URL/logs-000001" \
  -H 'Content-Type: application/json' \
  -d '{
    "aliases": {
      "logs": { "is_write_index": true }
    }
  }'

echo "ILM policy and template created."
