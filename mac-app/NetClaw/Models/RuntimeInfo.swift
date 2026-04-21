import Foundation

struct RuntimeInfo: Decodable {
    let proxyListenAddress: String
    let apiListenAddress: String
    let dataDir: String
    let certificatePath: String

    var proxyHost: String {
        proxyListenAddress.split(separator: ":").dropLast().joined(separator: ":")
    }

    var proxyPort: String {
        proxyListenAddress.split(separator: ":").last.map(String.init) ?? ""
    }
}
