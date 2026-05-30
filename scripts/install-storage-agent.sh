#!/bin/bash

set -e

echo "=== HomeServer Storage Agent Installation ==="

# --------------------------------------------------
# CONFIG
# --------------------------------------------------

STORAGE_ROOT="/storage"
DEFAULT_DATA_DIR="/var/lib/homeserver-data"
DEFAULT_MOUNT="$STORAGE_ROOT/default"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

BINARY_SOURCE="$SCRIPT_DIR/storage-agent-linux-amd64"
BINARY_DEST="/usr/local/bin/storage-agent"

ENV_FILE="/etc/homeserver/homeserver.env"

SERVICE_FILE="/etc/systemd/system/storage-agent.service"

TARGET_USER="${SUDO_USER:-$USER}"

# --------------------------------------------------
# VALIDATION
# --------------------------------------------------

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root:"
    echo "sudo ./install.sh"
    exit 1
fi

if [ -z "$TARGET_USER" ] || [ "$TARGET_USER" = "root" ]; then
    echo "Unable to determine the target user."
    echo "Run installation using sudo from a normal user account."
    exit 1
fi

if [ ! -f "$BINARY_SOURCE" ]; then
    echo "Error: storage-agent binary not found:"
    echo "$BINARY_SOURCE"
    exit 1
fi

if [ ! -f "$ENV_FILE" ]; then
    echo "Error: shared environment file not found:"
    echo "$ENV_FILE"
    echo "Run install.sh instead of this script directly."
    exit 1
fi

# --------------------------------------------------
# STORAGE DIRECTORIES
# --------------------------------------------------

echo "Creating storage directories..."

mkdir -p "$STORAGE_ROOT"
mkdir -p "$DEFAULT_DATA_DIR"
mkdir -p "$DEFAULT_MOUNT"

chmod 755 "$STORAGE_ROOT"
chmod 755 "$DEFAULT_DATA_DIR"
chmod 755 "$DEFAULT_MOUNT"

chown "$TARGET_USER:$TARGET_USER" "$STORAGE_ROOT"
chown -R "$TARGET_USER:$TARGET_USER" "$DEFAULT_DATA_DIR"
chown "$TARGET_USER:$TARGET_USER" "$DEFAULT_MOUNT"

# --------------------------------------------------
# DEFAULT STORAGE MOUNT
# --------------------------------------------------

if ! mountpoint -q "$DEFAULT_MOUNT"; then
    echo "Creating default bind mount..."
    mount --bind "$DEFAULT_DATA_DIR" "$DEFAULT_MOUNT"
else
    echo "Default bind mount already exists."
fi

if ! grep -qs "$DEFAULT_MOUNT" /etc/fstab; then
    echo "Persisting bind mount in /etc/fstab..."
    echo "$DEFAULT_DATA_DIR $DEFAULT_MOUNT none bind 0 0" >> /etc/fstab
else
    echo "Bind mount already present in fstab."
fi

# Re-apply ownership after mount
chown -R "$TARGET_USER:$TARGET_USER" "$DEFAULT_MOUNT"

# --------------------------------------------------
# INSTALL BINARY
# --------------------------------------------------

echo "Installing storage-agent..."

install -m 755 "$BINARY_SOURCE" "$BINARY_DEST"

# --------------------------------------------------
# CREATE SYSTEMD SERVICE
# --------------------------------------------------

echo "Creating systemd service..."

cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=HomeServer Storage Agent
After=network.target

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

echo "Enabling storage-agent..."

systemctl enable storage-agent

echo "Starting storage-agent..."

systemctl restart storage-agent

# --------------------------------------------------
# DONE
# --------------------------------------------------

echo
echo "======================================="
echo " Storage Agent Installed Successfully "
echo "======================================="
echo
echo "Binary : $BINARY_DEST"
echo "Service: storage-agent"
echo "Storage: $DEFAULT_MOUNT"
echo
echo "Useful commands:"
echo "  systemctl status storage-agent"
echo "  journalctl -u storage-agent -f"
echo