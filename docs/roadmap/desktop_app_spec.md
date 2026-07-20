# Specification: GITLINK Standalone Desktop Application

This specification outlines the architecture, packaging configuration, and layout for a native **GITLINK Standalone Desktop Application** suitable for local installation.

---

## 🏗️ Architectural Choice: Go-Native Wails Core

Instead of a resource-heavy Electron wrapper, GITLINK uses **Wails** (Go + Native WebViews) to bundle the Go backend services, web interface, and the Rust Git server daemon sidecar into a single, unified desktop installer.

```
+─────────────────────────────────────────────────────────────+
|                     GITLINK Desktop App                     |
+─────────────────────────────────────────────────────────────+
|   Web UI Frontend (HTML5/React/CSS - Acrylic blur window)   |
+─────────────────────────────────────────────────────────────+
                              │
                              ▼ Wails IPC Bindings
+─────────────────────────────────────────────────────────────+
|                    Go-Native Controller                     |
|  - Spawns Rust git-server.exe as sidecar daemon            |
|  - Mounts Go api-server endpoints locally                   |
|  - Manages SQLite Database & local filesystem paths         |
+─────────────────────────────────────────────────────────────+
```

---

## 📦 Sidecar Executable Architecture

To achieve a single-click installable app, the installer bundles the compiled binaries of the subsystems:

1. **Wails Binary (`gitlink.exe`)**: Spawns the web layout, handles OS window controls, and configures system tray menus.
2. **Rust Git Server Sidecar (`git-server.exe`)**: Bundled as a sidecar resource. On app boot, `gitlink.exe` spawns the Rust daemon on a dynamically assigned local TCP port or domain socket.
3. **P2P Client Manager (`libp2p-node`)**: Embedded as a Go routine directly within the Wails application binary.

---

## 🎨 Visual Design & Window Aesthetics
* **Premium Glassmorphism**: Utilizes Wails' frameless window configuration with **Mica/Acrylic blur** window backdrops matching user OS themes.
* **Frameless Layout**: Native title bars are hidden. Window drag zones and minimize/maximize/close buttons are customized directly within the web layout.
* **System Tray integration**: Right-clicking the system tray icon displays synchronization metrics, active peer count, and toggles background worker states.

---

## ⚙️ Desktop Configuration Panel

The application provides a desktop-native Settings panel:
* **Local Repo Path**: Set the base folder directory where P2P repositories are initialized on the hard drive (e.g. `C:\Users\Username\GitlinkRepos`).
* **P2P Listening Port**: Configure the port utilized by the local libp2p DHT node (default: `4001`).
* **Start on Boot**: Regulates startup configurations via OS registries (Windows Run key, macOS LaunchAgents).

---

## 🚀 Packaging & Installers (NSIS / DMG)

The production packaging pipeline builds installer binaries utilizing standard packaging tools:
- **Windows**: Compiles into a single-file executable using **NSIS (Nullsoft Scriptable Install System)** to create a standard Windows installer (`gitlink-setup.exe`).
- **macOS**: Compiles into a signed, notarized App Bundle packaged inside a **DMG** file (`gitlink.dmg`).
