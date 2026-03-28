import AppKit

let imagePath = "AppIcon_Rounded.png"
guard let image = NSImage(contentsOfFile: imagePath) else {
    print("Failed to load image at \(imagePath)")
    exit(1)
}

let pasteboard = NSPasteboard.general
pasteboard.clearContents()
let success = pasteboard.writeObjects([image])

if success {
    print("Successfully copied \(imagePath) to the clipboard!")
} else {
    print("Failed to copy to clipboard.")
    exit(1)
}