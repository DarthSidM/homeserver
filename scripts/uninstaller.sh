#!/bin/bash

set -e

echo "======================================="
echo "      HomeServer Uninstaller"
echo "======================================="

# --------------------------------------------------
# CONFIG
# --------------------------------------------------

HOME_SERVER_SERVICE="homeserver"
STORAGE_AGENT_SERVICE="storage-agent"

HOME_SERVER_BINARY="/usr/local/bin/homeserver"
STORAGE_AGENT_BINARY="/usr/local/bin/storage-agent"

HOME_SERVER_SERVICE_FILE="/etc/systemd/system/homeserver.service"
STORAGE_AGENT_SERVICE_FILE="/etc/systemd/system/storage-agent.service"

ENV_DIR="/etc/homeserver"
DATA_DIR="/var/lib/homeserver"

STORAGE_ROOT="/storage"
DEFAULT_DATA_DIR="/var/lib/homeserver-data"
DEFAULT_MOUNT="/storage/default"

# --------------------------------------------------
# ROOT CHECK
# --------------------------------------------------

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root:"
    echo "sudo ./uninstall.sh"
    exit 1
fi

# --------------------------------------------------
# CONFIRMATION
# --------------------------------------------------

echo
echo "WARNING:"
echo "This will permanently remove:"
echo
echo "  - HomeServer"
echo "  - Storage Agent"
echo "  - Database"
echo "  - Environment configuration"
echo "  - Storage mounts"
echo "  - Stored files"
echo
read -p "Continue? (yes/no): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    echo "Aborted."
    exit 0
fi

# --------------------------------------------------
# STOP SERVICES
# --------------------------------------------------

echo
echo "Stopping services..."

systemctl stop "$HOME_SERVER_SERVICE" 2>/dev/null || true
systemctl stop "$STORAGE_AGENT_SERVICE" 2>/dev/null || true

systemctl disable "$HOME_SERVER_SERVICE" 2>/dev/null || true
systemctl disable "$STORAGE_AGENT_SERVICE" 2>/dev/null || true

# --------------------------------------------------
# REMOVE SERVICE FILES
# --------------------------------------------------

echo
echo "Removing systemd services..."

rm -f "$HOME_SERVER_SERVICE_FILE"
rm -f "$STORAGE_AGENT_SERVICE_FILE"

systemctl daemon-reload

# --------------------------------------------------
# REMOVE BINARIES
# --------------------------------------------------

echo
echo "Removing binaries..."

rm -f "$HOME_SERVER_BINARY"
rm -f "$STORAGE_AGENT_BINARY"

# --------------------------------------------------
# REMOVE BIND MOUNT
# --------------------------------------------------

echo
echo "Removing storage mount..."

if mountpoint -q "$DEFAULT_MOUNT"; then
    umount "$DEFAULT_MOUNT"
fi

# --------------------------------------------------
# REMOVE FSTAB ENTRY
# --------------------------------------------------

echo
echo "Cleaning /etc/fstab..."

sed -i "\|$DEFAULT_MOUNT|d" /etc/fstab

# --------------------------------------------------
# REMOVE DATA
# --------------------------------------------------

echo
echo "Removing data..."

rm -rf "$DATA_DIR"
rm -rf "$DEFAULT_DATA_DIR"
rm -rf "$STORAGE_ROOT"

# --------------------------------------------------
# REMOVE CONFIGURATION
# --------------------------------------------------

echo
echo "Removing configuration..."

rm -rf "$ENV_DIR"

# --------------------------------------------------
# FINAL CLEANUP
# --------------------------------------------------

systemctl daemon-reload

echo
echo "======================================="
echo " HomeServer Uninstalled Successfully "
echo "======================================="
echo