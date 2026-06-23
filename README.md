# GophKeeper

GophKeeper is a client-server password manager: store logins, text notes, bank cards, and files securely, and access them from a CLI on Windows, Linux, or macOS.

## What it does

- **Register and log in** to your account on the server
- **Save private data** — credentials, plain text, binary files, bank cards, OTP seeds
- **Sync across devices** — the same account on multiple machines stays up to date
- **Decrypt only on your machine** — the master password never leaves the client; the server stores ciphertext only

## How it works

```mermaid
flowchart LR
  subgraph client [CLI on your machine]
    MasterPass[master password]
    Encrypt[encrypt secrets]
    CLI[gophkeeper CLI]
    MasterPass --> Encrypt
    CLI --> Encrypt
  end

  subgraph server [Server]
    Auth[register / login]
    Store[ciphertext only]
    Auth --> Store
  end

  CLI -->|gRPC + TLS| Auth
  Encrypt -->|encrypted blobs| Store
  Store -->|ciphertext| CLI
```

You use two passwords for different jobs:

- **Account password** — proves who you are to the server; you get a session token after register or login.
- **Master password** — derives the encryption key locally; secrets are encrypted before upload and decrypted after download. The server never sees plaintext.

When you log in or run `sync`, the client asks the server for everything that changed since the last sync and merges it locally (last-write-wins by version).

## Quick start

**You need:** Go 1.26+ and PostgreSQL.

**1. Start the server**

```bash
export DATABASE_DSN="postgres://user:pass@localhost:5432/gophkeeper?sslmode=disable"
export JWT_SECRET="change-me-in-production"
go run ./cmd/server -a localhost:9090
```

**2. Build the client** (optional; or use `go run ./cmd/client`)

```bash
go build -o gophkeeper ./cmd/client
```

**3. Register and add a secret**

```bash
./gophkeeper register --login alice --password secret --master mymaster
./gophkeeper secret create --name github --type credentials \
  --login user --password pass --master mymaster
./gophkeeper secret list
```

Login on another machine pulls your data after sync:

```bash
./gophkeeper login --login alice --password secret --master mymaster
```

Client settings and tokens are stored in `~/.gophkeeper/`.

---

Yandex Practicum Advanced Go — graduation project.
