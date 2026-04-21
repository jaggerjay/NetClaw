import SwiftUI

struct ConnectionStatusView: View {
    @Binding var apiBaseURLText: String
    let statusText: String
    let isConnected: Bool
    let isRefreshing: Bool
    let autoRefreshEnabled: Bool
    let authorityInfo: CertificateAuthorityInfo?
    let onRefresh: () -> Void
    let onApplyBaseURL: () -> Void
    let onToggleAutoRefresh: (Bool) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(spacing: 12) {
                Label(statusText, systemImage: isConnected ? "checkmark.circle.fill" : "exclamationmark.triangle.fill")
                    .foregroundStyle(isConnected ? .green : .orange)
                Spacer()
                if isRefreshing {
                    ProgressView()
                        .controlSize(.small)
                }
            }

            HStack {
                TextField("API Base URL", text: $apiBaseURLText)
                    .textFieldStyle(.roundedBorder)
                Button("Apply") {
                    onApplyBaseURL()
                }
                Button("Refresh") {
                    onRefresh()
                }
                .keyboardShortcut("r", modifiers: [.command])
            }

            Toggle("Auto refresh every 2 seconds", isOn: Binding(
                get: { autoRefreshEnabled },
                set: onToggleAutoRefresh
            ))
            .toggleStyle(.switch)

            Divider()

            VStack(alignment: .leading, spacing: 6) {
                Text("Test setup")
                    .font(.headline)
                Text("1. Start proxy-core locally")
                Text("2. Point your browser or curl at the proxy port")
                Text("3. Inspect captured sessions here")
            }
            .font(.caption)
            .foregroundStyle(.secondary)

            if let authorityInfo {
                Divider()
                VStack(alignment: .leading, spacing: 6) {
                    Text("Certificate Authority")
                        .font(.headline)
                    LabeledContent("Common Name", value: authorityInfo.commonName)
                    LabeledContent("Trusted", value: authorityInfo.trusted ? "Yes" : "Not yet")
                    LabeledContent("Certificate", value: authorityInfo.certificatePath)
                        .textSelection(.enabled)
                }
                .font(.caption)
            }
        }
        .padding()
        .background(.thinMaterial)
    }
}
