# ulti-mate.org Logging

This repo contains the logging pipeline for the [**ulti-mate.org**](https://ulti-mate.org) infrastructure.

### Setup:

remember to add to crontab:
0 2 * * * .../backup/snapshot_backup.sh >> .../backup/snapshot.log 2>&1

### TODO:

implement tls and secure access to containers.
the repo contains some config files for fluent-bit, that is something to do later on.

### References:

- https://docs.fluentbit.io/manual
- https://pkg.go.dev/github.com/elastic/go-elasticsearch/esapi
- https://www.elastic.co/docs/api/doc/kibana/
