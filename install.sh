#!/usr/bin/env bash
set -euo pipefail

if [[ $EUID -ne 0 ]]; then
    echo "Run as root" >&2
    exit 1
fi

echo "Building nm-netbird-service..."
go build -o /usr/libexec/nm-netbird-service .

echo "Installing NM VPN plugin registration..."
install -Dm644 nm-netbird-service.name /etc/NetworkManager/VPN/nm-netbird-service.name

echo "Installing NM config (marks wt0 as unmanaged)..."
install -Dm644 nm-netbird.conf /etc/NetworkManager/conf.d/nm-netbird.conf

echo "Reloading NetworkManager..."
systemctl reload NetworkManager

echo ""
echo "Done. Add a connection with:"
echo "  nmcli con add type vpn vpn-type netbird con-name 'NetBird'"
echo ""
echo "Then toggle it from the NM applet or:"
echo "  nmcli con up NetBird"
echo "  nmcli con down NetBird"
