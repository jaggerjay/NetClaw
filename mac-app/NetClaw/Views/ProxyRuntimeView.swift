import SwiftUI
import AppKit

struct ProxyRuntimeView: View {
    @Binding var workingDirectoryText: String
    @Binding var commandText: String
    let proxyStatusText: String
    let proxyValidationText: String
    let isProxyRunning: Bool
    let logText: String
    let onStart: () -> Void
    let onStop: () -> Void
    let onClearLog: () -> Void
    let onValidate: () -> Void
    let onUseSuggestedCommand: () -> Void
    let onUseDebugBuildCommand: () -> Void
    let onUseRepoRootSuggestion: () -> Void
    let onWorkingDirectorySelected: (String) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Label(proxyStatusText, systemImage: isProxyRunning ? "bolt.fill" : "bolt.slash")
                    .foregroundStyle(isProxyRunning ? .green : .secondary)
                Spacer()
                Button("Repo Hint") {
                    onUseRepoRootSuggestion()
                }
                Button("Suggested") {
                    onUseSuggestedCommand()
                }
                Button("Build+Run") {
                    onUseDebugBuildCommand()
                }
                Button("Validate") {
                    onValidate()
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
                HStack {
                    TextField("/path/to/netclaw-repo/proxy-core", text: $workingDirectoryText)
                        .textFieldStyle(.roundedBorder)
                    Button("Choose…") {
                        chooseDirectory()
                    }
                }
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
                    .frame(minHeight: 100)
                    .overlay(
                        RoundedRectangle(cornerRadius: 8)
                            .stroke(Color.secondary.opacity(0.2), lineWidth: 1)
                    )
            }

            if !proxyValidationText.isEmpty {
                Label(proxyValidationText, systemImage: proxyValidationText.contains("valid") ? "checkmark.seal" : "exclamationmark.triangle")
                    .font(.caption)
                    .foregroundStyle(proxyValidationText.contains("valid") ? .green : .orange)
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

            Text("Tip: choose your local proxy-core folder, validate the settings, then use Suggested for a fast start or Build+Run to compile a local binary first.")
                .font(.caption)
                .foregroundStyle(.secondary)
        }
        .padding()
        .background(.thinMaterial)
    }

    private func chooseDirectory() {
        let panel = NSOpenPanel()
        panel.canChooseFiles = false
        panel.canChooseDirectories = true
        panel.allowsMultipleSelection = false
        panel.canCreateDirectories = false
        panel.prompt = "Choose"
        panel.message = "Select your local proxy-core directory"

        if panel.runModal() == .OK, let url = panel.url {
            onWorkingDirectorySelected(url.path)
        }
    }
}
