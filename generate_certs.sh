#!/bin/bash
#
# This script is part of the ulti-mate organization suite,
# see: https://ulti-mate.org
#
# license: MIT
# author: Emiliano Rizzonelli (emiliano.rizzonelli@proton.me)
# usage: ./generate_certs.sh

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUT_DIR="${ROOT_DIR}/certs"
ES_IMAGE="docker.elastic.co/elasticsearch/elasticsearch:9.2.0"

echo "ROOT_DIR: $ROOT_DIR"
echo "OUT_DIR: $OUT_DIR"

if [ -d "$OUT_DIR" ] && [ "$(ls -A "$OUT_DIR" 2>/dev/null || true)" != "" ]; then
  ts=$(date -u +"%Y%m%dT%H%M%SZ")
  echo "backing up existing $OUT_DIR -> ${OUT_DIR}-backup-$ts"
  mv "$OUT_DIR" "${OUT_DIR}-backup-$ts"
fi

mkdir -p "$OUT_DIR"
chmod 0777 "$OUT_DIR"

run_certutil() {
  echo "running in docker: $*"
  docker run --rm -v "$OUT_DIR":/certs "$ES_IMAGE" sh -c "$*"
}

echo "creating CA..."
# create CA (non-interactive, no pass)
run_certutil 'elasticsearch-certutil ca --silent --out /certs/elastic-stack-ca.p12 --pass ""'

# TODO: migrate ip and dns to config file
echo "creating a certificate for nodes + kibana..."
run_certutil 'elasticsearch-certutil cert --silent --name "elastic-cert" --dns "elasticsearch" --dns "kibana" --dns "localhost" --ip "127.0.0.1" --ca /certs/elastic-stack-ca.p12 --ca-pass "" --out /certs/elastic-certificates.p12 --pass ""'

echo "extracting certificate and key (PEM) for kibana / server SSL..."
# extract cert and key using openssl (host executes these)
# if openssl not present in container, do it on host
openssl pkcs12 -in "$OUT_DIR/elastic-certificates.p12" -clcerts -nokeys -out "$OUT_DIR/elastic-cert.crt" -passin pass:
openssl pkcs12 -in "$OUT_DIR/elastic-certificates.p12" -nocerts -nodes -out "$OUT_DIR/elastic-cert.key" -passin pass:
openssl pkcs12 -in "$OUT_DIR/elastic-stack-ca.p12" -out "$OUT_DIR/elastic-stack-ca.crt" -nodes -nokeys -passin pass:

echo "combinie and list files"
cat "$OUT_DIR/elastic-cert.crt" "$OUT_DIR/elastic-cert.key" > "$OUT_DIR/elastic-cert-combined.pem" || true

ls -la "$OUT_DIR"

# The files are created by root inside the Elasticsearch container. Hand them to
# the current user through the same image, so no sudo (and no password prompt) is
# needed. Scripts that run unattended cannot answer a sudo prompt.
uid=$(id -u)
gid=$(id -g)
echo "giving $OUT_DIR to $(id -un) (uid:$uid gid:$gid)..."
docker run --rm --user 0:0 -v "$OUT_DIR":/certs --entrypoint sh "$ES_IMAGE" \
  -c "chown -R $uid:$gid /certs && chmod 755 /certs && chmod 644 /certs/*.p12 /certs/*.crt /certs/*.key 2>/dev/null || true"

echo "certificates are readable by elasticsearch."

echo "done."
