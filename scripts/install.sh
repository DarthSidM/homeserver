#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

ENV_DIR="/etc/homeserver"
ENV_FILE="$ENV_DIR/homeserver.env"

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root:"
    echo "sudo ./install.sh"
    exit 1
fi

echo "===================================="
echo "      HomeServer Installer"
echo "===================================="

echo
echo "Creating HomeServer environment..."

mkdir -p "$ENV_DIR"

JWT_SECRET=$(openssl rand -hex 32)

cat > "$ENV_FILE" <<EOF
PORT=8000
JWT_SECRET=$JWT_SECRET
DB_PATH=/var/lib/homeserver/homeserver.db
EOF

chmod 600 "$ENV_FILE"

echo
echo "[1/2] Installing HomeServer..."
bash "$SCRIPT_DIR/install-homeserver.sh"

echo
echo "[2/2] Installing Storage Agent..."
bash "$SCRIPT_DIR/install-storage-agent.sh"

echo
echo "===================================="
echo " Installation Complete"
echo "===================================="