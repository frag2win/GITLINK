# Specification: GITLINK Desktop Developer Tools (VS Code Extension)

This specification outlines the architecture, layout, and implementation roadmap for the **GITLINK VS Code Extension**. The extension acts as a first-class client of the GITLINK REST and WebSocket API surfaces, containing zero business logic.

---

## 🏗️ Guiding Architectural Principle

The extension must remain thin, delegating all synchronization, merge conflict calculations, database writes, and authentication checks to the `api-server` backend:

```
+──────────────────+       REST / WebSocket       +──────────────────+
|  VS Code Client  | <──────────────────────────> |   api-server     |
| (UI, State, WS)  |                              | (Business Logic) |
+──────────────────+                              +──────────────────+
```

---

## 📂 Modular File Structure

```text
extensions/
└── vscode-gitlink/
    ├── src/
    │   ├── api/            # Axios / Fetch client wrapping REST endpoints
    │   ├── auth/           # VS Code SecretStorage integrations for JWTs
    │   ├── websocket/      # Monotonic reconnect notification stream listener
    │   ├── commands/       # Registered command palette instructions
    │   ├── treeviews/      # Sidebar TreeDataProviders (Peers, Queue, Repos)
    │   ├── webviews/       # Rich HTML displays (Conflict diagnostics, PR Review)
    │   ├── statusbar/      # Simplified unified connectivity display
    │   ├── notifications/  # Toast notifications and notification panel cache
    │   ├── models/         # TypeScript type definitions matching backend models
    │   ├── cache/          # Offline/Online state management and local cache
    │   ├── utils/          # General helper functions
    │   └── extension.ts    # Main activation/deactivation entrypoint
    ├── media/              # Icons, stylesheets, and UI assets
    ├── package.json        # Extension manifests, commands, and contributions
    └── tsconfig.json       # TypeScript compiler options
```

---

## 🎨 Feature & API Specifications

### 1. SecretStorage Authentication
- **Flow**: User executes `GITLINK: Login` -> Prompts for API endpoint URL & credentials -> Exchanges credentials for JWT -> Stores JWT securely inside VS Code's native `SecretStorage` API (`context.secrets`).
- **REST client Authorization**: Every API request retrieves the token from SecretStorage, populating the `Authorization: Bearer <token>` header dynamically.

### 2. Operations & Telemetry Status Bar
- **Simplified Status**: Shows one unified connection indicator:
  - 🟢 `GITLINK Connected` (If `/health` returns `alive`).
  - 🔴 `GITLINK Offline` (If health ping fails).
- **Click Interaction**: Clicking the status bar item opens the GITLINK Operations Dashboard panel.

### 3. Sidebar Operations TreeView
- Exposes structured tree nodes mapping system states:
  - **Repositories**: List of active repositories and local branch checkouts.
  - **Connected Peers**: Mesh connections with live latency indicators (`ms`).
  - **Sync Queue**: In-flight synchronization tasks.
  - **Dead Letter Queue (DLQ)**: Tasks with failed states. Users can right-click any DLQ item to trigger `GITLINK: Retry DLQ Task` (`POST /api/v1/sync/dlq/:id/replay`).

### 4. Custom Webviews & Editor Integration
- **Conflict Diagnostic Webview**:
  - Automatically triggered upon merge conflict detections.
  - Calls `GET /api/v1/repos/:id/conflicts/analyze` and displays the computed 3-way merge base, divergent hunks, conflict reasons, and files.
- **Pull Request Review Panels**:
  - Inline code review decoration views rendering comments directly inside standard editor panes.
  - Allows marking threads as resolved directly inside the editor layout.

---

## 🚀 Incremental Development Strategy

```
+─────────────────────────────────────────────────────────────────────────────+
|                          Extension Development Stages                       |
+─────────────────────────────────────────────────────────────────────────────+
        |                       |                         |
        v                       v                         v
+────────────────+    +──────────────────+      +──────────────────+
| Stage 1: Core  |───>| Stage 2: Ops     |─────>| Stage 3: Collab  |
| - SecretStorage|    | - Status Bar     |      | - WS Stream      |
| - REST Client  |    | - Peer TreeView  |      | - PR Webviews    |
| - Configuration|    | - DLQ Controls   |      | - Inline Comments|
+────────────────+    +──────────────────+      +──────────────────+
```

1. **Stage 1 (Core)**: Establish authentication, configuration settings, base REST client wrapper, and activation lifecycle rules.
2. **Stage 2 (Operations)**: Wire connectivity status indicators, custom sidebars, and DLQ replaying capabilities.
3. **Stage 3 (Collaboration & Diagnostics)**: Integrate live WebSocket notification streaming, inline PR comments, and visual conflict diagnostic Webviews.
