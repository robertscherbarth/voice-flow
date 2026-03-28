import Cocoa
import SwiftUI

class AppDelegate: NSObject, NSApplicationDelegate {
    var statusItem: NSStatusItem!
    var preferencesWindow: NSWindow?
    
    let hotkeyManager = HotkeyManager()
    let audioRecorder = AudioRecorder()
    let agentClient = AgentClient()
    let agentManager = AgentManager()
    
    func applicationDidFinishLaunching(_ aNotification: Notification) {
        statusItem = NSStatusBar.system.statusItem(withLength: NSStatusItem.variableLength)
        
        updateIcon("⚪️") // Idle
        setupMenu()
        
        if #available(macOS 14.0, *) {
            FloatingIndicatorManager.shared.show()
            FloatingIndicatorManager.shared.updateState(.idle)
        }
        
        agentManager.start()
        
        hotkeyManager.onRecordingStart = { [weak self] in
            self?.startRecordingFlow()
        }
        
        hotkeyManager.onRecordingStop = { [weak self] in
            self?.stopRecordingFlow()
        }
        
        hotkeyManager.startListening()
        
        audioRecorder.requestPermission { granted in
            if !granted {
                print("Microphone access denied!")
            }
        }
        
        checkAccessibilityPermissions()
    }
    
    func checkAccessibilityPermissions() {
        let options = [kAXTrustedCheckOptionPrompt.takeUnretainedValue() as String: true] as CFDictionary
        let trusted = AXIsProcessTrustedWithOptions(options)
        if !trusted {
            print("Accessibility permission not granted. Please enable it in System Settings.")
        } else {
            print("Accessibility permission granted.")
        }
    }
    
    func applicationWillTerminate(_ aNotification: Notification) {
        agentManager.stop()
        hotkeyManager.stopListening()
    }
    
    func setupMenu() {
        let menu = NSMenu()
        menu.addItem(NSMenuItem(title: "Preferences...", action: #selector(showPreferences), keyEquivalent: ","))
        menu.addItem(NSMenuItem.separator())
        menu.addItem(NSMenuItem(title: "Quit", action: #selector(NSApplication.terminate(_:)), keyEquivalent: "q"))
        statusItem.menu = menu
    }
    
    func updateIcon(_ icon: String) {
        DispatchQueue.main.async {
            self.statusItem.button?.title = icon
        }
    }
    
    func startRecordingFlow() {
        updateIcon("🔴") // Recording
        if #available(macOS 14.0, *) {
            FloatingIndicatorManager.shared.updateState(.recording)
        }
        audioRecorder.startRecording()
    }
    
    func stopRecordingFlow() {
        updateIcon("⏳") // Processing
        if #available(macOS 14.0, *) {
            FloatingIndicatorManager.shared.updateState(.waiting)
        }
        guard let url = audioRecorder.stopRecording() else {
            updateIcon("⚪️")
            if #available(macOS 14.0, *) {
                FloatingIndicatorManager.shared.updateState(.idle)
            }
            return
        }
        
        let prefs = Preferences.shared
        agentClient.processAudio(fileURL: url, sttModel: prefs.sttModel, llmModel: prefs.llmModel, systemPrompt: prefs.systemPrompt) { [weak self] result in
            DispatchQueue.main.async {
                guard let self = self else { return }
                self.updateIcon("⚪️") // Back to idle
                if #available(macOS 14.0, *) {
                    FloatingIndicatorManager.shared.updateState(.idle)
                }
                switch result {
                case .success(let text):
                    print("Received text: \(text)")
                    ClipboardManager.paste(text: text)
                case .failure(let error):
                    print("Error processing audio: \(error)")
                    NSSound.beep()
                }
            }
        }
    }
    
    @objc func showPreferences() {
        if preferencesWindow == nil {
            let contentView = PreferencesView()
            let window = NSWindow(
                contentRect: NSRect(x: 0, y: 0, width: 400, height: 300),
                styleMask: [.titled, .closable, .miniaturizable],
                backing: .buffered, defer: false)
            window.isReleasedWhenClosed = false
            window.center()
            window.setFrameAutosaveName("Preferences")
            window.contentView = NSHostingView(rootView: contentView)
            window.title = "Preferences"
            preferencesWindow = window
        }
        preferencesWindow?.makeKeyAndOrderFront(nil)
        NSApp.activate(ignoringOtherApps: true)
    }
}
