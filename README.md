# Firstboot

Firstboot is a cross-platform, terminal UI for turning a fresh operating system into a practical development machine. It inventories the machine, presents only missing tools, lets you select them with checkboxes, and installs through the native package manager with a visible plan.

It supports Linux, macOS, and Windows. The project is intentionally inspired by the useful workflow of community post-install tools, but it does not copy third-party code or make destructive system changes by default.

## Quick install

Linux/macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/aethelreborn/firstboot/main/scripts/install.sh | sh
```

Windows PowerShell:

```powershell
iex (irm https://raw.githubusercontent.com/aethelreborn/firstboot/main/scripts/install.ps1)
```

The scripts download the latest release for the detected architecture. Until the first release is published, build from source with `go run ./cmd/firstboot`.

## TUI controls

- `↑`/`↓` or `j`/`k`: move
- `space`: toggle a package
- `a`: select all missing packages
- `n`: clear the selection
- `enter`: review and install selected packages
- `q`/`ctrl+c`: quit

The installer asks for elevated privileges only when the selected native package-manager command requires it. It never accepts or stores a password.

## Scope

The initial catalog covers Git, Docker, Node.js, Python, JDK, Rust, Go, and optional AI developer CLIs including Claude Code, OpenCode, Hermes, and OpenClaw. Package availability is detected per OS and package manager. Unsupported entries are shown as unavailable instead of being silently skipped.

The tool does not disable animations, modify shell profiles, or remove files automatically in the initial release. Those actions need explicit, reviewed recipes before they belong in a safe bootstrapper.

## Development

```sh
go test ./...
go run ./cmd/firstboot
```

Releases are built by GitHub Actions for Linux, macOS, and Windows on amd64 and arm64. See [CONTRIBUTING.md](CONTRIBUTING.md) for the release and contribution conventions.
