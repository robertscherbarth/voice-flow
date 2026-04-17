import Foundation

class AgentClient {
    let baseURL = "http://localhost:8080/process"
    
    struct ProcessResponse: Decodable {
        let text: String
    }
    
    func processAudio(fileURL: URL, systemPrompt: String, completion: @escaping (Result<String, Error>) -> Void) {
        var request = URLRequest(url: URL(string: baseURL)!)
        request.httpMethod = "POST"
        
        let boundary = UUID().uuidString
        request.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")
        
        var body = Data()
        let fields = ["system_prompt": systemPrompt]
        for (key, value) in fields {
            body.append("--\(boundary)\r\n".data(using: .utf8)!)
            body.append("Content-Disposition: form-data; name=\"\(key)\"\r\n\r\n".data(using: .utf8)!)
            body.append("\(value)\r\n".data(using: .utf8)!)
        }
        
        if let audioData = try? Data(contentsOf: fileURL) {
            body.append("--\(boundary)\r\n".data(using: .utf8)!)
            body.append("Content-Disposition: form-data; name=\"audio\"; filename=\"recording.m4a\"\r\n".data(using: .utf8)!)
            body.append("Content-Type: audio/mp4\r\n\r\n".data(using: .utf8)!)
            body.append(audioData)
            body.append("\r\n".data(using: .utf8)!)
        }
        body.append("--\(boundary)--\r\n".data(using: .utf8)!)
        request.httpBody = body
        
        let task = URLSession.shared.dataTask(with: request) { data, response, error in
            if let error = error {
                completion(.failure(error))
                return
            }
            
            guard let data = data else {
                completion(.failure(NSError(domain: "AgentClient", code: -1, userInfo: [NSLocalizedDescriptionKey: "No data received"])))
                return
            }
            
            // Check if it's SSE or JSON
            if let str = String(data: data, encoding: .utf8), str.hasPrefix("data: ") {
                let text = str.components(separatedBy: "\n")
                    .filter { $0.hasPrefix("data: ") }
                    .map { $0.dropFirst(6).trimmingCharacters(in: .whitespacesAndNewlines) }
                    .joined()
                completion(.success(text))
            } else {
                do {
                    let res = try JSONDecoder().decode(ProcessResponse.self, from: data)
                    completion(.success(res.text))
                } catch {
                    completion(.failure(error))
                }
            }
        }
        task.resume()
    }
}
