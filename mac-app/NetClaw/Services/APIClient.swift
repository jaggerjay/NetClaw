import Foundation

struct SessionQuery {
    var text: String = ""
    var method: String = ""
    var host: String = ""
    var hasError: Bool? = nil
    var tlsIntercepted: Bool? = nil
    var limit: Int? = nil

    var isEmpty: Bool {
        text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty &&
        method.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty &&
        host.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty &&
        hasError == nil &&
        tlsIntercepted == nil &&
        limit == nil
    }
}

final class APIClient {
    private let decoder: JSONDecoder = {
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        return decoder
    }()

    private let baseURL = URL(string: "http://127.0.0.1:9091")!

    func fetchSessions(query: SessionQuery = SessionQuery()) async throws -> [SessionSummary] {
        var components = URLComponents(url: baseURL.appendingPathComponent("/api/sessions"), resolvingAgainstBaseURL: false)!
        var items: [URLQueryItem] = []

        if !query.text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            items.append(URLQueryItem(name: "q", value: query.text))
        }
        if !query.method.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            items.append(URLQueryItem(name: "method", value: query.method))
        }
        if !query.host.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            items.append(URLQueryItem(name: "host", value: query.host))
        }
        if let hasError = query.hasError {
            items.append(URLQueryItem(name: "has_error", value: String(hasError)))
        }
        if let tlsIntercepted = query.tlsIntercepted {
            items.append(URLQueryItem(name: "tls_intercepted", value: String(tlsIntercepted)))
        }
        if let limit = query.limit {
            items.append(URLQueryItem(name: "limit", value: String(limit)))
        }
        if !items.isEmpty {
            components.queryItems = items
        }

        let (data, _) = try await URLSession.shared.data(from: components.url!)
        return try decoder.decode([SessionSummary].self, from: data)
    }

    func fetchSession(id: String) async throws -> SessionDetail {
        let (data, _) = try await URLSession.shared.data(from: baseURL.appendingPathComponent("/api/sessions/\(id)"))
        return try decoder.decode(SessionDetail.self, from: data)
    }
}
