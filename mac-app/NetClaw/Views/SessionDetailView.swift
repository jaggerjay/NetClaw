import SwiftUI
import AppKit

struct SessionDetailView: View {
    let session: SessionDetail?

    @State private var expandedBodySections: Set<String> = []
    private let previewCharacterLimit = 4000

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
            actionBar(session)
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
            VStack(alignment: .leading, spacing: 8) {
                HStack {
                    Spacer()
                    Button("Copy") {
                        copyToPasteboard(prettyHeaders(headers))
                    }
                    .buttonStyle(.borderless)
                }
                Text(prettyHeaders(headers).isEmpty ? "No headers" : prettyHeaders(headers))
                    .font(.system(.body, design: .monospaced))
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .fixedSize(horizontal: false, vertical: true)
                    .textSelection(.enabled)
                    .padding(10)
                    .background(Color.secondary.opacity(0.08))
                    .clipShape(RoundedRectangle(cornerRadius: 8))
            }
        }
    }

    @ViewBuilder
    private func bodySection(title: String, data: Data?, contentType: String, encoding: String?, truncated: Bool, emptyText: String) -> some View {
        GroupBox(title) {
            if let data, !data.isEmpty {
                let renderedBody = renderBody(data, contentType: contentType, encoding: encoding)
                let sectionID = title.lowercased().replacingOccurrences(of: " ", with: "-")
                let isExpanded = expandedBodySections.contains(sectionID)
                let previewText = previewBodyText(renderedBody, expanded: isExpanded)
                let canExpand = renderedBody.count > previewCharacterLimit

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
                        if canExpand {
                            pill(isExpanded ? "FULL" : "PREVIEW", color: .secondary)
                        }
                        Spacer()
                        Button("Copy") {
                            copyToPasteboard(renderedBody)
                        }
                        .buttonStyle(.borderless)
                    }

                    if let image = renderImage(data: data, contentType: contentType) {
                        Image(nsImage: image)
                            .resizable()
                            .scaledToFit()
                            .frame(maxHeight: 260)
                            .clipShape(RoundedRectangle(cornerRadius: 8))
                    } else {
                        if canExpand && !isExpanded {
                            Label("Previewing the first \(previewCharacterLimit) characters", systemImage: "eye")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }

                        Text(previewText)
                            .font(.system(.body, design: .monospaced))
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .fixedSize(horizontal: false, vertical: true)
                            .textSelection(.enabled)
                            .padding(10)
                            .background(Color.secondary.opacity(0.08))
                            .clipShape(RoundedRectangle(cornerRadius: 8))

                        if canExpand {
                            HStack {
                                Button(isExpanded ? "Show Less" : "Show All") {
                                    toggleBodyExpansion(sectionID)
                                }
                                .buttonStyle(.borderedProminent)
                                .controlSize(.small)
                                Spacer()
                            }
                        }
                    }
                }
            } else {
                Text(emptyText)
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
        }
    }

    @ViewBuilder
    private func actionBar(_ session: SessionDetail) -> some View {
        FlowLayout(spacing: 8) {
            actionButton("Copy URL") {
                copyToPasteboard(session.url)
            }
            actionButton("Copy curl") {
                copyToPasteboard(makeCurlCommand(session))
            }
            actionButton("Copy Req Headers") {
                copyToPasteboard(prettyHeaders(session.requestHeaders))
            }
            actionButton("Copy Res Headers") {
                copyToPasteboard(prettyHeaders(session.responseHeaders))
            }
            if let requestBody = session.requestBody, !requestBody.isEmpty {
                actionButton("Copy Req Body") {
                    copyToPasteboard(renderBody(requestBody, contentType: headerValue(session.requestHeaders, key: "Content-Type"), encoding: session.requestBodyEncoding))
                }
            }
            if let responseBody = session.responseBody, !responseBody.isEmpty {
                actionButton("Copy Res Body") {
                    copyToPasteboard(renderBody(responseBody, contentType: session.contentType, encoding: session.responseBodyEncoding))
                }
            }
        }
    }

    private func previewBodyText(_ value: String, expanded: Bool) -> String {
        guard !expanded, value.count > previewCharacterLimit else {
            return value
        }
        let end = value.index(value.startIndex, offsetBy: previewCharacterLimit)
        return String(value[..<end]) + "\n\n… [preview truncated, click Show All to view full body]"
    }

    private func toggleBodyExpansion(_ sectionID: String) {
        if expandedBodySections.contains(sectionID) {
            expandedBodySections.remove(sectionID)
        } else {
            expandedBodySections.insert(sectionID)
        }
    }

    private func prettyHeaders(_ headers: [String: String]) -> String {
        headers.keys.sorted().map { "\($0): \(headers[$0] ?? "")" }.joined(separator: "\n")
    }

    private func renderBody(_ data: Data, contentType: String, encoding: String?) -> String {
        if isJSON(contentType: contentType),
           let object = try? JSONSerialization.jsonObject(with: data),
           let pretty = try? JSONSerialization.data(withJSONObject: object, options: [.prettyPrinted, .sortedKeys]),
           let text = String(data: pretty, encoding: .utf8) {
            return text
        }
        if isFormURLEncoded(contentType: contentType),
           let text = String(data: data, encoding: .utf8) {
            return prettyFormURLEncoded(text)
        }
        if isXML(contentType: contentType),
           let prettyXML = prettyXMLString(data: data) {
            return prettyXML
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

    private func isFormURLEncoded(contentType: String) -> Bool {
        contentType.localizedCaseInsensitiveContains("application/x-www-form-urlencoded")
    }

    private func isXML(contentType: String) -> Bool {
        let lowered = contentType.lowercased()
        return lowered.contains("xml") || lowered.contains("+xml")
    }

    private func prettyFormURLEncoded(_ text: String) -> String {
        text
            .split(separator: "&", omittingEmptySubsequences: false)
            .map { pair in
                let parts = pair.split(separator: "=", maxSplits: 1, omittingEmptySubsequences: false)
                let rawKey = parts.first.map(String.init) ?? ""
                let rawValue = parts.count > 1 ? String(parts[1]) : ""
                return "\(decodeFormComponent(rawKey)): \(decodeFormComponent(rawValue))"
            }
            .joined(separator: "\n")
    }

    private func decodeFormComponent(_ value: String) -> String {
        value
            .replacingOccurrences(of: "+", with: " ")
            .removingPercentEncoding ?? value
    }

    private func prettyXMLString(data: Data) -> String? {
        guard let document = try? XMLDocument(data: data, options: [.nodePreserveAll]) else {
            return nil
        }
        document.characterEncoding = "utf-8"
        return document.xmlString(options: [.nodePrettyPrint])
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

    private func makeCurlCommand(_ session: SessionDetail) -> String {
        var parts: [String] = ["curl", "-X", shellEscape(session.method)]

        let headerBlacklist = Set(["host", "content-length"])
        for key in session.requestHeaders.keys.sorted() {
            guard !headerBlacklist.contains(key.lowercased()) else { continue }
            let value = session.requestHeaders[key] ?? ""
            parts.append("-H")
            parts.append(shellEscape("\(key): \(value)"))
        }

        if let body = session.requestBody, !body.isEmpty {
            parts.append("--data-binary")
            parts.append(shellEscape(renderBody(body, contentType: headerValue(session.requestHeaders, key: "Content-Type"), encoding: session.requestBodyEncoding)))
        }

        parts.append(shellEscape(session.url))
        return parts.joined(separator: " ")
    }

    private func shellEscape(_ value: String) -> String {
        "'" + value.replacingOccurrences(of: "'", with: "'\\''") + "'"
    }

    private func copyToPasteboard(_ value: String) {
        let pasteboard = NSPasteboard.general
        pasteboard.clearContents()
        pasteboard.setString(value, forType: .string)
    }

    @ViewBuilder
    private func actionButton(_ title: String, action: @escaping () -> Void) -> some View {
        Button(title, action: action)
            .buttonStyle(.bordered)
            .controlSize(.small)
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

private struct FlowLayout: Layout {
    var spacing: CGFloat = 8

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        let maxWidth = proposal.width ?? 600
        var x: CGFloat = 0
        var y: CGFloat = 0
        var rowHeight: CGFloat = 0

        for subview in subviews {
            let size = subview.sizeThatFits(.unspecified)
            if x + size.width > maxWidth, x > 0 {
                x = 0
                y += rowHeight + spacing
                rowHeight = 0
            }
            rowHeight = max(rowHeight, size.height)
            x += size.width + spacing
        }
        return CGSize(width: maxWidth, height: y + rowHeight)
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        var x = bounds.minX
        var y = bounds.minY
        var rowHeight: CGFloat = 0

        for subview in subviews {
            let size = subview.sizeThatFits(.unspecified)
            if x + size.width > bounds.maxX, x > bounds.minX {
                x = bounds.minX
                y += rowHeight + spacing
                rowHeight = 0
            }
            subview.place(at: CGPoint(x: x, y: y), proposal: ProposedViewSize(size))
            x += size.width + spacing
            rowHeight = max(rowHeight, size.height)
        }
    }
}
