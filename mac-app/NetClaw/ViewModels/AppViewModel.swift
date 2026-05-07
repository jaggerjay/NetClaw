import Foundation
import SwiftUI

@MainActor
final class AppViewModel: ObservableObject {
    @Published var sessions: [SessionSummary] = []
    @Published var selectedSession: SessionDetail?
    @Published var selectedSessionID: String?
    @Published var searchText: String = ""
    @Published var methodFilter: String = ""
    @Published var hostFilter: String = ""
    @Published var hasErrorOnly: Bool = false
    @Published var interceptedOnly: Bool = false
    @Published var statusText: String = "Proxy not connected"
    @Published var apiBaseURLText: String
    @Published var authorityInfo: CertificateAuthorityInfo?
    @Published var runtimeInfo: RuntimeInfo?
    @Published var isRefreshing: Bool = false
    @Published var autoRefreshEnabled: Bool
    @Published var proxyWorkingDirectoryText: String
    @Published var proxyCommandText: String
    @Published var proxyStatusText: String = "Proxy process not started"
    @Published var proxyValidationText: String = ""
    @Published var proxyLogText: String = ""
    @Published var isProxyRunning: Bool = false
    @Published var isExportingHAR: Bool = false
    @Published var lastErrorText: String?

    static let apiBaseURLKey = "netclaw.apiBaseURL"
    static let proxyWorkingDirectoryKey = "netclaw.proxyWorkingDirectory"
    static let proxyCommandKey = "netclaw.proxyCommand"
    static let autoRefreshKey = "netclaw.autoRefresh"

    var isConnected: Bool {
        statusText == "Connected"
    }

    private let defaults: UserDefaults
    private let apiClient: APIClient
    private let proxyController = ProxyProcessController()
    private var autoRefreshTask: Task<Void, Never>?

    init(defaults: UserDefaults = .standard) {
        self.defaults = defaults

        let defaultAPIBaseURL = defaults.string(forKey: Self.apiBaseURLKey) ?? "http://127.0.0.1:9091"
        let defaultWorkingDirectory = defaults.string(forKey: Self.proxyWorkingDirectoryKey) ?? FileManager.default.currentDirectoryPath
        let defaultAutoRefresh = defaults.object(forKey: Self.autoRefreshKey) as? Bool ?? true
        let defaultCommand = defaults.string(forKey: Self.proxyCommandKey) ?? ""

        apiBaseURLText = defaultAPIBaseURL
        proxyWorkingDirectoryText = defaultWorkingDirectory
        proxyCommandText = defaultCommand
        autoRefreshEnabled = defaultAutoRefresh
        apiClient = APIClient(baseURL: URL(string: defaultAPIBaseURL) ?? URL(string: "http://127.0.0.1:9091")!)

        proxyController.onOutput = { [weak self] text in
            self?.appendProxyLog(text)
        }
        proxyController.onStateChange = { [weak self] isRunning, message in
            self?.isProxyRunning = isRunning
            self?.proxyStatusText = message
        }
    }

    deinit {
        autoRefreshTask?.cancel()
        proxyController.stop()
    }

    func start() {
        if proxyCommandText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            useSuggestedProxyCommand()
        }
        configureAutoRefresh()
    }

    func setAutoRefreshEnabled(_ enabled: Bool) {
        autoRefreshEnabled = enabled
        defaults.set(enabled, forKey: Self.autoRefreshKey)
        configureAutoRefresh()
    }

    func updateWorkingDirectory(_ path: String) {
        proxyWorkingDirectoryText = path
        defaults.set(path, forKey: Self.proxyWorkingDirectoryKey)
        validateProxyLaunchSettings()
    }

    func applyAPIBaseURL() async {
        let trimmed = apiBaseURLText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard let url = URL(string: trimmed) else {
            statusText = "Invalid API base URL"
            lastErrorText = "The API base URL is not a valid URL"
            return
        }
        defaults.set(trimmed, forKey: Self.apiBaseURLKey)
        apiClient.updateBaseURL(url)
        await refresh()
    }

    func quickHealthCheck() async {
        await refresh()
    }

    func exportHARData() async throws -> Data {
        isExportingHAR = true
        defer { isExportingHAR = false }
        return try await apiClient.downloadHAR(query: currentQuery)
    }

    func recordHARExportResult(_ result: Result<URL, Error>) {
        switch result {
        case .success:
            lastErrorText = nil
        case .failure(let error):
            lastErrorText = "HAR export failed: \(error.localizedDescription)"
        }
    }

    func validateProxyLaunchSettings() {
        let result = proxyController.validateLaunch(
            command: proxyCommandText,
            workingDirectory: proxyWorkingDirectoryText
        )
        proxyValidationText = result.message
        if !result.success {
            lastErrorText = result.message
        }
    }

    func startProxy() {
        persistLaunchSettings()
        let validation = proxyController.validateLaunch(
            command: proxyCommandText,
            workingDirectory: proxyWorkingDirectoryText
        )
        proxyValidationText = validation.message
        guard validation.success else {
            proxyStatusText = validation.message
            lastErrorText = validation.message
            appendProxyLog("\n[netclaw] launch validation failed: \(validation.message)\n")
            return
        }

        do {
            try proxyController.start(
                command: proxyCommandText,
                workingDirectory: proxyWorkingDirectoryText
            )
            lastErrorText = nil
            appendProxyLog("\n[netclaw] start requested\n")
        } catch {
            proxyStatusText = error.localizedDescription
            lastErrorText = error.localizedDescription
            appendProxyLog("\n[netclaw] failed to start: \(error.localizedDescription)\n")
        }
    }

    func stopProxy() {
        proxyController.stop()
    }

    func clearProxyLog() {
        proxyLogText = ""
    }

    func useSuggestedProxyCommand() {
        let workingDir = proxyWorkingDirectoryText.trimmingCharacters(in: .whitespacesAndNewlines)
        let dataDir = workingDir.isEmpty ? ".netclaw-data/dev" : "\(workingDir)/.netclaw-data/dev"
        proxyCommandText = "go run ./cmd/netclaw-proxy -proxy-listen 127.0.0.1:9090 -api-listen 127.0.0.1:9091 -data-dir \"\(dataDir)\" -max-body-bytes 0"
        defaults.set(proxyCommandText, forKey: Self.proxyCommandKey)
        validateProxyLaunchSettings()
    }

    func useDebugBuildCommand() {
        proxyCommandText = "go build -o ./.netclaw-data/dev/netclaw-proxy ./cmd/netclaw-proxy && ./.netclaw-data/dev/netclaw-proxy -proxy-listen 127.0.0.1:9090 -api-listen 127.0.0.1:9091 -data-dir ./.netclaw-data/dev -max-body-bytes 0"
        defaults.set(proxyCommandText, forKey: Self.proxyCommandKey)
        validateProxyLaunchSettings()
    }

    func useRepoRootSuggestion() {
        let cwd = FileManager.default.currentDirectoryPath
        let candidate = (cwd as NSString).appendingPathComponent("proxy-core")
        updateWorkingDirectory(candidate)
    }

    func refresh() async {
        isRefreshing = true
        defer { isRefreshing = false }

        do {
            _ = try await apiClient.fetchHealth()
            async let authority = try? apiClient.fetchCertificateAuthorityInfo()
            async let runtimeInfo = try? apiClient.fetchRuntimeInfo()
            async let items = apiClient.fetchSessions(query: currentQuery)
            sessions = try await items
            authorityInfo = await authority
            self.runtimeInfo = await runtimeInfo
            statusText = "Connected"
            lastErrorText = nil

            if let selectedSessionID,
               sessions.contains(where: { $0.id == selectedSessionID }) {
                await select(sessionID: selectedSessionID)
            } else if let first = sessions.first, selectedSession == nil {
                await select(sessionID: first.id)
            }
        } catch {
            statusText = "Unable to reach local proxy API"
            lastErrorText = error.localizedDescription
            sessions = []
            authorityInfo = nil
            runtimeInfo = nil
            selectedSession = nil
        }
    }

    func select(sessionID: String) async {
        selectedSessionID = sessionID
        do {
            selectedSession = try await apiClient.fetchSession(id: sessionID)
        } catch {
            selectedSession = nil
        }
    }

    func clearFilters() {
        searchText = ""
        methodFilter = ""
        hostFilter = ""
        hasErrorOnly = false
        interceptedOnly = false
    }

    private func configureAutoRefresh() {
        autoRefreshTask?.cancel()
        guard autoRefreshEnabled else { return }

        autoRefreshTask = Task { [weak self] in
            guard let self else { return }
            while !Task.isCancelled {
                await self.refresh()
                try? await Task.sleep(for: .seconds(2))
            }
        }
    }

    private func appendProxyLog(_ text: String) {
        proxyLogText.append(text)
        if proxyLogText.count > 20000 {
            proxyLogText = String(proxyLogText.suffix(20000))
        }
    }

    private func persistLaunchSettings() {
        defaults.set(proxyWorkingDirectoryText, forKey: Self.proxyWorkingDirectoryKey)
        defaults.set(proxyCommandText, forKey: Self.proxyCommandKey)
    }

    func suggestedHARFilename() -> String {
        var parts = ["netclaw"]
        if !hostFilter.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            parts.append(hostFilter.replacingOccurrences(of: " ", with: "-"))
        }
        if !methodFilter.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            parts.append(methodFilter.lowercased())
        }
        parts.append("export.har")
        return parts.joined(separator: "-")
    }

    private var currentQuery: SessionQuery {
        SessionQuery(
            text: searchText,
            method: methodFilter,
            host: hostFilter,
            hasError: hasErrorOnly ? true : nil,
            tlsIntercepted: interceptedOnly ? true : nil,
            limit: 500
        )
    }
}
