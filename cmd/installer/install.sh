#!/bin/bash

set -e

echo "=== HomeServer Storage Setup ==="

# ---- CONFIG ----
STORAGE_ROOT="/storage"
DEFAULT_DATA_DIR="/var/lib/homeserver-data"
DEFAULT_MOUNT="$STORAGE_ROOT/default"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_SOURCE="$SCRIPT_DIR/../executable/storage-agent.exe"
BINARY_DEST="/usr/local/bin/storage-agent.exe"
SERVICE_FILE="/etc/systemd/system/storage-agent.service"

# ---- CHECK ROOT ----
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit 1
fi

# ---- CREATE DIRECTORIES ----
echo "Creating storage directories..."
mkdir -p "$STORAGE_ROOT"
mkdir -p "$DEFAULT_DATA_DIR"
mkdir -p "$DEFAULT_MOUNT"

# ---- SET PERMISSIONS ----
chmod 755 "$STORAGE_ROOT"
chmod 755 "$DEFAULT_DATA_DIR"
chmod 755 "$DEFAULT_MOUNT"

# ---- BIND MOUNT DEFAULT STORAGE ----
if ! mountpoint -q "$DEFAULT_MOUNT"; then
  echo "Creating bind mount for default storage..."
  mount --bind "$DEFAULT_DATA_DIR" "$DEFAULT_MOUNT"
else
  echo "Default mount already exists."
fi

# ---- PERSIST IN FSTAB ----
if ! grep -qs "$DEFAULT_MOUNT" /etc/fstab; then
  echo "Adding bind mount to /etc/fstab..."
  echo "$DEFAULT_DATA_DIR $DEFAULT_MOUNT none bind 0 0" >> /etc/fstab
else
  echo "Bind mount already present in fstab."
fi

# ---- INSTALL BINARY ----
echo "Installing storage-agent binary..."
if [ ! -f "$BINARY_SOURCE" ]; then
  echo "Error: storage-agent binary not found in current directory."
  exit 1
fi

cp "$BINARY_SOURCE" "$BINARY_DEST"
chmod +x "$BINARY_DEST"

# ---- CREATE SYSTEMD SERVICE ----
echo "Creating systemd service..."

cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=HomeServer Storage Agent
After=network.target

[Service]
ExecStart=$BINARY_DEST
Restart=always
RestartSec=5
User=root
Environment=ENV=production

[Install]
WantedBy=multi-user.target
EOF

# ---- RELOAD SYSTEMD ----
systemctl daemon-reload

# ---- ENABLE SERVICE ----
systemctl enable storage-agent

# ---- START SERVICE ----
systemctl restart storage-agent

echo "=== Installation Complete ==="
echo "Storage agent is now running."