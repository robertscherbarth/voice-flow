import Cocoa

class AgentManager {
    var process: Process?
    
    func start() {
        guard let agentPath = Bundle.main.path(forResource: "voice-agent", ofType: nil) else {
            print("Could not find voice-agent binary in bundle")
            return
        }
        
        process = Process()
        process?.executableURL = URL(fileURLWithPath: agentPath)
        
        // Optional: pipe output to a file or /dev/null
        let pipe = Pipe()
        process?.standardOutput = pipe
        process?.standardError = pipe
        
        do {
            try process?.run()
            print("Agent started at \(agentPath)")
        } catch {
            print("Failed to start agent: \(error)")
        }
    }
    
    func stop() {
        process?.terminate()
        process = nil
        print("Agent stopped")
    }
}
