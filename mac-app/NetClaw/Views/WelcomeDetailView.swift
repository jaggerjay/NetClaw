import SwiftUI

struct WelcomeDetailView: View {
    var body: some View {
        ContentUnavailableView {
            Label("No Session Selected", systemImage: "network")
        } description: {
            Text("Choose a captured request from the sidebar to inspect headers, bodies, timing, and TLS interception details.")
        } actions: {
            VStack(alignment: .leading, spacing: 8) {
                Text("Quick start")
                    .font(.headline)
                Text("• Start proxy-core")
                Text("• Send traffic through the local proxy")
                Text("• Click Refresh if auto-refresh is off")
            }
            .font(.caption)
            .foregroundStyle(.secondary)
        }
    }
}
