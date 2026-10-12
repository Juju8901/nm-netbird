# nm-netbird

> ⚠️ **Vibecoded.** This was written in one shot with Claude. It works (probably), but treat it accordingly — test before deploying, read the code before trusting it.

A NetworkManager VPN plugin that wraps [NetBird](https://github.com/netbirdio/netbird), letting you connect/disconnect from the GNOME network applet instead of a separate desktop UI.

## How it works

NM launches `nm-netbird-service` as a subprocess when you activate the VPN connection. The service implements the `org.freedesktop.NetworkManager.VPN.Plugin` D-Bus interface and translates NM's connect/disconnect calls into `netbird up` / `netbird down`.

**Auth flow**: if `netbird up` fails with an auth error (expired or missing token), the plugin automatically runs `netbird login`, which opens your browser for the OAuth flow. Once you authenticate, it retries `netbird up` and finishes connecting. NM shows "Waiting for browser authentication…" in the meantime.

```
Connect clicked
  └─ netbird up
     ├─ ok  → connected
     └─ auth error
           └─ netbird login  (browser OAuth)
                └─ netbird up  → connected (or failure)
```

## Requirements

- `netbird` CLI installed and daemon running (`systemctl enable --now netbird`)
- Go 1.21+ (build-time only)
- NetworkManager

## Install

```bash
git clone https://github.com/Juju8901/nm-netbird
cd nm-netbird
sudo ./install.sh
```

Then add the connection once:

```bash
nmcli con add type vpn vpn-type netbird con-name 'NetBird'
```

After that, the VPN appears in the GNOME network applet like any other VPN.

## Uninstall

```bash
sudo rm /usr/libexec/nm-netbird-service \
        /etc/NetworkManager/VPN/nm-netbird-service.name \
        /etc/NetworkManager/conf.d/nm-netbird.conf
sudo systemctl reload NetworkManager
nmcli con delete NetBird
```

## Caveats

- NetBird's `wt0` interface is marked unmanaged in NM (via `nm-netbird.conf`) to prevent NM from conflicting with NetBird's own interface management. The VPN toggle still works; NM just doesn't configure addresses/routes on `wt0` itself.
- If your distro uses a different interface name than `wt0`, update `nm-netbird.conf` and the `TunIface` default in `netbird.go`.
- Auth error detection is heuristic (checks stderr for keywords like "unauthenticated"). If `netbird up`'s error output changes in a future NetBird release, update `isAuthError` in `netbird.go`.
