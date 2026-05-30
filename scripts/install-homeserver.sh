#!/bin/bash

set -e

echo "=== HomeServer Installation ==="

# --------------------------------------------------
# CONFIG
# --------------------------------------------------

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

BINARY_SOURCE="$SCRIPT_DIR/homeserver-linux-amd64"
BINARY_DEST="/usr/local/bin/homeserver"

DATA_DIR="/var/lib/homeserver"
DB_PATH="$DATA_DIR/homeserver.db"

ENV_FILE="/etc/homeserver/homeserver.env"

SERVICE_FILE="/etc/systemd/system/homeserver.service"

# --------------------------------------------------
# VALIDATION
# --------------------------------------------------

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root:"
    echo "sudo ./install.sh"
    exit 1
fi

if [ ! -f "$BINARY_SOURCE" ]; then
    echo "Error: HomeServer binary not found:"
    echo "$BINARY_SOURCE"
    exit 1
fi

if [ ! -f "$ENV_FILE" ]; then
    echo "Error: Shared environment file not found:"
    echo "$ENV_FILE"
    echo "Run install.sh instead of this script directly."
    exit 1
fi

# --------------------------------------------------
# DATA DIRECTORY
# --------------------------------------------------

echo "Creating HomeServer data directory..."

mkdir -p "$DATA_DIR"

chmod 755 "$DATA_DIR"

# --------------------------------------------------
# INSTALL BINARY
# --------------------------------------------------

echo "Installing HomeServer binary..."

install -m 755 "$BINARY_SOURCE" "$BINARY_DEST"

# --------------------------------------------------
# CREATE SYSTEMD SERVICE
# --------------------------------------------------

echo "Creating systemd service..."

cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=HomeServer
After=network.target storage-agent.service
Requires=storage-agent.service

[Service]
Type=simple
ExecStart=$BINARY_DEST
Restart=always
RestartSec=5
User=root
EnvironmentFile=$ENV_FILE

[Install]
WantedBy=multi-user.target
EOF

# --------------------------------------------------
# ENABLE + START
# --------------------------------------------------

echo "Reloading systemd..."

systemctl daemon-reload

echo "Enabling HomeServer..."

systemctl enable homeserver

echo "Starting HomeServer..."

systemctl restart homeserver

# --------------------------------------------------
# DONE
# --------------------------------------------------

echo
echo "======================================="
echo " HomeServer Installed Successfully "
echo "======================================="
echo
echo "Binary : $BINARY_DEST"
echo "Database: $DB_PATH"
echo "Service : homeserver"
echo
echo "Useful commands:"
echo "  systemctl status homeserver"
echo "  journalctl -u homeserver -f"
echo