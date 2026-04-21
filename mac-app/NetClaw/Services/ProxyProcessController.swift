import Foundation

struct ProxyLaunchValidationResult {
    let success: Bool
    let message: String
}

final class ProxyProcessController {
    var onOutput: ((String) -> Void)?
    var onStateChange: ((Bool, String) -> Void)?

    private var process: Process?
    private var outputPipe: Pipe?

    var isRunning: Bool {
        process?.isRunning == true
    }

    func validateLaunch(command: String, workingDirectory: String?) -> ProxyLaunchValidationResult {
        let trimmedCommand = command.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmedCommand.isEmpty {
            return .init(success: false, message: "Launch command is empty")
        }

        let trimmedDirectory = workingDirectory?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        if trimmedDirectory.isEmpty {
            return .init(success: false, message: "Choose a working directory first")
        }

        var isDirectory: ObjCBool = false
        let exists = FileManager.default.fileExists(atPath: trimmedDirectory, isDirectory: &isDirectory)
        if !exists || !isDirectory.boolValue {
            return .init(success: false, message: "Working directory does not exist")
        }

        if trimmedCommand.contains("go ") || trimmedCommand.hasPrefix("go") {
            let goCheck = Process()
            goCheck.executableURL = URL(fileURLWithPath: "/bin/bash")
            goCheck.arguments = ["-lc", "command -v go >/dev/null 2>&1"]
            goCheck.currentDirectoryURL = URL(fileURLWithPath: trimmedDirectory, isDirectory: true)
            do {
                try goCheck.run()
                goCheck.waitUntilExit()
                if goCheck.terminationStatus != 0 {
                    return .init(success: false, message: "Go toolchain not found in PATH for this app")
                }
            } catch {
                return .init(success: false, message: "Unable to validate Go toolchain: \(error.localizedDescription)")
            }
        }

        return .init(success: true, message: "Launch settings look valid")
    }

    func start(command: String, workingDirectory: String?) throws {
        guard !isRunning else {
            throw ProxyProcessError.alreadyRunning
        }

        let validation = validateLaunch(command: command, workingDirectory: workingDirectory)
        guard validation.success else {
            throw ProxyProcessError.validationFailed(validation.message)
        }

        let trimmedCommand = command.trimmingCharacters(in: .whitespacesAndNewlines)
        let trimmedWorkingDirectory = workingDirectory?.trimmingCharacters(in: .whitespacesAndNewlines)

        let proc = Process()
        proc.executableURL = URL(fileURLWithPath: "/bin/bash")
        proc.arguments = ["-lc", trimmedCommand]

        if let trimmedWorkingDirectory, !trimmedWorkingDirectory.isEmpty {
            proc.currentDirectoryURL = URL(fileURLWithPath: trimmedWorkingDirectory, isDirectory: true)
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
                    reason = process.terminationStatus == 0 ? "Proxy stopped cleanly" : "Proxy stopped (exit \(process.terminationStatus))"
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
    case validationFailed(String)

    var errorDescription: String? {
        switch self {
        case .alreadyRunning:
            return "Proxy process is already running"
        case .emptyCommand:
            return "Enter a proxy launch command first"
        case .validationFailed(let message):
            return message
        }
    }
}
