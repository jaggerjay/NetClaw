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
    @Published var apiBaseURLText: String = "http://127.0.0.1:9091"
    @Published var authorityInfo: CertificateAuthorityInfo?
    @Published var isRefreshing: Bool = false
    @Published var autoRefreshEnabled: Bool = true

    var isConnected: Bool {
        statusText == "Connected"
    }

    private let apiClient = APIClient()
    private var autoRefreshTask: Task<Void, Never>?

    deinit {
        autoRefreshTask?.cancel()
    }

    func start() {
        configureAutoRefresh()
    }

    func setAutoRefreshEnabled(_ enabled: Bool) {
        autoRefreshEnabled = enabled
        configureAutoRefresh()
    }

    func applyAPIBaseURL() async {
        guard let url = URL(string: apiBaseURLText.trimmingCharacters(in: .whitespacesAndNewlines)) else {
            statusText = "Invalid API base URL"
            return
        }
        apiClient.updateBaseURL(url)
        await refresh()
    }

    func refresh() async {
        isRefreshing = true
        defer { isRefreshing = false }

        do {
            _ = try await apiClient.fetchHealth()
            async let authority = try? apiClient.fetchCertificateAuthorityInfo()
            async let items = apiClient.fetchSessions(query: currentQuery)
            sessions = try await items
            authorityInfo = await authority
            statusText = "Connected"

            if let selectedSessionID,
               sessions.contains(where: { $0.id == selectedSessionID }) {
                await select(sessionID: selectedSessionID)
            } else if let first = sessions.first, selectedSession == nil {
                await select(sessionID: first.id)
            }
        } catch {
            statusText = "Unable to reach local proxy API"
            sessions = []
            authorityInfo = nil
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
