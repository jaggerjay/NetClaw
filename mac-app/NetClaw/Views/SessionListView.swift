import SwiftUI

struct SessionListView: View {
    let sessions: [SessionSummary]
    let onSelect: (SessionSummary) -> Void

    var body: some View {
        List(sessions) { session in
            Button {
                onSelect(session)
            } label: {
                VStack(alignment: .leading, spacing: 4) {
                    HStack {
                        Text(session.method)
                            .font(.caption.monospaced())
                            .foregroundStyle(.blue)
                        Text(session.host)
                            .font(.headline)
                        Spacer()
                        Text("\(session.statusCode)")
                            .foregroundStyle(session.hasError ? .red : .secondary)
                    }
                    Text(session.url)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                    HStack {
                        Text("\(session.durationMs) ms")
                        Text("•")
                        Text(ByteCountFormatter.string(fromByteCount: session.responseSize, countStyle: .file))
                    }
                    .font(.caption2)
                    .foregroundStyle(.secondary)
                }
                .padding(.vertical, 4)
            }
            .buttonStyle(.plain)
        }
        .listStyle(.sidebar)
    }
}
