import Cocoa

class HotkeyManager {
    var monitor: Any?
    var isRecording = false
    var isPermanentMode = false
    var lastFnPressTime: Date?
    var onRecordingStart: (() -> Void)?
    var onRecordingStop: (() -> Void)?

    func startListening() {
        let options: NSDictionary = [kAXTrustedCheckOptionPrompt.takeUnretainedValue() as String : true]
        let accessEnabled = AXIsProcessTrustedWithOptions(options)

        if !accessEnabled {
            print("Accessibility access not enabled!")
        }

        monitor = NSEvent.addGlobalMonitorForEvents(matching: .flagsChanged) { [weak self] event in
            guard let self = self else { return }
            
            // Check if ONLY the function key is pressed
            let fnKeyPressed = event.modifierFlags.contains(.function)
            let otherModifiers = event.modifierFlags.intersection([.command, .option, .control, .shift])
            
            guard otherModifiers.isEmpty else {
                if self.isRecording && !self.isPermanentMode {
                    self.isRecording = false
                    self.onRecordingStop?()
                }
                return
            }
            
            if fnKeyPressed {
                let now = Date()
                if let lastPress = self.lastFnPressTime, now.timeIntervalSince(lastPress) < 0.4 {
                    self.isPermanentMode.toggle()
                    self.lastFnPressTime = nil
                    
                    if self.isPermanentMode {
                        if !self.isRecording {
                            self.isRecording = true
                            self.onRecordingStart?()
                        }
                    } else {
                        if self.isRecording {
                            self.isRecording = false
                            self.onRecordingStop?()
                        }
                    }
                } else {
                    self.lastFnPressTime = now
                    if !self.isRecording && !self.isPermanentMode {
                        self.isRecording = true
                        self.onRecordingStart?()
                    }
                }
            } else {
                if self.isRecording && !self.isPermanentMode {
                    self.isRecording = false
                    self.onRecordingStop?()
                }
            }
        }
    }

    func stopListening() {
        if let monitor = monitor {
            NSEvent.removeMonitor(monitor)
            self.monitor = nil
        }
    }
}
