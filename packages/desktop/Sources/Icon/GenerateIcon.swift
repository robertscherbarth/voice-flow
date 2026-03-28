import AppKit
import SwiftUI

// Core logic to export the App Icon
let iconSize = CGSize(width: 1024, height: 1024)

// Gradient
let gradient = NSGradient(colors: [
    NSColor(red: 0.20, green: 0.44, blue: 0.45, alpha: 1.0), // Teal
    NSColor(red: 0.18, green: 0.29, blue: 0.45, alpha: 1.0), // Center
    NSColor(red: 0.43, green: 0.24, blue: 0.55, alpha: 1.0)  // Purple
])!

let image = NSImage(size: iconSize)
image.lockFocus()

let context = NSGraphicsContext.current!.cgContext
let bounds = NSRect(origin: .zero, size: iconSize)

// Draw Background Gradient
gradient.draw(in: bounds, angle: -45)

// Draw Wave Path
context.setStrokeColor(NSColor.white.cgColor)
context.setLineWidth(24)
context.setLineCap(.round)
context.setLineJoin(.round)

let path = NSBezierPath()
let midY = iconSize.height / 2
let width = iconSize.width

path.move(to: NSPoint(x: 0, y: midY))
path.curve(to: NSPoint(x: width * 0.25, y: midY - 60),
           controlPoint1: NSPoint(x: width * 0.1, y: midY),
           controlPoint2: NSPoint(x: width * 0.15, y: midY - 140))

path.curve(to: NSPoint(x: width * 0.35, y: midY + 120),
           controlPoint1: NSPoint(x: width * 0.3, y: midY + 40),
           controlPoint2: NSPoint(x: width * 0.3, y: midY + 140))

path.curve(to: NSPoint(x: width * 0.65, y: midY - 100),
           controlPoint1: NSPoint(x: width * 0.45, y: midY + 80),
           controlPoint2: NSPoint(x: width * 0.55, y: midY - 180))

path.curve(to: NSPoint(x: width * 0.85, y: midY + 60),
           controlPoint1: NSPoint(x: width * 0.72, y: midY - 40),
           controlPoint2: NSPoint(x: width * 0.78, y: midY + 140))

path.curve(to: NSPoint(x: width, y: midY),
           controlPoint1: NSPoint(x: width * 0.9, y: midY),
           controlPoint2: NSPoint(x: width * 0.95, y: midY))

path.stroke()

// To draw the SF Symbol (microphone), we need to load an NSImage
if let micImage = NSImage(systemSymbolName: "mic", accessibilityDescription: nil) {
    let micConfig = NSImage.SymbolConfiguration(pointSize: 420, weight: .medium)
    let configuredMic = micImage.withSymbolConfiguration(micConfig)!
    configuredMic.isTemplate = true
    
    // Create a rect in the center
    let imageRect = NSRect(x: (iconSize.width - 420) / 2, y: (iconSize.height - 560) / 2, width: 420, height: 560)
    
    // Draw background drop shadow behind the mic
    let shadow = NSShadow()
    shadow.shadowColor = NSColor.black.withAlphaComponent(0.4)
    shadow.shadowOffset = NSSize(width: 0, height: -10)
    shadow.shadowBlurRadius = 10
    shadow.set()
    
    NSColor.white.set()
    configuredMic.draw(in: imageRect)
}

image.unlockFocus()

// Save to disk
if let tiffData = image.tiffRepresentation,
   let bitmapImage = NSBitmapImageRep(data: tiffData),
   let pngData = bitmapImage.representation(using: .png, properties: [:]) {
    let url = URL(fileURLWithPath: "AppIcon.png")
    try pngData.write(to: url)
    print("Successfully generated \(url.path)")
}