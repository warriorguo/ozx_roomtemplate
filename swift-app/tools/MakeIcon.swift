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
let wordmarkFontSize: CGFloat = 320 // tuned so the OZX+subtitle stack fits centred

let subtitle = "room-tpl"
let subtitleFontSize: CGFloat = 130
let subtitleColor = CGColor(red: 1, green: 1, blue: 1, alpha: 0.78) // softer than the main wordmark
let stackGap: CGFloat = 24 // visual gap between OZX baseline and subtitle cap-height

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

// 3. OZX wordmark + "room-tpl" subtitle, stacked and centred. We measure
//    both first so we can centre the combined block vertically — measuring
//    keeps the visual centre aligned even when the relative font sizes
//    change later.
func makeLine(_ text: String, size px: CGFloat, weight: NSFont.Weight, color: CGColor, kern: CGFloat) -> (CTLine, CGRect) {
    let f = NSFont.systemFont(ofSize: px, weight: weight)
    let a: [NSAttributedString.Key: Any] = [
        .font: f,
        .foregroundColor: NSColor(cgColor: color)!,
        .kern: kern
    ]
    let line = CTLineCreateWithAttributedString(NSAttributedString(string: text, attributes: a))
    return (line, CTLineGetBoundsWithOptions(line, .useOpticalBounds))
}

let (wordLine, wordBounds) = makeLine(wordmark, size: wordmarkFontSize, weight: .heavy, color: wordmarkColor, kern: -6)
let (subLine,  subBounds)  = makeLine(subtitle, size: subtitleFontSize, weight: .semibold, color: subtitleColor, kern: -1)

// Total stack height = OZX height + gap + subtitle height. We compose in
// Core Graphics (origin bottom-left), so the subtitle baseline ends up
// below the OZX baseline.
let stackHeight = wordBounds.height + stackGap + subBounds.height
let stackBottom = (size - stackHeight) / 2

// Subtitle below.
let subX = (size - subBounds.width) / 2 - subBounds.origin.x
let subY = stackBottom - subBounds.origin.y
ctx.textPosition = CGPoint(x: subX, y: subY)
CTLineDraw(subLine, ctx)

// OZX above the subtitle (offset by subtitle height + gap).
let wordX = (size - wordBounds.width) / 2 - wordBounds.origin.x
let wordY = stackBottom + subBounds.height + stackGap - wordBounds.origin.y
ctx.textPosition = CGPoint(x: wordX, y: wordY)
CTLineDraw(wordLine, ctx)

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
