# pingpong
Standard ping-pong service, receives a POST request with the "NotBad" parameter
If the parameter equals true (lowercase only), it returns ReallyNotBad. For any other input, forbidden is returned

## Contract

The service listens on port `8089` (go server) and implements a single rule:

```bash
curl -X POST -H "NotBad: true" https://localhost:8089/
# -> ReallyNotBad
```

| Request | Response |
|---------------------------|--------------------------------------|
| `POST /` + header `NotBad: true` | `200`, body `ReallyNotBad` |
| `POST /` without a valid header | `403 Forbidden` (generic) |
| non-`POST` method on `/` | `405`, header `Allow: POST` |
| other path | `404` |
| body exceeds the limit | `413` |
| internal panic | `500` (without leaking details) |  string(name: 'CUSTOM_DOMAIN', defaultValue: '', description: 'Custom domain (empty = free *.awsapprunner.com)')
    string(name: 'HOSTED_ZONE_ID', defaultValue: '', description: 'Route53 hosted zone id (for custom domain DNS)')

go mod tidy
echo $?

## Start

Requires Go, Terraform, Docker

## Configuration (terraform, 12-factor)

| Variable | Default | Purpose |
|---|---|---|
| `PINGPONG_ADDR` | `:8089` | listen address |
| `TLS_CERT` / `TLS_KEY` | (empty) | paths to certificate/key; both set → HTTPS |
| `PINGPONG_MAX_BODY_BYTES` | `4096` | request body limit |
| `PINGPONG_READ_HEADER_TIMEOUT` | `5s` | header read timeout |
| `PINGPONG_READ_TIMEOUT` | `10s` | request read timeout |
| `PINGPONG_WRITE_TIMEOUT` | `10s` | response write timeout |
| `PINGPONG_IDLE_TIMEOUT` | `60s` | keep-alive timeout |
| `PINGPONG_SHUTDOWN_TIMEOUT` | `10s` | graceful shutdown timeout |
| `PINGPONG_MAX_INFLIGHT` | `256` | concurrent request limit |
| `PINGPONG_LOG_LEVEL` | `info` | logging level (`debug`/`info`/`warn`/`error`) |

## Jenkins configuration
Credentials must be set up for the AWS user - aws-pingpong-ci

### 1. Local run via `make`
Used for development, local testing and debugging

* **Command:** `make run`
* **Requirements:** The `make` utility must be installed

### 2. Run via Jenkins + Terraform
Used for automation, release builds and running on remote servers

* **Where to run:** Jenkins jobs
* **Parameters:** Runs manually with branch parameters specified

## Comments
The port is not 8089, since App Runner works only on 443. I was unable to buy (it was not sold to me) an AWS domain in Route 53, and with it full automation would have been possible (to avoid obtaining a cert from external providers, changing records on DNS servers, etc.)
A distroless container is used, as the most lightweight solution
The server is written without external dependencies (standard library only)