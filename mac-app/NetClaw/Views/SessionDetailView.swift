import SwiftUI

struct SessionDetailView: View {
    let session: SessionDetail?

    var body: some View {
        Group {
            if let session {
                ScrollView {
                    VStack(alignment: .leading, spacing: 16) {
                        header(session)
                        metadata(session)
                        if let error = session.error, !error.isEmpty {
                            errorBox(error)
                        }
                        headersSection(title: "Request Headers", headers: session.requestHeaders)
                        bodySection(title: "Request Body", data: session.requestBody, emptyText: "No captured request body")
                        headersSection(title: "Response Headers", headers: session.responseHeaders)
                        bodySection(title: "Response Body", data: session.responseBody, emptyText: "No captured response body")
                    }
                    .padding()
                }
            } else {
                WelcomeDetailView()
            }
        }
    }

    @ViewBuilder
    private func header(_ session: SessionDetail) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            Text(session.url)
                .font(.title3)
                .textSelection(.enabled)
            HStack(spacing: 12) {
                pill(session.method, color: .blue)
                pill(session.scheme.uppercased(), color: session.tlsIntercepted ? .green : .secondary)
                if session.statusCode > 0 {
                    pill("\(session.statusCode)", color: session.error == nil ? .secondary : .red)
                }
                Label("\(session.durationMs) ms", systemImage: "clock")
                    .foregroundStyle(.secondary)
            }
        }
    }

    @ViewBuilder
    private func metadata(_ session: SessionDetail) -> some View {
        GroupBox("Session") {
            VStack(alignment: .leading, spacing: 8) {
                LabeledContent("Host", value: session.host)
                LabeledContent("Port", value: String(session.port))
                LabeledContent("Client", value: session.clientAddress)
                LabeledContent("Started", value: session.startTime.formatted(date: .abbreviated, time: .standard))
                LabeledContent("Ended", value: session.endTime.formatted(date: .omitted, time: .standard))
                LabeledContent("Request Size", value: ByteCountFormatter.string(fromByteCount: session.requestSize, countStyle: .file))
                LabeledContent("Response Size", value: ByteCountFormatter.string(fromByteCount: session.responseSize, countStyle: .file))
                LabeledContent("Content Type", value: session.contentType.isEmpty ? "—" : session.contentType)
                LabeledContent("TLS Intercepted", value: session.tlsIntercepted ? "Yes" : "No")
                LabeledContent("Capture Mode", value: session.captureMode.isEmpty ? "—" : session.captureMode)
                if let tunnelTargetAddress = session.tunnelTargetAddress, !tunnelTargetAddress.isEmpty {
                    LabeledContent("Tunnel Target", value: tunnelTargetAddress)
                }
                if session.tunnelBytesUp > 0 || session.tunnelBytesDown > 0 {
                    LabeledContent("Tunnel Up", value: ByteCountFormatter.string(fromByteCount: session.tunnelBytesUp, countStyle: .file))
                    LabeledContent("Tunnel Down", value: ByteCountFormatter.string(fromByteCount: session.tunnelBytesDown, countStyle: .file))
                }
                if let fallbackReason = session.fallbackReason, !fallbackReason.isEmpty {
                    LabeledContent("Fallback Reason", value: fallbackReason)
                }
            }
            .font(.caption)
            .textSelection(.enabled)
        }
    }

    @ViewBuilder
    private func errorBox(_ error: String) -> some View {
        GroupBox {
            Text(error)
                .frame(maxWidth: .infinity, alignment: .leading)
                .foregroundStyle(.red)
                .textSelection(.enabled)
        } label: {
            Label("Error", systemImage: "exclamationmark.triangle.fill")
                .foregroundStyle(.red)
        }
    }

    @ViewBuilder
    private func headersSection(title: String, headers: [String: String]) -> some View {
        GroupBox(title) {
            Text(prettyHeaders(headers).isEmpty ? "No headers" : prettyHeaders(headers))
                .font(.system(.body, design: .monospaced))
                .frame(maxWidth: .infinity, alignment: .leading)
                .textSelection(.enabled)
        }
    }

    @ViewBuilder
    private func bodySection(title: String, data: Data?, emptyText: String) -> some View {
        GroupBox(title) {
            if let data, !data.isEmpty {
                Text(renderBody(data))
                    .font(.system(.body, design: .monospaced))
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .textSelection(.enabled)
            } else {
                Text(emptyText)
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
        }
    }

    private func prettyHeaders(_ headers: [String: String]) -> String {
        headers.keys.sorted().map { "\($0): \(headers[$0] ?? "")" }.joined(separator: "\n")
    }

    private func renderBody(_ data: Data) -> String {
        if let text = String(data: data, encoding: .utf8) {
            return text
        }
        return data.base64EncodedString()
    }

    private func pill(_ text: String, color: Color) -> some View {
        Text(text)
            .font(.caption.monospaced())
            .padding(.horizontal, 8)
            .padding(.vertical, 4)
            .background(color.opacity(0.15))
            .clipShape(Capsule())
            .foregroundStyle(color)
    }
}
