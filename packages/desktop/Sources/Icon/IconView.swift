import SwiftUI

struct AppIconView: View {
    var body: some View {
        ZStack {
            // 1. Background Gradient
            LinearGradient(
                gradient: Gradient(colors: [
                    Color(red: 0.20, green: 0.44, blue: 0.45), // Top Left (Teal/Greenish)
                    Color(red: 0.18, green: 0.29, blue: 0.45), // Center
                    Color(red: 0.43, green: 0.24, blue: 0.55)  // Bottom Right (Purple)
                ]),
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
            .edgesIgnoringSafeArea(.all)
            
            // 2. Soundwave Path
            SoundwaveView()
                .stroke(Color.white, style: StrokeStyle(lineWidth: 24, lineCap: .round, lineJoin: .round))
                // Fading effect
                .shadow(color: Color.black.opacity(0.3), radius: 8, x: 0, y: 8)
            
            // 3. Microphone Symbol
            Image(systemName: "mic")
                .resizable()
                .aspectRatio(contentMode: .fit)
                .frame(width: 320, height: 320)
                .foregroundColor(.white)
                // SF Symbol stroke style to match the image
                .fontWeight(.medium) 
                .shadow(color: Color.black.opacity(0.4), radius: 10, x: 0, y: 10)
        }
        .frame(width: 1024, height: 1024)
        .clipShape(RoundedRectangle(cornerRadius: 220, style: .continuous)) // Apple icon standard corner radius ratio
    }
}

// Custom shape for the smooth sine-like soundwave
struct SoundwaveView: Shape {
    func path(in rect: CGRect) -> Path {
        var path = Path()
        
        let midY = rect.midY
        let width = rect.width
        
        path.move(to: CGPoint(x: 0, y: midY))
        
        // Left small wave
        path.addCurve(
            to: CGPoint(x: width * 0.25, y: midY - 60),
            control1: CGPoint(x: width * 0.1, y: midY),
            control2: CGPoint(x: width * 0.15, y: midY - 140)
        )
        
        // Trough
        path.addCurve(
            to: CGPoint(x: width * 0.35, y: midY + 120),
            control1: CGPoint(x: width * 0.3, y: midY + 40),
            control2: CGPoint(x: width * 0.3, y: midY + 140)
        )
        
        // Pass behind the mic (we just continue the path, mic goes over it in ZStack)
        path.addCurve(
            to: CGPoint(x: width * 0.65, y: midY - 100),
            control1: CGPoint(x: width * 0.45, y: midY + 80),
            control2: CGPoint(x: width * 0.55, y: midY - 180)
        )
        
        // Right small wave
        path.addCurve(
            to: CGPoint(x: width * 0.85, y: midY + 60),
            control1: CGPoint(x: width * 0.72, y: midY - 40),
            control2: CGPoint(x: width * 0.78, y: midY + 140)
        )
        
        // End point
        path.addCurve(
            to: CGPoint(x: width, y: midY),
            control1: CGPoint(x: width * 0.9, y: midY),
            control2: CGPoint(x: width * 0.95, y: midY)
        )
        
        return path
    }
}

struct AppIconView_Previews: PreviewProvider {
    static var previews: some View {
        AppIconView()
            .previewLayout(.sizeThatFits)
            .padding()
            .background(Color.gray)
    }
}