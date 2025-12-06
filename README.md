# Mini KV Store Go

**A production-ready, segmented key-value storage engine written in Go**

[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://golang.org)
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
It's a high-performance, append-only key-value storage engine with HTTP API capabilities, implementing core database concepts like segmented logs, compaction, bloom filters, index snapshots, and crash recovery.

### Why This Project?

This is a direct translation of the Rust version to Go, maintaining the same architecture and features while leveraging Go's simplicity and excellent concurrency primitives.

---

## ✨ Features

### Core Storage Engine
- 🔐 **Durable & crash-safe** - Append-only log with fsync guarantees
- 📦 **Segmented architecture** - Automatic rotation when segments reach size limits
- ⚡ **Lightning-fast reads** - O(1) lookups via in-memory HashMap index
- 🗜️ **Manual compaction** - Space reclamation on demand
- ✅ **Data integrity** - CRC32 checksums on every record
- 💾 **Index snapshots** - Fast restarts without full replay
- 🪦 **Tombstone deletions** - Efficient deletion in append-only architecture
- 🌸 **Bloom filters** - Optimized negative lookups

### Production Ready
- 🌐 **HTTP REST API** - Server built with Gorilla Mux
- 🖥️ **Interactive CLI** - REPL for testing and exploration
- 📊 **Metrics endpoint** - `/metrics` for monitoring
- 🩺 **Health checks** - `/health` endpoint for load balancers
- 🛑 **Graceful shutdown** - SIGTERM/SIGINT handling with snapshot save
- 🧪 **Comprehensive tests** - Unit and integration test suites
- 🐳 **Docker support** - Multi-container deployment with docker-compose
- 🔧 **CI/CD pipeline** - Automated testing and builds
- 🚦 **Rate limiting** - Configurable request body limits

### Developer Experience
- 📖 **Clean code** - Idiomatic Go with standard library
- 🎨 **Modular design** - Clear separation of concerns
- 🛠️ **Makefile included** - Simple commands for common tasks
- ⚙️ **Config via env vars** - Easy deployment configuration

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Make** - For build automation (optional)
- **Git** - For cloning the repository

### Installation

```bash
# Clone the repository
git clone https://github.com/whispem/mini-kvstore-go
cd mini-kvstore-go

# Download dependencies
go mod download

# Build binaries
make build

# Run tests
make test
```

### Running the CLI

```bash
# Start the interactive REPL
make run

# Or directly:
go run ./cmd/kvstore
```

**CLI Commands:**

```bash
mini-kvstore-go (type help for instructions)
> set name Alice          # Store a key-value pair
OK

> get name                # Retrieve a value
Alice

> list                    # List all keys
  name

> stats                   # Show storage statistics
Store Statistics:
  Keys: 1
  Segments: 1
  Total size: 0.00 MB
  Active segment: 1
  Oldest segment: 0

> compact                 # Reclaim space
Compaction finished

> quit                    # Exit (saves snapshot)
```

### Running the HTTP Server

```bash
# Start the volume server on port 9002
make server

# Or with custom configuration
PORT=9000 VOLUME_ID=my-vol DATA_DIR=./data go run ./cmd/volume-server
```

---

## 🌐 REST API Documentation

### Health Check

```bash
GET /health

# Response (200 OK)
{
  "status": "healthy",
  "volume_id": "vol-1",
  "keys": 42,
  "segments": 2,
  "total_mb": 1.5,
  "uptime_secs": 3600
}
```

### Metrics

```bash
GET /metrics

# Response (200 OK)
{
  "total_keys": 1000,
  "total_segments": 3,
  "total_bytes": 1572864,
  "total_mb": 1.5,
  "active_segment_id": 3,
  "oldest_segment_id": 0,
  "volume_id": "vol-1",
  "uptime_secs": 3600,
  "avg_value_size_bytes": 1572.864
}
```

### Store a Blob

```bash
POST /blobs/:key
Content-Type: application/octet-stream

# Example
curl -X POST http://localhost:9002/blobs/user:123 \
  -H "Content-Type: application/octet-stream" \
  -d "Hello, World!"

# Response (201 Created)
{
  "key": "user:123",
  "etag": "3e25960a",
  "size": 13,
  "volume_id": "vol-1"
}
```

### Retrieve a Blob

```bash
GET /blobs/:key

# Example
curl http://localhost:9002/blobs/user:123

# Response (200 OK)
Hello, World!

# Not Found (404)
{
  "error": "Blob not found"
}
```

### Delete a Blob

```bash
DELETE /blobs/:key

# Example
curl -X DELETE http://localhost:9002/blobs/user:123

# Response (204 No Content)
```

### List All Blobs

```bash
GET /blobs

# Response (200 OK)
[
  "user:123",
  "user:456",
  "config:settings"
]
```

---

## 🏗️ Architecture

### System Overview

```
┌─────────────────────────────────────────────────┐
│              Client Applications                 │
│         (CLI, HTTP Clients, Go API)             │
└────────────────────┬────────────────────────────┘
                     │
         ┌───────────▼───────────┐
         │     HTTP Server       │
         │    (Gorilla Mux)      │
         └───────────┬───────────┘
                     │
         ┌───────────▼───────────┐
         │   BlobStorage Layer   │
         │   (High-level API)    │
         └───────────┬───────────┘
                     │
         ┌───────────▼───────────┐
         │      KVStore Core     │
         │   (Storage Engine)    │
         └───────────┬───────────┘
                     │
    ┌────────────────┼────────────────┐
    │                │                │
    ▼                ▼                ▼
┌────────┐    ┌────────────┐    ┌──────────┐
│ Index  │    │  Segment   │    │  Bloom   │
│  Map   │    │  Manager   │    │  Filter  │
└────────┘    └──────┬─────┘    └──────────┘
                     │
         ┌───────────▼───────────┐
         │    Segment Files      │
         │ segment-0.dat         │
         │ segment-1.dat         │
         │ index.snapshot        │
         └───────────────────────┘
```

### On-Disk Format

Each segment file contains a sequence of records:

```
╔════════════════════════════════════════════╗
║              Segment Record                ║
╠════════════════════════════════════════════╣
║  MAGIC      │ 2 bytes │ 0xF0 0xF1         ║
║  op_code    │ 1 byte  │ 1=SET, 2=DELETE   ║
║  key_len    │ 4 bytes │ u32 little-endian ║
║  val_len    │ 4 bytes │ u32 little-endian ║
║  key        │ N bytes │ UTF-8 string      ║
║  value      │ M bytes │ Binary data       ║
║  checksum   │ 4 bytes │ CRC32             ║
╚════════════════════════════════════════════╝
```

---

## 💻 Programmatic Usage

### Basic Operations

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/whispem/mini-kvstore-go/pkg/store"
)

func main() {
    // Open or create a store
    kvstore, err := store.Open("my_database")
    if err != nil {
        log.Fatal(err)
    }
    defer kvstore.Close()
    
    // Store data
    kvstore.Set("user:1:name", []byte("Alice"))
    kvstore.Set("user:1:email", []byte("alice@example.com"))
    
    // Retrieve data
    name, err := kvstore.Get("user:1:name")
    if err == nil {
        fmt.Printf("Name: %s\n", string(name))
    }
    
    // Delete data
    kvstore.Delete("user:1:email")
    
    // List all keys
    keys := kvstore.ListKeys()
    for _, key := range keys {
        fmt.Printf("Key: %s\n", key)
    }
    
    // Get statistics
    stats := kvstore.Stats()
    fmt.Printf("Keys: %d, Segments: %d\n", stats.NumKeys, stats.NumSegments)
    
    // Manual compaction
    kvstore.Compact()
    
    // Save index snapshot
    kvstore.SaveSnapshot()
}
```

### Using BlobStorage (Higher-Level API)

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/whispem/mini-kvstore-go/pkg/volume"
)

func main() {
    storage, err := volume.NewBlobStorage("data", "vol-1")
    if err != nil {
        log.Fatal(err)
    }
    defer storage.Close()
    
    // Store with metadata
    meta, err := storage.Put("image:123", []byte("<binary data>"))
    if err == nil {
        fmt.Printf("Stored: etag=%s, size=%d\n", meta.ETag, meta.Size)
    }
    
    // Retrieve
    data, err := storage.Get("image:123")
    if err == nil {
        fmt.Printf("Retrieved %d bytes\n", len(data))
    }
    
    // Delete
    storage.Delete("image:123")
}
```

---

## 🐳 Docker Deployment

### Single Container

```bash
# Build image
docker build -t mini-kvstore-go:latest .

# Run container
docker run -d \
  -p 9002:9002 \
  -e VOLUME_ID=vol-1 \
  -e DATA_DIR=/data \
  -v $(pwd)/data:/data \
  --name kvstore \
  mini-kvstore-go:latest
```

### Multi-Volume Cluster

```bash
# Start 3-node cluster
docker-compose up -d

# Nodes available at:
# - http://localhost:9001 (vol-1)
# - http://localhost:9002 (vol-2)
# - http://localhost:9003 (vol-3)

# View logs
docker-compose logs -f

# Stop cluster
docker-compose down
```

---

## 🧪 Testing

```bash
# Run all tests
make test

# Or with go directly
go test -v -race -cover ./...

# Run specific test
go test -v -run TestSetAndGet ./pkg/store

# Run benchmarks
make bench
```

---

## 📂 Project Structure

```
mini-kvstore-go/
├── cmd/
│   ├── kvstore/          # CLI binary
│   └── volume-server/    # HTTP server binary
├── pkg/
│   ├── store/            # Storage engine
│   │   ├── bloom.go
│   │   ├── compaction.go
│   │   ├── engine.go
│   │   ├── errors.go
│   │   ├── index.go
│   │   ├── record.go
│   │   ├── segment.go
│   │   ├── snapshot.go
│   │   └── stats.go
│   ├── volume/           # HTTP API layer
│   │   ├── handlers.go
│   │   ├── server.go
│   │   └── storage.go
│   └── config/           # Configuration
│       └── config.go
├── internal/
│   └── testutil/         # Test utilities
├── examples/             # Example programs
├── .github/
│   └── workflows/        # CI/CD
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

---

## 🛠️ Development

### Using the Makefile

```bash
make help           # Show all available commands
make build          # Build binaries
make test           # Run all tests
make bench          # Run benchmarks
make fmt            # Format code
make lint           # Run linter
make clean          # Clean build artifacts
make docker         # Build Docker image
make docker-up      # Start cluster
```

### Code Quality Standards

- **Formatting:** `go fmt` for consistent style
- **Linting:** `golangci-lint` for code quality
- **Testing:** Comprehensive test suite with race detection
- **CI:** Automated checks on every push

```bash
# Pre-commit checks
make pre-commit
```

---

## 🗺️ Roadmap

### Completed ✅
- [x] Append-only log architecture
- [x] In-memory index
- [x] Crash recovery & persistence
- [x] Manual compaction
- [x] CRC32 checksums
- [x] Interactive CLI/REPL
- [x] HTTP REST API
- [x] Docker support
- [x] CI/CD pipeline
- [x] Bloom filters
- [x] Index snapshots

### Planned 📋
- [ ] Background compaction
- [ ] Range queries
- [ ] Write-ahead log (WAL)
- [ ] Compression (LZ4/Zstd)
- [ ] Replication protocol
- [ ] gRPC API option
- [ ] Metrics export (Prometheus)

---

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Fork and clone
git clone https://github.com/YOUR_USERNAME/mini-kvstore-go
cd mini-kvstore-go

# Install dependencies
go mod download

# Install development tools
make install-tools

# Run tests
make test

# Format code
make fmt
```

---

## 📜 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

- **Original Rust version** - [mini-kvstore-v2](https://github.com/whispem/mini-kvstore-v2)
- **Database Internals** - Alex Petrov's book
- **DDIA** - Martin Kleppmann's book
- **Bitcask** - For the append-only log design
- **Go Community** - For excellent tooling and libraries

---

## 👤 Author

**Em' ([@whispem](https://github.com/whispem))**

This is a Go port of the original Rust implementation, maintaining the same architecture and features while leveraging Go's strengths.

---

## 📬 Contact & Support

- 🐛 **Issues:** [GitHub Issues](https://github.com/whispem/mini-kvstore-go/issues)
- 📧 **Email:** contact.whispem@gmail.com

---

**Built with ❤️ in Go**

[⬆ Back to Top](#mini-kv-store-go-)
