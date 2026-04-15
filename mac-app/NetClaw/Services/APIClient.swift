import Foundation

final class APIClient {
    private let decoder: JSONDecoder = {
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        return decoder
    }()

    private let baseURL = URL(string: "http://127.0.0.1:9091")!

    func fetchSessions() async throws -> [SessionSummary] {
        let (data, _) = try await URLSession.shared.data(from: baseURL.appendingPathComponent("/api/sessions"))
        return try decoder.decode([SessionSummary].self, from: data)
    }

    func fetchSession(id: String) async throws -> SessionDetail {
        let (data, _) = try await URLSession.shared.data(from: baseURL.appendingPathComponent("/api/sessions/\(id)"))
        return try decoder.decode(SessionDetail.self, from: data)
    }
}
