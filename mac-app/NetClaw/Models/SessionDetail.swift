import Foundation

struct SessionDetail: Identifiable, Decodable {
    let id: String
    let startTime: Date
    let endTime: Date
    let scheme: String
    let method: String
    let host: String
    let port: Int
    let path: String
    let url: String
    let statusCode: Int
    let durationMs: Int
    let clientAddress: String
    let requestHeaders: [String: String]
    let responseHeaders: [String: String]
    let requestBody: Data?
    let responseBody: Data?
    let requestSize: Int64
    let responseSize: Int64
    let contentType: String
    let error: String?
    let tlsIntercepted: Bool
    let captureMode: String
    let fallbackReason: String?
    let tunnelBytesUp: Int64
    let tunnelBytesDown: Int64
    let tunnelTargetAddress: String?
}
