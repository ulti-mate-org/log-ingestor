# log-ingestor

`log-ingestor` is a lightweight, efficient **log ingestion pipeline** built in Go, designed for **small to medium volumes** of structured JSON logs.

Its focus is on **simplicity**, **performance**, and **a minimal operational footprint**.

## Provides

* **Go** (net/http + Prometheus client)
* **Elasticsearch** (data store)
* **Kibana** (log exploration UI)
* Future: **Fluent Bit** (avoid sending logs from backend)

## Usage

The whole pipeline is built on docker containers. After cloning the repo you can simply run the setup_logging.sh script and it all should work.
```
git clone https://github.com/ulti-mate-org/log-ingestor && cd log-ingestor
chmod +x generate_certs.sh
./generate_certs.sh
chmod +x setup_logging.sh
./setup_logging.sh
```

## Setup

### 1. Generate TLS certificates

```
./generate_certs.sh
```

### 2. Start Whole Stack (with TLS enabled)

```
./setup_logging.sh
```

### 3. Log in to Kibana

Open:

```
https://localhost:5601/
```

Credentials:

* **Username:** `elastic`
* **Password:** value of your `ELASTIC_PASSWORD` environment variable (printed by setup_logging.sh)

## Note

In order to have the pipeline working you will need to be sending the logs to the /ingest endpoint of the log-ingestor.
To checkout a simple implementation of a working logger see [this *logger template* in javascript](logger.template.js).

## Roadmap / TODO

* [ ] Integrate Fluent Bit for file/tail ingestion
* [x] Add batching support before indexing to ES
* [x] Add retry/backoff logic on failures
* [ ] Add structured log validation schema
* [ ] Add querying or streaming endpoints (optional future work)
* [ ] Add configuration parsing and seeding (no dotenv)

### References:

- https://docs.fluentbit.io/manual
- https://pkg.go.dev/github.com/elastic/go-elasticsearch/esapi
- https://www.elastic.co/docs/api/doc/kibana/
- https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#hdr-Metrics 
- https://www.elastic.co/docs/deploy-manage/deploy/self-managed/configure-elasticsearch

