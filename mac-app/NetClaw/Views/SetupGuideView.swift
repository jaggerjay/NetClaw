import SwiftUI

struct SetupGuideView: View {
    let runtimeInfo: RuntimeInfo?
    let authorityInfo: CertificateAuthorityInfo?

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Setup Guide")
                .font(.headline)

            if let runtimeInfo {
                GroupBox("Local Proxy") {
                    VStack(alignment: .leading, spacing: 8) {
                        LabeledContent("Proxy Address", value: runtimeInfo.proxyListenAddress)
                        LabeledContent("API Address", value: runtimeInfo.apiListenAddress)
                        LabeledContent("Data Directory", value: runtimeInfo.dataDir)
                            .textSelection(.enabled)
                        Text("In macOS Wi-Fi settings, set a manual Web Proxy (HTTP) and Secure Web Proxy (HTTPS) to this address if you want system traffic to pass through NetClaw.")
                            .foregroundStyle(.secondary)
                    }
                    .font(.caption)
                }
            } else {
                Text("Runtime info will appear after the local API responds.")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            GroupBox("Certificate Trust") {
                VStack(alignment: .leading, spacing: 8) {
                    if let authorityInfo {
                        LabeledContent("Certificate", value: authorityInfo.certificatePath)
                            .textSelection(.enabled)
                        Text("1. Open Keychain Access")
                        Text("2. Import the NetClaw root CA certificate")
                        Text("3. Set trust to 'Always Trust' for SSL when testing HTTPS MITM")
                        Text("4. Restart the target app or browser if needed")
                    } else {
                        Text("The root CA path will appear here once proxy-core is running.")
                            .foregroundStyle(.secondary)
                    }
                }
                .font(.caption)
            }

            GroupBox("Testing Tips") {
                VStack(alignment: .leading, spacing: 8) {
                    Text("• Start with plain HTTP to verify the proxy path")
                    Text("• Then test HTTPS with a simple site after trusting the root CA")
                    Text("• If HTTPS breaks for a host, try a different target first — some apps use pinning")
                }
                .font(.caption)
                .foregroundStyle(.secondary)
            }
        }
        .padding()
        .background(.thinMaterial)
    }
}
