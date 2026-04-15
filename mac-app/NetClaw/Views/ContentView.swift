import SwiftUI

struct ContentView: View {
    @StateObject private var viewModel = AppViewModel()

    var body: some View {
        NavigationSplitView {
            VStack(spacing: 0) {
                HStack {
                    TextField("Search host / URL / method", text: $viewModel.searchText)
                    Button("Refresh") {
                        Task { await viewModel.refresh() }
                    }
                }
                .padding()

                SessionListView(sessions: viewModel.filteredSessions) { session in
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
        .safeAreaInset(edge: .bottom) {
            HStack {
                Circle()
                    .fill(viewModel.statusText == "Connected" ? Color.green : Color.orange)
                    .frame(width: 8, height: 8)
                Text(viewModel.statusText)
                    .font(.caption)
                Spacer()
            }
            .padding()
            .background(.bar)
        }
    }
}
