import Cocoa

class AgentManager {
    var process: Process?
    
    func start() {
        guard let agentPath = Bundle.main.path(forResource: "server", ofType: nil) else {
            print("Could not find server binary in bundle")
            return
        }
        
        process = Process()
        process?.executableURL = URL(fileURLWithPath: agentPath)
        
        // Ensure the server can find prompt/optimize.yaml inside the bundle's Resources directory
        if let resourcePath = Bundle.main.resourcePath {
            process?.currentDirectoryURL = URL(fileURLWithPath: resourcePath)
        }
        
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
