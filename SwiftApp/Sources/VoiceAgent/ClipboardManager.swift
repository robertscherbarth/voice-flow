import Cocoa
import ApplicationServices

class ClipboardManager {
    static func paste(text: String) {
        let pasteboard = NSPasteboard.general
        pasteboard.clearContents()
        pasteboard.setString(text, forType: .string)
        
        // Wait for the pasteboard to sync and the target app to be ready
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.2) {
            if AXIsProcessTrusted() {
                simulateCmdV()
            } else {
                print("Cannot paste: Accessibility permission not granted.")
                // Request again just in case
                let options = [kAXTrustedCheckOptionPrompt.takeUnretainedValue() as String: true] as CFDictionary
                AXIsProcessTrustedWithOptions(options)
            }
        }
    }
    
    static private func simulateCmdV() {
        let src = CGEventSource(stateID: .hidSystemState)
        
        let vDown = CGEvent(keyboardEventSource: src, virtualKey: 0x09, keyDown: true)
        let vUp = CGEvent(keyboardEventSource: src, virtualKey: 0x09, keyDown: false)
        
        // Apply the Command modifier flag
        vDown?.flags = .maskCommand
        vUp?.flags = .maskCommand
        
        // Post the events
        let loc = CGEventTapLocation.cghidEventTap
        vDown?.post(tap: loc)
        vUp?.post(tap: loc)
        
        print("Pasted via CGEvent")
    }
}
