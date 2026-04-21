import Foundation

final class ProxyProcessController {
    var onOutput: ((String) -> Void)?
    var onStateChange: ((Bool, String) -> Void)?

    private var process: Process?
    private var outputPipe: Pipe?

    var isRunning: Bool {
        process?.isRunning == true
    }

    func start(command: String, workingDirectory: String?) throws {
        guard !isRunning else {
            throw ProxyProcessError.alreadyRunning
        }

        let trimmedCommand = command.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmedCommand.isEmpty else {
            throw ProxyProcessError.emptyCommand
        }

        let proc = Process()
        proc.executableURL = URL(fileURLWithPath: "/bin/bash")
        proc.arguments = ["-lc", trimmedCommand]

        if let workingDirectory, !workingDirectory.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            proc.currentDirectoryURL = URL(fileURLWithPath: workingDirectory, isDirectory: true)
        }

        let pipe = Pipe()
        proc.standardOutput = pipe
        proc.standardError = pipe

        pipe.fileHandleForReading.readabilityHandler = { [weak self] handle in
            let data = handle.availableData
            guard !data.isEmpty else { return }
            let text = String(data: data, encoding: .utf8) ?? data.base64EncodedString()
            DispatchQueue.main.async {
                self?.onOutput?(text)
            }
        }

        proc.terminationHandler = { [weak self] process in
            DispatchQueue.main.async {
                self?.outputPipe?.fileHandleForReading.readabilityHandler = nil
                self?.outputPipe = nil
                self?.process = nil

                let reason: String
                switch process.terminationReason {
                case .exit:
                    reason = "Proxy stopped (exit \(process.terminationStatus))"
                case .uncaughtSignal:
                    reason = "Proxy stopped by signal \(process.terminationStatus)"
                @unknown default:
                    reason = "Proxy stopped"
                }
                self?.onStateChange?(false, reason)
            }
        }

        process = proc
        outputPipe = pipe
        try proc.run()
        onStateChange?(true, "Proxy process running")
    }

    func stop() {
        guard let process else { return }
        if process.isRunning {
            process.terminate()
            onStateChange?(false, "Stopping proxy process…")
        }
    }
}

enum ProxyProcessError: LocalizedError {
    case alreadyRunning
    case emptyCommand

    var errorDescription: String? {
        switch self {
        case .alreadyRunning:
            return "Proxy process is already running"
        case .emptyCommand:
            return "Enter a proxy launch command first"
        }
    }
}
