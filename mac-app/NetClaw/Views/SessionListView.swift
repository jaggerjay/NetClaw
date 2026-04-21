import SwiftUI

struct SessionListView: View {
    let sessions: [SessionSummary]
    let selectedSessionID: String?
    let onSelect: (SessionSummary) -> Void

    var body: some View {
        List(sessions, selection: .constant(selectedSessionID)) { session in
            Button {
                onSelect(session)
            } label: {
                VStack(alignment: .leading, spacing: 6) {
                    HStack {
                        Text(session.method)
                            .font(.caption.monospaced())
                            .padding(.horizontal, 6)
                            .padding(.vertical, 2)
                            .background(methodColor(for: session.method).opacity(0.15))
                            .clipShape(RoundedRectangle(cornerRadius: 4))
                            .foregroundStyle(methodColor(for: session.method))
                        Text(session.host)
                            .font(.headline)
                            .lineLimit(1)
                        Spacer()
                        Text(statusText(for: session))
                            .foregroundStyle(session.hasError ? .red : .secondary)
                    }
                    Text(session.url)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                    HStack(spacing: 8) {
                        Label("\(session.durationMs) ms", systemImage: "clock")
                        Label(ByteCountFormatter.string(fromByteCount: session.responseSize, countStyle: .file), systemImage: "arrow.down.circle")
                        if !session.contentType.isEmpty {
                            Text(session.contentType)
                                .lineLimit(1)
                        }
                    }
                    .font(.caption2)
                    .foregroundStyle(.secondary)
                }
                .padding(.vertical, 4)
                .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            .tag(session.id)
        }
        .listStyle(.sidebar)
    }

    private func statusText(for session: SessionSummary) -> String {
        session.hasError ? "Error" : "\(session.statusCode)"
    }

    private func methodColor(for method: String) -> Color {
        switch method.uppercased() {
        case "GET": return .blue
        case "POST": return .green
        case "PUT", "PATCH": return .orange
        case "DELETE": return .red
        case "CONNECT": return .purple
        default: return .secondary
        }
    }
}
