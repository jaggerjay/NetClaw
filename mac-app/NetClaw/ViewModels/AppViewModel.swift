import Foundation
import SwiftUI

@MainActor
final class AppViewModel: ObservableObject {
    @Published var sessions: [SessionSummary] = []
    @Published var selectedSession: SessionDetail?
    @Published var searchText: String = ""
    @Published var statusText: String = "Proxy not connected"

    private let apiClient = APIClient()

    func refresh() async {
        do {
            let items = try await apiClient.fetchSessions()
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

    var filteredSessions: [SessionSummary] {
        guard !searchText.isEmpty else { return sessions }
        return sessions.filter {
            $0.url.localizedCaseInsensitiveContains(searchText) ||
            $0.host.localizedCaseInsensitiveContains(searchText) ||
            $0.method.localizedCaseInsensitiveContains(searchText)
        }
    }
}
