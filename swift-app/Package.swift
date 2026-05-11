// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "OZXRoomEditor",
    platforms: [.macOS(.v12)],
    targets: [
        .executableTarget(
            name: "OZXRoomEditor",
            path: "Sources/OZXRoomEditor"
        )
    ]
)
