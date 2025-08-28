# go-demo

Small Go demos focused on **PKI, TLS, and crypto**.
Each demo is a separate runnable app under `cmd/` and shares helpers in `internal/`.

## Contents

* **certinfo** — CLI that prints useful X.509 certificate details (SAN, KU/EKU, SKI/AKI, OCSP/CRL/AIA, fingerprints).
* **certinfo-web** — Minimal HTTP server with an upload form that parses a cert and renders details via HTML templates.

> Add/remove demos as you build them. Keep shared logic in `internal/`.

---

## Repo layout

```
go-demo/
  go.mod
  README.md
  .gitignore
  .gitattributes
  .golangci.yml

  cmd/
    certinfo/
      main.go
    certinfo-web/
      main.go
      templates/
        layout.html
        index.html

  internal/
    pki/
      x509util.go      # parsing, summaries, ASN.1 helpers
```

---

## Prerequisites

* Go ≥ 1.21
* OpenSSL (for local test material)
* (Optional) `golangci-lint` for linting

  * Install to `~/go/bin`:

    ```bash
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
      | sh -s -- -b "$HOME/go/bin" v2.4.0
    ```
  * Ensure `~/go/bin` is on your `PATH`.

---

## Build & run

Using Make (recommended):

```bash
make build      # builds all demos to ./bin
make test       # go test ./...
make lint       # golangci-lint run
make clean      # removes ./bin
```

Or directly with Go:

```bash
go build -o bin/certinfo ./cmd/certinfo
go run ./cmd/certinfo path/to/cert.pem
```

---

## Generating example certs (do NOT commit keys)

Use these throwaway commands to create local test material under `examples/`:

```bash
mkdir -p examples

# 1) Root CA (self-signed, CA:TRUE)
openssl req -x509 -newkey rsa:2048 -nodes -days 365 \
  -keyout examples/root.key -out examples/root.crt \
  -subj "/CN=MiniPKI Root" \
  -addext "basicConstraints=critical,CA:TRUE,pathlen:2" \
  -addext "keyUsage=critical,keyCertSign,cRLSign"

# 2) Server key/CSR with SANs
openssl req -new -newkey rsa:2048 -nodes \
  -keyout examples/server.key -out examples/server.csr \
  -subj "/CN=mtls.local" \
  -addext "subjectAltName=DNS:mtls.local,DNS:localhost,IP:127.0.0.1"

# 3) Sign server cert with the root (create minimal ext file on the fly)
cat > examples/server.ext <<'EOF'
basicConstraints=CA:FALSE
keyUsage=critical,digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
subjectAltName=DNS:mtls.local,DNS:localhost,IP:127.0.0.1
authorityKeyIdentifier=keyid,issuer
subjectKeyIdentifier=hash
EOF

openssl x509 -req -in examples/server.csr -CA examples/root.crt -CAkey examples/root.key \
  -CAcreateserial -out examples/server.crt -days 365 -extfile examples/server.ext
```

> You now have:
>
> * `examples/root.crt` (trust anchor for clients),
> * `examples/server.crt` + `examples/server.key` (for servers),
> * **Never commit `.key` files**. The repo’s `.gitignore` excludes them.

---

## Running the demos

### certinfo (CLI)

```bash
go run ./cmd/certinfo examples/server.crt
# or
bin/certinfo examples/server.crt
```

### certinfo-web (HTTP server)

```bash
go run ./cmd/certinfo-web
# open http://local
```
