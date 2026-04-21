import SwiftUI
import AppKit

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
                        bodySection(
                            title: "Request Body",
                            data: session.requestBody,
                            contentType: headerValue(session.requestHeaders, key: "Content-Type"),
                            encoding: session.requestBodyEncoding,
                            truncated: session.requestBodyTruncated,
                            emptyText: "No captured request body"
                        )
                        headersSection(title: "Response Headers", headers: session.responseHeaders)
                        bodySection(
                            title: "Response Body",
                            data: session.responseBody,
                            contentType: session.contentType,
                            encoding: session.responseBodyEncoding,
                            truncated: session.responseBodyTruncated,
                            emptyText: "No captured response body"
                        )
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
    private func bodySection(title: String, data: Data?, contentType: String, encoding: String?, truncated: Bool, emptyText: String) -> some View {
        GroupBox(title) {
            if let data, !data.isEmpty {
                VStack(alignment: .leading, spacing: 8) {
                    HStack(spacing: 8) {
                        if !contentType.isEmpty {
                            pill(contentType, color: .secondary)
                        }
                        if let encoding, !encoding.isEmpty {
                            pill(encoding.uppercased(), color: .secondary)
                        }
                        if truncated {
                            pill("TRUNCATED", color: .orange)
                        }
                    }

                    if let image = renderImage(data: data, contentType: contentType) {
                        Image(nsImage: image)
                            .resizable()
                            .scaledToFit()
                            .frame(maxHeight: 260)
                            .clipShape(RoundedRectangle(cornerRadius: 8))
                    } else {
                        Text(renderBody(data, contentType: contentType, encoding: encoding))
                            .font(.system(.body, design: .monospaced))
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .textSelection(.enabled)
                    }
                }
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

    private func renderBody(_ data: Data, contentType: String, encoding: String?) -> String {
        if isJSON(contentType: contentType),
           let object = try? JSONSerialization.jsonObject(with: data),
           let pretty = try? JSONSerialization.data(withJSONObject: object, options: [.prettyPrinted]),
           let text = String(data: pretty, encoding: .utf8) {
            return text
        }
        if let text = String(data: data, encoding: .utf8) {
            return text
        }
        if encoding == "base64" {
            return data.base64EncodedString()
        }
        return data.base64EncodedString()
    }

    private func isJSON(contentType: String) -> Bool {
        contentType.localizedCaseInsensitiveContains("json")
    }

    private func renderImage(data: Data, contentType: String) -> NSImage? {
        guard contentType.localizedCaseInsensitiveContains("image/") else {
            return nil
        }
        return NSImage(data: data)
    }

    private func headerValue(_ headers: [String: String], key: String) -> String {
        headers.first { $0.key.caseInsensitiveCompare(key) == .orderedSame }?.value ?? ""
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
