import SwiftUI

class Preferences {
    static let shared = Preferences()
    
    var sttModel: String {
        UserDefaults.standard.string(forKey: "sttModel") ?? "voxtral-mini-latest"
    }
    var llmModel: String {
        UserDefaults.standard.string(forKey: "llmModel") ?? "mistral-small-latest"
    }
    var systemPrompt: String {
        UserDefaults.standard.string(forKey: "systemPrompt") ?? "You are a precise audio transcription editor. Your ONLY job is to remove filler words (like 'um', 'uh', 'like') and fix obvious grammatical errors from the provided spoken transcript.\n\nCRITICAL RULES:\n1. DO NOT change the perspective or pronouns (e.g., if the user says 'you', keep it as 'you'; do NOT change it to 'I').\n2. DO NOT rewrite the sentence to sound better if it changes the original meaning or tone.\n3. If the text is already clear, return it exactly as provided.\n4. Return ONLY the final text. Do not add quotes, explanations, or conversational filler."
    }
}

struct PreferencesView: View {
    @AppStorage("sttModel") private var sttModel = "voxtral-mini-latest"
    @AppStorage("llmModel") private var llmModel = "mistral-small-latest"
    @AppStorage("systemPrompt") private var systemPrompt = "You are a precise audio transcription editor. Your ONLY job is to remove filler words (like 'um', 'uh', 'like') and fix obvious grammatical errors from the provided spoken transcript.\n\nCRITICAL RULES:\n1. DO NOT change the perspective or pronouns (e.g., if the user says 'you', keep it as 'you'; do NOT change it to 'I').\n2. DO NOT rewrite the sentence to sound better if it changes the original meaning or tone.\n3. If the text is already clear, return it exactly as provided.\n4. Return ONLY the final text. Do not add quotes, explanations, or conversational filler."
    
    var body: some View {
        Form {
            Section(header: Text("Mistral Models")) {
                TextField("STT Model", text: $sttModel)
                TextField("LLM Model", text: $llmModel)
            }
            
            Section(header: Text("System Prompt")) {
                TextEditor(text: $systemPrompt)
                    .frame(height: 150)
                    .overlay(RoundedRectangle(cornerRadius: 8).stroke(Color.gray.opacity(0.2)))
            }
        }
        .padding()
        .frame(width: 450, height: 350)
    }
}
