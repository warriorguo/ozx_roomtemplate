# OZX Room Editor — native macOS app

Self-contained `.app` bundle that wraps the `ozx-roomeditor` Go server in a
native Cocoa window via WKWebView. Same editor as `cd tile-backend &&
make build-local && ./bin/ozx-roomeditor`, but launched like a normal Mac
app — no terminal, no separate browser window, with `⌘Q` for graceful
shutdown.

Follows the pattern used by the `videoz` project.

## Build

```bash
cd swift-app
make build              # → build/OZX Room Editor.app   (~8 MB)
make run                # build + open the app
```

`make build` runs `tile-backend/make build-local` first to produce the
embedded Go binary, then `swift build -c release`, then assembles the
`.app` layout below.

## Bundle layout

```
OZX Room Editor.app/
└── Contents/
    ├── Info.plist                     # CFBundle*, ATS localhost exception
    ├── MacOS/
    │   └── OZXRoomEditor              # Swift launcher
    └── Resources/
        └── ozx-roomeditor             # Go server (go:embed SPA inside)
```

## How it works

On launch the Swift app:

1. Allocates a free TCP port (`bind(:0)` then close).
2. Spawns `Resources/ozx-roomeditor` with `--port <n> --no-browser` and
   redirects its stdio to `~/Library/Logs/ozx-roomeditor.log`.
3. Polls `http://localhost:<n>/health` until 200 OK (≤5 s timeout).
4. Shows an `NSWindow` containing a `WKWebView` loading
   `http://localhost:<n>/`.

`applicationWillTerminate` sends `SIGTERM` to the child and waits up to
2 s, then falls back to `SIGKILL` if it's still alive.
`applicationShouldTerminateAfterLastWindowClosed = true`, so closing the
window quits the app.

The user config (`~/.config/ozx-roomeditor/config.json`) is shared with
the standalone Go binary — set `project_root` there or use the in-window
project banner.

## Layout

```
swift-app/
├── Package.swift                       # SPM (no Xcode required)
├── Makefile                            # build / run / clean
├── Resources/
│   └── Info.plist
└── Sources/OZXRoomEditor/
    ├── main.swift                      # NSApplication entry point
    ├── AppDelegate.swift               # lifecycle + menu bar
    ├── MainWindow.swift                # NSWindow + WKWebView
    └── GoServer.swift                  # subprocess manager
```

## Out of scope

These are intentionally deferred — open follow-up tickets when ready:

- Code signing / notarization (currently launches with Gatekeeper warning).
- DMG packaging.
- Custom app icon.
- Sparkle / auto-update integration.

## Dev iteration without rebundling

```bash
# Run a normal Go server in one shell:
cd tile-backend && go run cmd/ozx-roomeditor/main.go --port 9999 --no-browser

# Point the dev WKWebView at it (skip GoServer subprocess management):
# … just edit GoServer.locateBinary() temporarily, or hardcode `url` in
# AppDelegate.applicationDidFinishLaunching.
```

For the common case of "did my Swift change compile", `swift build` is
fine without rebundling — but you'll need a recent
`tile-backend/bin/ozx-roomeditor` so the dev-fallback path in
`GoServer.locateBinary()` resolves.
