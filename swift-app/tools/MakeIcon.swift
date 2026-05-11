#!/usr/bin/env swift
// MakeIcon.swift — generates the 1024×1024 master PNG for the OZX Room
// Editor app icon. Pure Core Graphics, no external dependencies. The icns
// is assembled from this master by the Makefile (sips + iconutil).
//
// Usage:
//   swift MakeIcon.swift <output-png-path>
//
// The icon: deep-blue squircle, faint tile-grid hatch, white "OZX" wordmark.

import AppKit
import CoreGraphics
import Foundation

// MARK: - Configuration

let size: CGFloat = 1024
// Apple's "squircle" mask isn't exposed as a public API; a rounded rect with
// ~22.4% corner radius is a good approximation that lines up with what
// Pages/Numbers/etc. ship and is what most icon templates use.
let cornerRadius: CGFloat = size * 0.2237

// Background gradient — same hue family as the in-app project banner
// (#1976D2) but a touch deeper for icon contrast.
let gradientTop = CGColor(red: 0x15 / 255, green: 0x65 / 255, blue: 0xC0 / 255, alpha: 1)
let gradientBottom = CGColor(red: 0x0D / 255, green: 0x47 / 255, blue: 0xA1 / 255, alpha: 1)

let gridColor = CGColor(red: 1, green: 1, blue: 1, alpha: 0.07)
let gridStep: CGFloat = 64
let gridLineWidth: CGFloat = 2

let wordmarkColor = CGColor(red: 1, green: 1, blue: 1, alpha: 1)
let wordmark = "OZX"
let wordmarkFontSize: CGFloat = 360 // tuned to fit ~76% of the canvas width

// MARK: - Argument parsing

guard CommandLine.arguments.count == 2 else {
    FileHandle.standardError.write("usage: MakeIcon.swift <output-png-path>\n".data(using: .utf8)!)
    exit(2)
}
let outputPath = CommandLine.arguments[1]

// MARK: - Drawing

let width = Int(size)
let height = Int(size)
let bytesPerRow = width * 4
let colorSpace = CGColorSpaceCreateDeviceRGB()
guard let ctx = CGContext(
    data: nil,
    width: width,
    height: height,
    bitsPerComponent: 8,
    bytesPerRow: bytesPerRow,
    space: colorSpace,
    bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue
) else {
    FileHandle.standardError.write("CGContext init failed\n".data(using: .utf8)!)
    exit(1)
}

// Clip to the rounded silhouette so the transparent corners are correctly
// premultiplied — otherwise the masked-out region picks up gradient pixels.
let rect = CGRect(x: 0, y: 0, width: size, height: size)
let path = CGPath(roundedRect: rect, cornerWidth: cornerRadius, cornerHeight: cornerRadius, transform: nil)
ctx.addPath(path)
ctx.clip()

// 1. Background gradient.
let gradient = CGGradient(
    colorsSpace: colorSpace,
    colors: [gradientTop, gradientBottom] as CFArray,
    locations: [0, 1]
)!
ctx.drawLinearGradient(
    gradient,
    start: CGPoint(x: 0, y: size),
    end: CGPoint(x: 0, y: 0),
    options: []
)

// 2. Faint tile-grid hatch — nods to the room-template editor without
//    being literal. Lines are inside the clip so they fade at the corners.
ctx.setStrokeColor(gridColor)
ctx.setLineWidth(gridLineWidth)
ctx.beginPath()
var x: CGFloat = gridStep
while x < size {
    ctx.move(to: CGPoint(x: x, y: 0))
    ctx.addLine(to: CGPoint(x: x, y: size))
    x += gridStep
}
var y: CGFloat = gridStep
while y < size {
    ctx.move(to: CGPoint(x: 0, y: y))
    ctx.addLine(to: CGPoint(x: size, y: y))
    y += gridStep
}
ctx.strokePath()

// 3. "OZX" wordmark, white, centered. CoreText handles the heavy lifting:
//    we ask for a heavy/black weight from the system UI font.
let font = NSFont.systemFont(ofSize: wordmarkFontSize, weight: .heavy)
let attrs: [NSAttributedString.Key: Any] = [
    .font: font,
    .foregroundColor: NSColor(cgColor: wordmarkColor)!,
    .kern: -6 // tighten the OZX tracking so the letters feel like a wordmark
]
let attributed = NSAttributedString(string: wordmark, attributes: attrs)
let line = CTLineCreateWithAttributedString(attributed)
let bounds = CTLineGetBoundsWithOptions(line, .useOpticalBounds)

// Center the optical bounding box of the wordmark in the canvas.
let textX = (size - bounds.width) / 2 - bounds.origin.x
let textY = (size - bounds.height) / 2 - bounds.origin.y
ctx.textPosition = CGPoint(x: textX, y: textY)
CTLineDraw(line, ctx)

// MARK: - Write PNG

guard let cgImage = ctx.makeImage() else {
    FileHandle.standardError.write("makeImage failed\n".data(using: .utf8)!)
    exit(1)
}
let bitmap = NSBitmapImageRep(cgImage: cgImage)
guard let png = bitmap.representation(using: .png, properties: [:]) else {
    FileHandle.standardError.write("PNG encode failed\n".data(using: .utf8)!)
    exit(1)
}
let outputURL = URL(fileURLWithPath: outputPath)
do {
    try png.write(to: outputURL)
} catch {
    FileHandle.standardError.write("write \(outputPath): \(error)\n".data(using: .utf8)!)
    exit(1)
}
print("Wrote \(outputPath) (\(png.count) bytes)")
