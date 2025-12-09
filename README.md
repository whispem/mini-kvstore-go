# Mini KV Store Go 🦦

**A production-ready, segmented key-value storage engine written in Go**

[![Go Version](https://img.shields.io/badge/go-1.21%2B-00ADD8.svg)](https://golang.org)
[![CI](https://img.shields.io/badge/CI-passing-brightgreen)](https://github.com/whispem/mini-kvstore-go/actions)
[![Docker](https://img.shields.io/badge/docker-ready-2496ED?logo=docker&logoColor=white)](https://github.com/whispem/mini-kvstore-go/blob/main/Dockerfile)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[Features](#-features) •
[Quick Start](#-quick-start) •
[Architecture](#-architecture) •
[API Documentation](#-api-documentation) •
[Contributing](#-contributing)

---

## 📚 About

Mini KV Store Go is a **Go port** of the original Rust implementation ([mini-kvstore-v2](https://github.com/whispem/mini-kvstore-v2)).  
It’s a high-performance, append-only key-value storage engine with HTTP API, implementing segmented logs, compaction, bloom filters, index snapshots, and crash recovery.

**Why This Project?**  
A direct translation from Rust to Go, keeping the original architecture, and showing that clear DB design patterns are language-agnostic.  
Leverages Go’s concurrency and simplicity for production-like services.

---

## ✨ Features

**Core Storage Engine**

- Durable & crash-safe (append-only log with fsync)
- Segmented architecture with automatic rotation
- O(1) fast reads via in-memory HashMap index
- Manual compaction for space reclamation
- CRC32 checksums for data integrity
- Index snapshots enable quick restarts
- Tombstone deletions efficiently mark deleted keys
- Bloom filters speed up negative lookups

**Production Ready**

- HTTP REST API (Gorilla Mux)
- Interactive CLI (REPL) for testing and exploration
- `/metrics` endpoint for monitoring
- `/health` endpoint for load balancers
- Graceful shutdown (SIGTERM/SIGINT with snapshot saving)
- Unit and integration test suites
- Docker & docker-compose support
- CI/CD pipeline for automated testing and builds
- Configurable request body limits

**Developer Experience**

- Clean, idiomatic Go code
- Modular, maintainable design
- Makefile for common tasks
- Configurable via environment variables

---

## 🚀 Quick Start

**Prerequisites:**  
- Go 1.21+ ([install](https://golang.org/doc/install))
- Make (optional)
- Git

```bash
git clone https://github.com/whispem/mini-kvstore-go
cd mini-kvstore-go
go mod download
make build
make test
```

**REPL CLI:**
```bash
make run
# > set key Alice
# > get key
# > list
# > stats
# > compact
# > quit
```

**HTTP server:**
```bash
make server
# or:
PORT=9000 VOLUME_ID=my-vol DATA_DIR=./data go run ./cmd/volume-server
```

---

## 🌐 API Documentation

- `/health` — server health
- `/metrics` — real-time stats
- `POST /blobs/:key` — store blob
- `GET /blobs/:key` — fetch blob
- `DELETE /blobs/:key` — delete blob
- `GET /blobs` — list all keys

Examples:
```bash
curl -X POST http://localhost:9002/blobs/user:123 -d "Hello, World!"
curl http://localhost:9002/blobs/user:123
curl -X DELETE http://localhost:9002/blobs/user:123
```

---

## 🏗️ Architecture

```
[CLI / HTTP Client]
        │
    [HTTP Server]
        │
   [Blob Storage]
        │
   [KVStore Core]
     /  |  \
[Index][Segments][Bloom]
```
- Append-only segments on disk, in-memory index and bloom for speed
- Snapshot & compaction for instant restart / low disk use

---

## 💻 Programmatic Usage

```go
store, err := pkgstore.Open("data")
store.Set("foo", []byte("bar"))
val, _ := store.Get("foo")
// ...
store.Compact()
store.SaveSnapshot()
```

---

## 🐳 Docker

**Standalone:**
```bash
docker build -t mini-kvstore-go:latest .
docker run -d -p 9002:9002 -v $(pwd)/data:/data --name kvstore mini-kvstore-go:latest
```
**Cluster:**
```bash
docker-compose up -d
# nodes: localhost:9001 ... :9003
docker-compose logs -f
```

---

## 🧪 Testing

```bash
make test
go test -v -cover ./...
```
- Unit & integration tests for core & API

---

## 📂 Project Structure

```
mini-kvstore-go/
├── cmd/
├── pkg/
├── internal/
├── examples/
├── .github/
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

---

## 🛠️ Development

- **Formatting:** `go fmt`
- **Linting:** `golangci-lint`
- **Tests:** Makefile or `go test ./...`
- **Pre-commit checks:** `make pre-commit`

---

## 🗺️ Roadmap

- [x] Segmented log, in-memory index
- [x] Compaction, crash safety
- [x] CLI, HTTP, Docker, CI
- [x] Bloom filters, index snapshots
- [ ] Background compaction
- [ ] Range queries
- [ ] WAL, compression
- [ ] Replication, gRPC, Prometheus

---

## 🤝 Contributing

- Found a bug? Create an issue!
- Suggest a feature or improvement
- PRs to docs, examples, or code are welcome

---

## 📜 License

MIT — see [LICENSE](LICENSE)

---

## 👤 Author

**Em' ([@whispem](https://github.com/whispem))** — Go port of the Rust engine, made to understand DBs from the inside.

---

Built with ❤️ in Go

[⬆ Back to Top](#mini-kv-store-go-)
