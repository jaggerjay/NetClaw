import SwiftUI

struct ContentView: View {
    @StateObject private var viewModel = AppViewModel()

    var body: some View {
        NavigationSplitView {
            VStack(spacing: 0) {
                ScrollView {
                    VStack(spacing: 12) {
                        ProxyRuntimeView(
                            workingDirectoryText: $viewModel.proxyWorkingDirectoryText,
                            commandText: $viewModel.proxyCommandText,
                            proxyStatusText: viewModel.proxyStatusText,
                            proxyValidationText: viewModel.proxyValidationText,
                            isProxyRunning: viewModel.isProxyRunning,
                            logText: viewModel.proxyLogText,
                            onStart: { viewModel.startProxy() },
                            onStop: { viewModel.stopProxy() },
                            onClearLog: { viewModel.clearProxyLog() },
                            onValidate: { viewModel.validateProxyLaunchSettings() },
                            onUseSuggestedCommand: { viewModel.useSuggestedProxyCommand() },
                            onUseDebugBuildCommand: { viewModel.useDebugBuildCommand() },
                            onUseRepoRootSuggestion: { viewModel.useRepoRootSuggestion() },
                            onWorkingDirectorySelected: { path in viewModel.updateWorkingDirectory(path) }
                        )

                        ConnectionStatusView(
                            apiBaseURLText: $viewModel.apiBaseURLText,
                            statusText: viewModel.statusText,
                            isConnected: viewModel.isConnected,
                            isRefreshing: viewModel.isRefreshing,
                            autoRefreshEnabled: viewModel.autoRefreshEnabled,
                            authorityInfo: viewModel.authorityInfo,
                            lastErrorText: viewModel.lastErrorText,
                            onRefresh: { Task { await viewModel.refresh() } },
                            onQuickCheck: { Task { await viewModel.quickHealthCheck() } },
                            onApplyBaseURL: { Task { await viewModel.applyAPIBaseURL() } },
                            onToggleAutoRefresh: { value in viewModel.setAutoRefreshEnabled(value) }
                        )

                        SetupGuideView(
                            runtimeInfo: viewModel.runtimeInfo,
                            authorityInfo: viewModel.authorityInfo
                        )
                    }
                    .padding()
                }

                Divider()

                VStack(spacing: 8) {
                    HStack {
                        TextField("Search host / URL / method", text: $viewModel.searchText)
                            .textFieldStyle(.roundedBorder)
                        Button("Refresh") {
                            Task { await viewModel.refresh() }
                        }
                    }

                    HStack {
                        TextField("Method", text: $viewModel.methodFilter)
                            .textFieldStyle(.roundedBorder)
                        TextField("Host", text: $viewModel.hostFilter)
                            .textFieldStyle(.roundedBorder)
                    }

                    HStack {
                        Toggle("Errors only", isOn: $viewModel.hasErrorOnly)
                        Toggle("MITM only", isOn: $viewModel.interceptedOnly)
                        Spacer()
                        Button("Clear") {
                            viewModel.clearFilters()
                            Task { await viewModel.refresh() }
                        }
                    }
                    .font(.caption)
                }
                .padding()

                SessionListView(
                    sessions: viewModel.sessions,
                    selectedSessionID: viewModel.selectedSessionID
                ) { session in
                    Task { await viewModel.select(sessionID: session.id) }
                }
            }
            .navigationTitle("NetClaw")
            .frame(minWidth: 420)
        } detail: {
            SessionDetailView(session: viewModel.selectedSession)
        }
        .navigationSplitViewStyle(.balanced)
        .task {
            viewModel.start()
            await viewModel.applyAPIBaseURL()
        }
        .onSubmit {
            Task { await viewModel.refresh() }
        }
        .safeAreaInset(edge: .bottom) {
            HStack {
                Circle()
                    .fill(viewModel.isConnected ? Color.green : Color.orange)
                    .frame(width: 8, height: 8)
                Text(viewModel.statusText)
                    .font(.caption)
                Spacer()
                if viewModel.isProxyRunning {
                    Text("proxy running")
                        .font(.caption)
                        .foregroundStyle(.green)
                }
                Text("\(viewModel.sessions.count) session(s)")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
            .padding()
            .background(.bar)
        }
    }
}
