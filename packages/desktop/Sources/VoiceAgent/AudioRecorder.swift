import Cocoa
import AVFoundation

class AudioRecorder: NSObject, AVAudioRecorderDelegate {
    var audioRecorder: AVAudioRecorder?
    let tempURL = URL(fileURLWithPath: NSTemporaryDirectory()).appendingPathComponent("recording.wav")
    var startTime: Date?

    func requestPermission(completion: @escaping (Bool) -> Void) {
        switch AVCaptureDevice.authorizationStatus(for: .audio) {
        case .authorized:
            completion(true)
        case .notDetermined:
            AVCaptureDevice.requestAccess(for: .audio) { granted in
                DispatchQueue.main.async {
                    completion(granted)
                }
            }
        case .denied, .restricted:
            completion(false)
        @unknown default:
            completion(false)
        }
    }

    func startRecording() {
        let settings: [String: Any] = [
            AVFormatIDKey: Int(kAudioFormatLinearPCM),
            AVSampleRateKey: 16000.0,
            AVNumberOfChannelsKey: 1,
            AVLinearPCMBitDepthKey: 16,
            AVLinearPCMIsFloatKey: false,
            AVLinearPCMIsBigEndianKey: false
        ]

        do {
            audioRecorder = try AVAudioRecorder(url: tempURL, settings: settings)
            audioRecorder?.delegate = self
            audioRecorder?.record()
            startTime = Date()
            print("Recording started at \(tempURL.path)")
        } catch {
            print("Failed to start recording: \(error.localizedDescription)")
        }
    }

    func stopRecording() -> URL? {
        audioRecorder?.stop()
        print("Recording stopped")
        
        if let start = startTime, Date().timeIntervalSince(start) < 0.3 {
            print("Recording too short, discarding")
            return nil
        }
        
        return tempURL
    }
}
