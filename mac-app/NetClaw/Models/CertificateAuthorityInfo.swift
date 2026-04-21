import Foundation

struct CertificateAuthorityInfo: Decodable {
    let storageDir: String
    let certificatePath: String
    let privateKeyPath: String
    let commonName: String
    let trusted: Bool
}

struct HealthStatus: Decodable {
    let ok: Bool
}
