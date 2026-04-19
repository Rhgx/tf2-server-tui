# TF2 Server Browser TUI

[![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/dl/)
[![TF2](https://img.shields.io/badge/Team%20Fortress%202-CF6A32?logo=steam&logoColor=white)](https://store.steampowered.com/app/440/Team_Fortress_2/)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux-lightgrey)](#build)
[![License](https://img.shields.io/badge/license-GPLv3-blue.svg)](LICENSE)

A native Go terminal UI for browsing Team Fortress 2 servers, checking live status, and launching directly into TF2 through Steam.

---

## Features

- Query TF2 servers directly over the Source server info protocol
- Keyboard-navigable server list with auto-refresh
- Launch via `steam://connect/<ip>:<port>` on Enter
- In-app settings screen for saved servers, imports, and refresh interval
- Single portable native binary with no Node.js or runtime dependency

## Install Go

See the [official install guide](https://go.dev/doc/install). Verified with Go `1.26.2`.

**Windows** ([MSI](https://go.dev/dl/go1.26.2.windows-amd64.msi) or winget):

```powershell
winget install --exact --id GoLang.Go --accept-package-agreements --accept-source-agreements
```

**Linux:** install via your distro's package manager or from [go.dev](https://go.dev/dl/).

Verify:

```bash
go version
```

## Build

Build manually from the repo root:

```bash
go mod tidy
go build -trimpath -ldflags="-s -w" -o tf2-server-tui .
```

On Windows, use `-o tf2-server-tui.exe` instead.

## Releases

GitHub Actions can build release archives for Windows and Linux automatically.

To publish a release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

Pushing a `v*` tag runs the workflow in `.github/workflows/release.yml`, builds both binaries, packages them, generates SHA-256 checksums, and uploads everything to a GitHub Release.

## License

This project is licensed under GPL v3. See `LICENSE` for the repository license reference.

## Run

```bash
./tf2-server-tui                      # default config
./tf2-server-tui -config ./servers.json
```

## Configuration

`servers.json` is local user config and is not tracked in git. The app auto-creates it on first run if missing.

The repo includes `servers.example.json` as a reference:

```json
{
  "servers": [
    {
      "ip": "141.95.110.33",
      "port": 27015,
      "label": "castaway.tf | Germany"
    },
    {
      "ip": "162.120.2.24",
      "port": 1045,
      "label": "Zesty Vanilla Server | Germany"
    }
  ],
  "refreshInterval": 60
}
```

## Controls

| Key           | Action              |
| ------------- | ------------------- |
| `Up` / `Down` | Move selection      |
| `Enter`       | Connect to selected |
| `R`           | Refresh now         |
| `E`           | Open settings       |
| `Q`           | Quit                |

## Linux Notes

- Launching uses `xdg-open`, `gio open`, or `steam` to handle `steam://connect/...`
- Requires Steam and a desktop opener installed

## Portable Distribution

Ship the binary alongside `servers.json`. End users only need Steam and TF2 installed, not the Go toolchain.
