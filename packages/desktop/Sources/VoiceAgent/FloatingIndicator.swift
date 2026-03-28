import Cocoa
import SwiftUI

enum RecordingState: Equatable {
    case idle
    case recording
    case waiting
}

class FloatingIndicatorManager: NSObject {
    static let shared = FloatingIndicatorManager()
    private var panel: NSPanel?
    
    class IndicatorState: ObservableObject {
        @Published var state: RecordingState = .idle
    }
    
    let stateObj = IndicatorState()
    
    func show() {
        if panel == nil {
            let view = FloatingIndicatorView(stateObj: stateObj)
            let hostingController = NSHostingController(rootView: view)
            
            let p = NSPanel(
                contentRect: NSRect(x: 0, y: 0, width: 60, height: 60),
                styleMask: [.borderless, .nonactivatingPanel],
                backing: .buffered,
                defer: false
            )
            p.level = .floating
            p.backgroundColor = .clear
            p.isOpaque = false
            p.hasShadow = false
            p.isMovableByWindowBackground = true
            p.collectionBehavior = [.canJoinAllSpaces, .fullScreenAuxiliary, .stationary]
            
            p.contentViewController = hostingController
            
            if let screen = NSScreen.main {
                let rect = p.frame
                let screenRect = screen.visibleFrame
                let newFrame = NSRect(
                    x: screenRect.midX - (rect.width / 2),
                    y: screenRect.maxY - rect.height - 20,
                    width: rect.width,
                    height: rect.height
                )
                p.setFrame(newFrame, display: true)
                print("Setting FloatingIndicator frame to: \(newFrame)")
            } else {
                p.center()
                print("Centered FloatingIndicator")
            }
            
            self.panel = p
        }
        panel?.orderFrontRegardless()
        print("Ordered FloatingIndicator front")
    }
    
    func updateState(_ newState: RecordingState) {
        DispatchQueue.main.async {
            self.stateObj.state = newState
        }
    }
}

@available(macOS 14.0, *)
struct FloatingIndicatorView: View {
    @ObservedObject var stateObj: FloatingIndicatorManager.IndicatorState
    
    var body: some View {
        ZStack {
            Circle()
                .fill(.regularMaterial)
                .shadow(color: .black.opacity(0.3), radius: 4, x: 0, y: 2)
            
            icon
                .font(.system(size: 20, weight: .semibold))
                .foregroundColor(tintColor)
                .contentTransition(.symbolEffect(.replace))
        }
        .frame(width: 44, height: 44)
        .padding(8)
    }
    
    @ViewBuilder
    var icon: some View {
        switch stateObj.state {
        case .idle:
            Image(systemName: "mic.slash.fill")
                .symbolEffect(.bounce, value: stateObj.state)
        case .recording:
            Image(systemName: "mic.fill")
                .symbolEffect(.pulse.byLayer, options: .repeating, isActive: stateObj.state == .recording)
        case .waiting:
            Image(systemName: "waveform")
                .symbolEffect(.variableColor.iterative, options: .repeating, isActive: stateObj.state == .waiting)
        }
    }
    
    var tintColor: Color {
        switch stateObj.state {
        case .idle: return .secondary
        case .recording: return .red
        case .waiting: return .orange
        }
    }
}
