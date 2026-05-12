import AppKit
import WebKit

/// Native window hosting a WKWebView pointed at the embedded Go server. Kept
/// deliberately spartan — no custom title bar tricks or window-state
/// persistence; macOS handles all of that for free at .titled + .resizable.
final class MainWindowController: NSWindowController, WKUIDelegate, WKScriptMessageHandler {
    private let webView: WKWebView

    init(initialURL: URL) {
        let config = WKWebViewConfiguration()
        let prefs = WKPreferences()
        prefs.javaScriptCanOpenWindowsAutomatically = true
        config.preferences = prefs
        config.websiteDataStore = .nonPersistent() // each launch starts clean

        // Clipboard bridge: WKWebView's navigator.clipboard.writeText and
        // document.execCommand('copy') are both unreliable here, so the SPA
        // posts the text via window.webkit.messageHandlers.copy.postMessage
        // and we write it to NSPasteboard from Swift. The script-message
        // handler is added *after* WKWebView init below (see init body) —
        // WKWebViewConfiguration is snapshot at WebView construction time,
        // so mutating this controller before isn't enough.
        config.userContentController = WKUserContentController()

        let webView = WKWebView(frame: .zero, configuration: config)
        webView.allowsBackForwardNavigationGestures = false
        self.webView = webView

        // Size the window to fill ~90% of the current screen on first launch.
        // setFrameAutosaveName persists user resizes, so this initial sizing
        // only applies until the user moves or resizes the window once.
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
        window.contentView = webView

        super.init(window: window)
        // WKWebView's JS dialog methods (alert / confirm / prompt) are silent
        // no-ops without a uiDelegate. Wire ourselves in so the React side
        // can use window.confirm() for destructive actions (e.g. the sidebar
        // delete button) and window.alert() for error toasts.
        webView.uiDelegate = self
        // Attach the clipboard bridge to the WebView's *live* content
        // controller, not to the one we passed into config (which was
        // snapshotted on init and no longer wired to anything).
        webView.configuration.userContentController.add(self, name: "copy")
        webView.load(URLRequest(url: initialURL))
    }

    required init?(coder: NSCoder) {
        fatalError("init(coder:) not used")
    }

    deinit {
        webView.configuration.userContentController.removeScriptMessageHandler(forName: "copy")
    }

    /// Replaces the current URL — handy if we ever add a "reload server" menu item.
    func load(_ url: URL) {
        webView.load(URLRequest(url: url))
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

    // MARK: - WKScriptMessageHandler (clipboard bridge)

    func userContentController(_ userContentController: WKUserContentController,
                               didReceive message: WKScriptMessage) {
        guard message.name == "copy" else { return }
        // JS may post either a string or a number — coerce to string. Anything
        // else, just ignore.
        let text: String
        switch message.body {
        case let s as String:
            text = s
        case let n as NSNumber:
            text = n.stringValue
        default:
            NSLog("copy bridge: unexpected body type \(type(of: message.body))")
            return
        }
        let pb = NSPasteboard.general
        pb.clearContents()
        let ok = pb.setString(text, forType: .string)
        NSLog("copy bridge: wrote \(text.count) chars to pasteboard, ok=\(ok)")
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
