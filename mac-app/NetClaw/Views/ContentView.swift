import SwiftUI

struct ContentView: View {
    @StateObject private var viewModel = AppViewModel()

    var body: some View {
        NavigationSplitView {
            VStack(spacing: 0) {
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

                SessionListView(sessions: viewModel.sessions) { session in
                    Task { await viewModel.select(sessionID: session.id) }
                }
            }
            .navigationTitle("NetClaw")
        } detail: {
            SessionDetailView(session: viewModel.selectedSession)
        }
        .task {
            await viewModel.refresh()
        }
        .onSubmit {
            Task { await viewModel.refresh() }
        }
        .safeAreaInset(edge: .bottom) {
            HStack {
                Circle()
                    .fill(viewModel.statusText == "Connected" ? Color.green : Color.orange)
                    .frame(width: 8, height: 8)
                Text(viewModel.statusText)
                    .font(.caption)
                Spacer()
                Text("\(viewModel.sessions.count) session(s)")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
            .padding()
            .background(.bar)
        }
    }
}
