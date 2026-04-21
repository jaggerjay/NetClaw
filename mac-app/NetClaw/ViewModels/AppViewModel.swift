import Foundation
import SwiftUI

@MainActor
final class AppViewModel: ObservableObject {
    @Published var sessions: [SessionSummary] = []
    @Published var selectedSession: SessionDetail?
    @Published var searchText: String = ""
    @Published var methodFilter: String = ""
    @Published var hostFilter: String = ""
    @Published var hasErrorOnly: Bool = false
    @Published var interceptedOnly: Bool = false
    @Published var statusText: String = "Proxy not connected"

    private let apiClient = APIClient()

    func refresh() async {
        do {
            let items = try await apiClient.fetchSessions(query: currentQuery)
            sessions = items
            statusText = "Connected"
        } catch {
            statusText = "Unable to reach local proxy API"
        }
    }

    func select(sessionID: String) async {
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
