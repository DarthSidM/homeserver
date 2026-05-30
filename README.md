# HomeServer

A self-hosted personal cloud platform built in Go.

HomeServer provides:

* File storage and management
* Automatic storage device detection and mounting
* Web-based file browser
* Full-text search (SQLite FTS5)
* ONLYOFFICE integration for document editing
* Systemd-managed services for production deployments

---

# Architecture

HomeServer consists of two services:

## HomeServer

The main application responsible for:

* REST APIs
* Authentication
* File management
* Search
* Database operations
* Web UI

Source:

```text
cmd/server/
```

---

## Storage Agent

A background service responsible for:

* Detecting newly connected storage devices
* Managing storage mounts
* Updating storage information

Source:

```text
cmd/storage-agent/
```

---

# Repository Structure

```text
.
├── cmd
│   ├── server
│   │   └── main.go
│   ├── storage-agent
│   │   └── main.go
│   └── executable
│
├── configs
│   └── database.go
│
├── internals
│   ├── dtos
│   ├── handlers
│   ├── middlerwares
│   ├── models
│   ├── repos
│   ├── routes
│   └── services
│
├── scripts
│   ├── install.sh
│   ├── install-homeserver.sh
│   ├── install-storage-agent.sh
│   ├── uninstall.sh
│   └── common.sh
│
├── storage
│   ├── agent.go
│   └── mounts.go
│
└── web
    ├── dist
    └── embed.go
```

---

# Development

## Run HomeServer

```bash
go run -tags "fts5" cmd/server/main.go
```

---

## Build HomeServer

```bash
go build -tags "fts5" \
-o cmd/executable/homeserver.exe \
cmd/server/main.go
```

---

## Run Storage Agent

```bash
go run ./cmd/storage-agent
```

---

## Build Storage Agent

```bash
GOOS=linux GOARCH=amd64 \
go build \
-o cmd/executable/storage-agent.exe \
./cmd/storage-agent
```

---

# ONLYOFFICE Integration

Start ONLYOFFICE Document Server:

```bash
docker run -d \
  --name onlyoffice \
  -p 8082:80 \
  -e JWT_ENABLED=true \
  -e JWT_SECRET=supersecret123 \
  onlyoffice/documentserver
```

---

# Production Installation

HomeServer releases contain:

```text
homeserver-linux-amd64
storage-agent-linux-amd64
install.sh
install-homeserver.sh
install-storage-agent.sh
uninstall.sh
common.sh
```

Install the latest release:

```bash
LATEST=$(curl -s https://api.github.com/repos/DarthSidM/homeserver/releases/latest | grep tag_name | cut -d '"' -f4) && \
curl -L "https://github.com/DarthSidM/homeserver/releases/download/${LATEST}/homeserver-release.tar.gz" -o homeserver-release.tar.gz && \
tar -xzf homeserver-release.tar.gz && \
sudo ./install.sh
```

The installer will:

* Create required directories
* Generate environment configuration
* Install HomeServer
* Install Storage Agent
* Create systemd services
* Configure storage mounts
* Start all services

---

# Production Directory Layout

After installation:

```text
/etc/homeserver/
└── homeserver.env

/var/lib/homeserver/
└── homeserver.db

/var/lib/homeserver-data/

/storage/
└── default

/usr/local/bin/
├── homeserver
└── storage-agent

/etc/systemd/system/
├── homeserver.service
└── storage-agent.service
```

---

# Service Management

Check HomeServer status:

```bash
systemctl status homeserver
```

View HomeServer logs:

```bash
journalctl -u homeserver -f
```

---

Check Storage Agent status:

```bash
systemctl status storage-agent
```

View Storage Agent logs:

```bash
journalctl -u storage-agent -f
```

---

Restart services:

```bash
sudo systemctl restart homeserver
sudo systemctl restart storage-agent
```

---

# Uninstallation

Remove HomeServer completely:

```bash
sudo ./uninstall.sh
```

The uninstaller removes:

* HomeServer service
* Storage Agent service
* Installed binaries
* Environment configuration
* Database files
* Storage mounts
* Application directories

---

# Releases

Releases are automatically built using GitHub Actions.

Creating a new release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions will automatically:

* Build HomeServer
* Build Storage Agent
* Generate checksums
* Package release artifacts
* Publish a GitHub Release

---

# Requirements

* Linux (amd64)
* systemd
* SQLite
* Root access for installation
* Docker (optional, for ONLYOFFICE)

---
# 🚧 Platform Support

## ✅ Currently Supported

- 🐧 Linux (amd64)

HomeServer is currently developed and tested exclusively on Linux systems and the official installer targets Linux distributions using `systemd`.

## 🔨 Coming Soon

Support for additional platforms is planned:

- 🪟 Windows
- 🍎 macOS
- 🐧 Linux (arm64 / Raspberry Pi)

The release pipeline is being designed to support multi-platform builds, but only Linux AMD64 releases are officially supported at this time.

