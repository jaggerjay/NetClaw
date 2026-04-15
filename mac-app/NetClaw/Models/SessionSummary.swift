import Foundation

struct SessionSummary: Identifiable, Decodable {
    let id: String
    let startTime: Date
    let method: String
    let host: String
    let url: String
    let statusCode: Int
    let durationMs: Int
    let contentType: String
    let responseSize: Int64
    let hasError: Bool
}
