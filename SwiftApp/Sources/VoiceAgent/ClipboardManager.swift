import Cocoa

class ClipboardManager {
    static func paste(text: String) {
        let pasteboard = NSPasteboard.general
        pasteboard.clearContents()
        pasteboard.setString(text, forType: .string)
        
        // Give the pasteboard a tiny moment to register the change
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
            simulateCmdV()
        }
    }
    
    static private func simulateCmdV() {
        let src = CGEventSource(stateID: .hidSystemState)
        
        // 9 is the virtual key code for 'v'
        let cmdVDown = CGEvent(keyboardEventSource: src, virtualKey: 0x09, keyDown: true)
        let cmdVUp = CGEvent(keyboardEventSource: src, virtualKey: 0x09, keyDown: false)
        
        let flags = CGEventFlags.maskCommand
        cmdVDown?.flags = flags
        cmdVUp?.flags = flags
        
        let loc = CGEventTapLocation.cghidEventTap
        cmdVDown?.post(tap: loc)
        cmdVUp?.post(tap: loc)
    }
}
