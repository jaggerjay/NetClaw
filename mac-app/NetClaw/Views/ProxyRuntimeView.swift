import SwiftUI

struct ProxyRuntimeView: View {
    @Binding var workingDirectoryText: String
    @Binding var commandText: String
    let proxyStatusText: String
    let isProxyRunning: Bool
    let logText: String
    let onStart: () -> Void
    let onStop: () -> Void
    let onClearLog: () -> Void
    let onUseSuggestedCommand: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Label(proxyStatusText, systemImage: isProxyRunning ? "bolt.fill" : "bolt.slash")
                    .foregroundStyle(isProxyRunning ? .green : .secondary)
                Spacer()
                Button("Suggested") {
                    onUseSuggestedCommand()
                }
                Button("Start Proxy") {
                    onStart()
                }
                .disabled(isProxyRunning)
                Button("Stop") {
                    onStop()
                }
                .disabled(!isProxyRunning)
            }

            VStack(alignment: .leading, spacing: 6) {
                Text("Working Directory")
                    .font(.caption)
                    .foregroundStyle(.secondary)
                TextField("/path/to/netclaw-repo/proxy-core", text: $workingDirectoryText)
                    .textFieldStyle(.roundedBorder)
            }

            VStack(alignment: .leading, spacing: 6) {
                HStack {
                    Text("Launch Command")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                    Spacer()
                    Text("Runs with /bin/bash -lc")
                        .font(.caption2)
                        .foregroundStyle(.secondary)
                }
                TextEditor(text: $commandText)
                    .font(.system(.caption, design: .monospaced))
                    .frame(minHeight: 80)
                    .overlay(
                        RoundedRectangle(cornerRadius: 8)
                            .stroke(Color.secondary.opacity(0.2), lineWidth: 1)
                    )
            }

            VStack(alignment: .leading, spacing: 6) {
                HStack {
                    Text("Proxy Logs")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                    Spacer()
                    Button("Clear Log") {
                        onClearLog()
                    }
                    .buttonStyle(.borderless)
                }
                ScrollView {
                    Text(logText.isEmpty ? "No proxy logs yet" : logText)
                        .font(.system(.caption, design: .monospaced))
                        .textSelection(.enabled)
                        .frame(maxWidth: .infinity, alignment: .leading)
                }
                .frame(minHeight: 120, maxHeight: 180)
                .padding(8)
                .background(Color.black.opacity(0.06))
                .clipShape(RoundedRectangle(cornerRadius: 8))
            }

            Text("Tip: set the working directory to your local proxy-core folder, then use the suggested command or customize it.")
                .font(.caption)
                .foregroundStyle(.secondary)
        }
        .padding()
        .background(.thinMaterial)
    }
}
