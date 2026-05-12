import AppKit
import WebKit

/// Native window hosting a WKWebView pointed at the embedded Go server. Kept
/// deliberately spartan — no custom title bar tricks or window-state
/// persistence; macOS handles all of that for free at .titled + .resizable.
final class MainWindowController: NSWindowController, WKUIDelegate, WKScriptMessageHandler {
    private var webView: WKWebView!

    init(initialURL: URL) {
        // Build the window first with an empty placeholder so we can call
        // super.init and gain `self`. The WebView is constructed *after*
        // super.init so we can register script message handlers (which need
        // `self`) before the WKWebView snapshots its configuration.
        let visibleFrame = NSScreen.main?.visibleFrame ?? NSRect(x: 0, y: 0, width: 1600, height: 1000)
        let width = max(1280, visibleFrame.width * 0.9)
        let height = max(800, visibleFrame.height * 0.9)

        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: width, height: height),
            styleMask: [.titled, .closable, .miniaturizable, .resizable],
            backing: .buffered,
            defer: false)
        window.title = "OZX Room Editor"
        window.minSize = NSSize(width: 1024, height: 700)
        window.center()
        // Bumping the autosave-name suffix discards any saved frame from a
        // previous version so this enlarged default actually applies on the
        // next launch; user-initiated resizes from here on persist normally.
        window.setFrameAutosaveName("MainWindow.v2")

        super.init(window: window)

        // Now build the WebView. Script-message handlers must be on the
        // controller *before* WKWebView is constructed: the configuration is
        // copied at init time, so post-creation mutations don't reach the
        // running WebView.
        let config = WKWebViewConfiguration()
        let prefs = WKPreferences()
        prefs.javaScriptCanOpenWindowsAutomatically = true
        config.preferences = prefs
        config.websiteDataStore = .nonPersistent() // each launch starts clean

        let contentController = WKUserContentController()
        contentController.add(self, name: "copy")
        contentController.add(self, name: "openWith")
        config.userContentController = contentController

        let webView = WKWebView(frame: .zero, configuration: config)
        webView.allowsBackForwardNavigationGestures = false
        webView.uiDelegate = self
        self.webView = webView
        window.contentView = webView
        webView.load(URLRequest(url: initialURL))
    }

    required init?(coder: NSCoder) {
        fatalError("init(coder:) not used")
    }

    deinit {
        let ucc = webView.configuration.userContentController
        ucc.removeScriptMessageHandler(forName: "copy")
        ucc.removeScriptMessageHandler(forName: "openWith")
    }

    /// Replaces the current URL — handy if we ever add a "reload server" menu item.
    func load(_ url: URL) {
        webView.load(URLRequest(url: url))
    }

    // MARK: - WKScriptMessageHandler (clipboard bridge)

    func userContentController(_ userContentController: WKUserContentController,
                               didReceive message: WKScriptMessage) {
        switch message.name {
        case "copy":
            handleCopy(message.body)
        case "openWith":
            handleOpenWith(message.body)
        default:
            NSLog("unknown script message: \(message.name)")
        }
    }

    private func handleCopy(_ body: Any) {
        let text: String
        switch body {
        case let s as String: text = s
        case let n as NSNumber: text = n.stringValue
        default:
            NSLog("copy bridge: unexpected body type \(type(of: body))")
            return
        }
        let pb = NSPasteboard.general
        pb.clearContents()
        let ok = pb.setString(text, forType: .string)
        NSLog("copy bridge: wrote \(text.count) chars to pasteboard, ok=\(ok)")
    }

    /// Expects `{ "app": "<absolute path to .app>", "args": ["--room", "/foo.json"] }`.
    /// Uses NSWorkspace.openApplication so the .app bundle is launched the same
    /// way Finder / `open -a` would, with the supplied arguments forwarded to
    /// the executable.
    private func handleOpenWith(_ body: Any) {
        guard let dict = body as? [String: Any],
              let appPath = dict["app"] as? String,
              !appPath.isEmpty else {
            NSLog("openWith bridge: missing or invalid 'app' field: \(body)")
            return
        }
        let args = (dict["args"] as? [String]) ?? []
        let appURL = URL(fileURLWithPath: appPath)
        let configuration = NSWorkspace.OpenConfiguration()
        configuration.arguments = args
        configuration.activates = true

        NSWorkspace.shared.openApplication(at: appURL, configuration: configuration) { _, err in
            if let err = err {
                NSLog("openWith bridge: launch failed (\(appPath) \(args)): \(err.localizedDescription)")
            } else {
                NSLog("openWith bridge: launched \(appPath) with args=\(args)")
            }
        }
    }

    // MARK: - WKUIDelegate

    func webView(_ webView: WKWebView,
                 runJavaScriptAlertPanelWithMessage message: String,
                 initiatedByFrame frame: WKFrameInfo,
                 completionHandler: @escaping () -> Void) {
        let alert = NSAlert()
        alert.messageText = "OZX Room Editor"
        alert.informativeText = message
        alert.alertStyle = .informational
        alert.addButton(withTitle: "OK")
        alert.beginSheetModal(for: window ?? NSApp.keyWindow ?? NSWindow()) { _ in
            completionHandler()
        }
    }

    func webView(_ webView: WKWebView,
                 runJavaScriptConfirmPanelWithMessage message: String,
                 initiatedByFrame frame: WKFrameInfo,
                 completionHandler: @escaping (Bool) -> Void) {
        let alert = NSAlert()
        alert.messageText = "OZX Room Editor"
        alert.informativeText = message
        alert.alertStyle = .warning
        alert.addButton(withTitle: "OK")
        alert.addButton(withTitle: "Cancel")
        alert.beginSheetModal(for: window ?? NSApp.keyWindow ?? NSWindow()) { response in
            completionHandler(response == .alertFirstButtonReturn)
        }
    }

    func webView(_ webView: WKWebView,
                 runJavaScriptTextInputPanelWithPrompt prompt: String,
                 defaultText: String?,
                 initiatedByFrame frame: WKFrameInfo,
                 completionHandler: @escaping (String?) -> Void) {
        let alert = NSAlert()
        alert.messageText = "OZX Room Editor"
        alert.informativeText = prompt
        alert.alertStyle = .informational
        alert.addButton(withTitle: "OK")
        alert.addButton(withTitle: "Cancel")
        let input = NSTextField(frame: NSRect(x: 0, y: 0, width: 320, height: 24))
        input.stringValue = defaultText ?? ""
        alert.accessoryView = input
        alert.beginSheetModal(for: window ?? NSApp.keyWindow ?? NSWindow()) { response in
            if response == .alertFirstButtonReturn {
                completionHandler(input.stringValue)
            } else {
                completionHandler(nil)
            }
        }
    }
}
