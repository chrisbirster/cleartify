import { BrowserRouter, MemoryRouter, Navigate, Outlet, Route, Routes, useLocation } from "react-router-dom";

import type { AppData, WorkObject } from "./lib/types";
import { AppShell } from "./layouts/AppShell";
import { AgentWorkspacePage } from "./pages/AgentWorkspacePage";
import { ArtifactEditorPage } from "./pages/ArtifactEditorPage";
import { ArtifactsPage } from "./pages/ArtifactsPage";
import { ClusterDetailPage } from "./pages/ClusterDetailPage";
import { DecisionEditorPage } from "./pages/DecisionEditorPage";
import { DecisionsPage } from "./pages/DecisionsPage";
import { InboxPage } from "./pages/InboxPage";
import { MePage } from "./pages/MePage";
import { PlanPage } from "./pages/PlanPage";
import { PrototypeNewPage } from "./pages/PrototypeNewPage";
import { PrototypesPage } from "./pages/PrototypesPage";
import { ReleaseDetailPage } from "./pages/ReleaseDetailPage";
import { ReleasesPage } from "./pages/ReleasesPage";
import { ReviewPage } from "./pages/ReviewPage";
import { SignalEditorPage } from "./pages/SignalEditorPage";
import { SignalsPage } from "./pages/SignalsPage";
import { TeamPage } from "./pages/TeamPage";
import { TodayPage } from "./pages/TodayPage";
import { WorkObjectDetailPage } from "./pages/WorkObjectDetailPage";
import { WorkObjectsPage } from "./pages/WorkObjectsPage";

export const NAV_ITEMS = [
  { label: "Today", to: "/" },
  { label: "Context Inbox", to: "/inbox" },
  { label: "Work Objects", to: "/work" },
  { label: "Signals", to: "/signals" },
  { label: "Decisions", to: "/decisions" },
  { label: "Artifacts", to: "/artifacts" },
  { label: "Prototypes", to: "/prototypes" },
  { label: "Releases", to: "/releases" },
];

type AppRoutesProps = {
  data: AppData;
  onWorkObjectUpdated?: (workObject: WorkObject) => void;
  onDataChanged?: () => Promise<unknown>;
};

function ShellRoute({ data }: AppRoutesProps) {
  return (
    <AppShell session={data.session} navItems={NAV_ITEMS}>
      <Outlet />
    </AppShell>
  );
}

export function AppRoutes({ data, onWorkObjectUpdated, onDataChanged }: AppRoutesProps) {
  return (
    <Routes>
      <Route path="/app/*" element={<LegacyAppRedirect />} />
      <Route path="/" element={<ShellRoute data={data} />}>
        <Route
          index
          element={
            <TodayPage
              signals={data.signals}
              clusters={data.clusters}
              workObjects={data.workObjects}
              workspaceMode={data.session.workspace?.mode}
            />
          }
        />
        <Route
          path="inbox"
          element={<InboxPage clusters={data.clusters} workspaceMode={data.session.workspace?.mode} onDataChanged={onDataChanged} />}
        />
        <Route path="inbox/:clusterId" element={<ClusterDetailPage clusters={data.clusters} onDataChanged={onDataChanged} />} />
        <Route
          path="work"
          element={<WorkObjectsPage workObjects={data.workObjects} clusters={data.clusters} workspaceMode={data.session.workspace?.mode} />}
        />
        <Route
          path="work/:workObjectId"
          element={<WorkObjectDetailPage workObjects={data.workObjects} clusters={data.clusters} onWorkObjectUpdated={onWorkObjectUpdated} />}
        />
        <Route path="me" element={<MePage />} />
        <Route path="team" element={<TeamPage />} />
        <Route path="signals" element={<SignalsPage />} />
        <Route path="signals/new" element={<SignalEditorPage mode="create" onDataChanged={onDataChanged} />} />
        <Route path="signals/:signalId" element={<SignalEditorPage mode="edit" onDataChanged={onDataChanged} />} />
        <Route path="decisions" element={<DecisionsPage workObjects={data.workObjects} />} />
        <Route
          path="decisions/new"
          element={<DecisionEditorPage mode="create" workObjects={data.workObjects} onDataChanged={onDataChanged} />}
        />
        <Route
          path="decisions/:decisionId"
          element={<DecisionEditorPage mode="edit" workObjects={data.workObjects} onDataChanged={onDataChanged} />}
        />
        <Route path="artifacts" element={<ArtifactsPage workObjects={data.workObjects} />} />
        <Route
          path="artifacts/new"
          element={<ArtifactEditorPage mode="create" workObjects={data.workObjects} onDataChanged={onDataChanged} />}
        />
        <Route
          path="artifacts/:artifactId"
          element={<ArtifactEditorPage mode="edit" workObjects={data.workObjects} onDataChanged={onDataChanged} />}
        />
        <Route path="prototypes" element={<PrototypesPage workObjects={data.workObjects} />} />
        <Route path="prototypes/new" element={<PrototypeNewPage />} />
        <Route path="plans/:workObjectId" element={<PlanPage workObjects={data.workObjects} clusters={data.clusters} />} />
        <Route
          path="agents/:workObjectId"
          element={<AgentWorkspacePage workObjects={data.workObjects} clusters={data.clusters} onWorkObjectUpdated={onWorkObjectUpdated} />}
        />
        <Route
          path="review/:workObjectId"
          element={<ReviewPage workObjects={data.workObjects} clusters={data.clusters} onWorkObjectUpdated={onWorkObjectUpdated} />}
        />
        <Route path="releases" element={<ReleasesPage workObjects={data.workObjects} clusters={data.clusters} />} />
        <Route path="releases/:releaseId" element={<ReleaseDetailPage />} />
      </Route>
    </Routes>
  );
}

function LegacyAppRedirect() {
  const location = useLocation();
  const nextPath = location.pathname.replace(/^\/app(?=\/|$)/, "") || "/";

  return <Navigate replace to={`${nextPath}${location.search}${location.hash}`} />;
}

export function AppRouter({ data, onWorkObjectUpdated, onDataChanged }: AppRoutesProps) {
  return (
    <BrowserRouter>
      <AppRoutes data={data} onWorkObjectUpdated={onWorkObjectUpdated} onDataChanged={onDataChanged} />
    </BrowserRouter>
  );
}

export function AppMemoryRouter({
  data,
  onWorkObjectUpdated,
  onDataChanged,
  initialEntries = ["/"],
}: AppRoutesProps & { initialEntries?: string[] }) {
  return (
    <MemoryRouter initialEntries={initialEntries}>
      <AppRoutes data={data} onWorkObjectUpdated={onWorkObjectUpdated} onDataChanged={onDataChanged} />
    </MemoryRouter>
  );
}
