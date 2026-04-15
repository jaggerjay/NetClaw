import SwiftUI

struct SessionDetailView: View {
    let session: SessionDetail?

    var body: some View {
        Group {
            if let session {
                VStack(alignment: .leading, spacing: 12) {
                    Text(session.url)
                        .font(.headline)
                    HStack(spacing: 16) {
                        Label(session.method, systemImage: "arrow.up.right")
                        Label("\(session.statusCode)", systemImage: "arrow.down.right")
                        Label("\(session.durationMs) ms", systemImage: "clock")
                    }
                    .font(.subheadline)

                    Divider()

                    Text("Request Headers")
                        .font(.headline)
                    ScrollView {
                        Text(prettyHeaders(session.requestHeaders))
                            .font(.system(.body, design: .monospaced))
                            .frame(maxWidth: .infinity, alignment: .leading)
                    }

                    Divider()

                    Text("Response Headers")
                        .font(.headline)
                    ScrollView {
                        Text(prettyHeaders(session.responseHeaders))
                            .font(.system(.body, design: .monospaced))
                            .frame(maxWidth: .infinity, alignment: .leading)
                    }
                }
                .padding()
            } else {
                ContentUnavailableView("No Session Selected", systemImage: "network", description: Text("Choose a captured request from the list."))
            }
        }
    }

    private func prettyHeaders(_ headers: [String: String]) -> String {
        headers.keys.sorted().map { "\($0): \(headers[$0] ?? "")" }.joined(separator: "\n")
    }
}
