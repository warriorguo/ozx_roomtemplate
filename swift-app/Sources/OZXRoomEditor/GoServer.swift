import Foundation

/// Spawns and supervises the bundled `ozx-roomeditor` Go binary as a child
/// process. The app picks a free TCP port at launch so the server never
/// collides with whatever the user has set in `~/.config/ozx-roomeditor/config.json`.
///
/// stdout/stderr from the Go process are appended to
/// `~/Library/Logs/ozx-roomeditor.log` — useful when something goes wrong but
/// kept out of NSLog so the macOS Console doesn't fill up.
final class GoServer {
    enum StartError: Error, LocalizedError {
        case binaryNotFound
        case portAllocationFailed
        case healthCheckTimeout
        case process(Error)

        var errorDescription: String? {
            switch self {
            case .binaryNotFound: return "Bundled ozx-roomeditor binary not found in app resources."
            case .portAllocationFailed: return "Could not allocate a local port for the editor."
            case .healthCheckTimeout: return "Editor server did not become ready in time."
            case .process(let err): return "Failed to launch editor server: \(err.localizedDescription)"
            }
        }
    }

    /// URL the WebView should load once `start()` resolves successfully.
    private(set) var url: URL?

    private var process: Process?
    private var logHandle: FileHandle?

    /// Launches the server. Throws on failure to find the binary, allocate a
    /// port, or pass a /health check within the timeout.
    func start() throws {
        guard let binaryURL = Self.locateBinary() else {
            throw StartError.binaryNotFound
        }
        let port = try Self.allocateFreePort()

        // Open a log file for the child process. Append mode so prior runs
        // remain readable when debugging.
        let logURL = Self.logURL()
        try? FileManager.default.createDirectory(
            at: logURL.deletingLastPathComponent(),
            withIntermediateDirectories: true)
        if !FileManager.default.fileExists(atPath: logURL.path) {
            FileManager.default.createFile(atPath: logURL.path, contents: nil)
        }
        let handle = try FileHandle(forWritingTo: logURL)
        handle.seekToEndOfFile()
        let header = "\n--- ozx-roomeditor launch \(Date()) port=\(port) ---\n"
        if let data = header.data(using: .utf8) { handle.write(data) }
        self.logHandle = handle

        let p = Process()
        p.executableURL = binaryURL
        // --no-browser keeps the embedded server from also launching Safari
        // alongside the WKWebView we're about to point at it.
        p.arguments = ["--port", String(port), "--no-browser"]
        p.standardOutput = handle
        p.standardError = handle
        do {
            try p.run()
        } catch {
            throw StartError.process(error)
        }
        self.process = p
        let url = URL(string: "http://localhost:\(port)/")!
        try waitForReady(url: url, timeout: 5.0)
        self.url = url
    }

    /// Sends SIGTERM and waits briefly; falls back to SIGKILL if still alive.
    /// Safe to call multiple times.
    func stop() {
        guard let p = process, p.isRunning else { return }
        p.terminate()
        let deadline = Date().addingTimeInterval(2.0)
        while p.isRunning && Date() < deadline {
            Thread.sleep(forTimeInterval: 0.05)
        }
        if p.isRunning {
            kill(p.processIdentifier, SIGKILL)
        }
        try? logHandle?.close()
        logHandle = nil
    }

    // MARK: - Helpers

    /// Returns the bundled binary's URL — looks first inside the .app's
    /// Resources directory (production), then falls back to a sibling path
    /// when running via `swift run` during development.
    private static func locateBinary() -> URL? {
        if let resource = Bundle.main.url(forResource: "ozx-roomeditor", withExtension: nil) {
            return resource
        }
        // Dev fallback: ../tile-backend/bin/ozx-roomeditor relative to the
        // running executable, useful for `swift run` outside the .app bundle.
        let exe = Bundle.main.executableURL ?? URL(fileURLWithPath: CommandLine.arguments[0])
        let candidates = [
            exe.deletingLastPathComponent()
                .appendingPathComponent("../../../../tile-backend/bin/ozx-roomeditor")
                .standardizedFileURL,
            URL(fileURLWithPath: FileManager.default.currentDirectoryPath)
                .appendingPathComponent("tile-backend/bin/ozx-roomeditor"),
            URL(fileURLWithPath: FileManager.default.currentDirectoryPath)
                .appendingPathComponent("../tile-backend/bin/ozx-roomeditor"),
        ]
        for c in candidates where FileManager.default.isExecutableFile(atPath: c.path) {
            return c
        }
        return nil
    }

    /// Asks the kernel for a free TCP port by binding to :0 then closing.
    /// There's a small race where the port can be reused before the child
    /// claims it, but in practice that doesn't happen on a single-user
    /// desktop within microseconds.
    private static func allocateFreePort() throws -> Int {
        let fd = socket(AF_INET, SOCK_STREAM, 0)
        guard fd >= 0 else { throw StartError.portAllocationFailed }
        defer { close(fd) }

        var enable: Int32 = 1
        setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &enable, socklen_t(MemoryLayout<Int32>.size))

        var addr = sockaddr_in()
        addr.sin_family = sa_family_t(AF_INET)
        addr.sin_addr.s_addr = INADDR_ANY.bigEndian
        addr.sin_port = 0
        let size = socklen_t(MemoryLayout<sockaddr_in>.size)

        let bindResult = withUnsafePointer(to: &addr) { ptr -> Int32 in
            ptr.withMemoryRebound(to: sockaddr.self, capacity: 1) { sa in
                bind(fd, sa, size)
            }
        }
        guard bindResult == 0 else { throw StartError.portAllocationFailed }

        var bound = sockaddr_in()
        var bsize = size
        let getResult = withUnsafeMutablePointer(to: &bound) { ptr -> Int32 in
            ptr.withMemoryRebound(to: sockaddr.self, capacity: 1) { sa in
                getsockname(fd, sa, &bsize)
            }
        }
        guard getResult == 0 else { throw StartError.portAllocationFailed }

        return Int(UInt16(bigEndian: bound.sin_port))
    }

    /// Polls `<url>/health` until 200 OK or the deadline elapses.
    private func waitForReady(url: URL, timeout: TimeInterval) throws {
        let healthURL = url.appendingPathComponent("health")
        let deadline = Date().addingTimeInterval(timeout)
        var lastErr: Error?
        while Date() < deadline {
            let sem = DispatchSemaphore(value: 0)
            var ok = false
            let task = URLSession.shared.dataTask(with: healthURL) { _, response, err in
                if let http = response as? HTTPURLResponse, http.statusCode == 200 {
                    ok = true
                } else if let err = err {
                    lastErr = err
                }
                sem.signal()
            }
            task.resume()
            _ = sem.wait(timeout: .now() + 0.5)
            if ok { return }
            Thread.sleep(forTimeInterval: 0.1)
        }
        if let lastErr = lastErr {
            NSLog("Editor /health probe last error: \(lastErr.localizedDescription)")
        }
        throw StartError.healthCheckTimeout
    }

    private static func logURL() -> URL {
        let logs = FileManager.default
            .urls(for: .libraryDirectory, in: .userDomainMask).first?
            .appendingPathComponent("Logs")
            ?? URL(fileURLWithPath: NSTemporaryDirectory())
        return logs.appendingPathComponent("ozx-roomeditor.log")
    }
}
