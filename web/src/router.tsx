import { BrowserRouter, MemoryRouter, Outlet, Route, Routes } from "react-router-dom";

import type { AppData, WorkObject } from "./lib/types";
import { AppShell } from "./layouts/AppShell";
import { AgentWorkspacePage } from "./pages/AgentWorkspacePage";
import { ClusterDetailPage } from "./pages/ClusterDetailPage";
import { InboxPage } from "./pages/InboxPage";
import { PlanPage } from "./pages/PlanPage";
import { PrototypeNewPage } from "./pages/PrototypeNewPage";
import { PrototypesPage } from "./pages/PrototypesPage";
import { ReleaseDetailPage } from "./pages/ReleaseDetailPage";
import { ReleasesPage } from "./pages/ReleasesPage";
import { ReviewPage } from "./pages/ReviewPage";
import { TodayPage } from "./pages/TodayPage";
import { WorkObjectDetailPage } from "./pages/WorkObjectDetailPage";
import { WorkObjectsPage } from "./pages/WorkObjectsPage";

export const NAV_ITEMS = [
  { label: "Today", to: "/" },
  { label: "Context Inbox", to: "/inbox" },
  { label: "Work Objects", to: "/work" },
  { label: "Prototypes", to: "/prototypes" },
  { label: "Releases", to: "/releases" },
];

type AppRoutesProps = {
  data: AppData;
  onWorkObjectUpdated?: (workObject: WorkObject) => void;
};

function ShellRoute({ data }: AppRoutesProps) {
  return (
    <AppShell session={data.session} navItems={NAV_ITEMS}>
      <Outlet />
    </AppShell>
  );
}

export function AppRoutes({ data, onWorkObjectUpdated }: AppRoutesProps) {
  return (
    <Routes>
      <Route path="/" element={<ShellRoute data={data} />}>
        <Route index element={<TodayPage signals={data.signals} clusters={data.clusters} workObjects={data.workObjects} />} />
        <Route path="inbox" element={<InboxPage clusters={data.clusters} />} />
        <Route path="inbox/:clusterId" element={<ClusterDetailPage clusters={data.clusters} />} />
        <Route path="work" element={<WorkObjectsPage workObjects={data.workObjects} clusters={data.clusters} />} />
        <Route
          path="work/:workObjectId"
          element={<WorkObjectDetailPage workObjects={data.workObjects} clusters={data.clusters} onWorkObjectUpdated={onWorkObjectUpdated} />}
        />
        <Route path="prototypes" element={<PrototypesPage workObjects={data.workObjects} />} />
        <Route path="prototypes/new" element={<PrototypeNewPage />} />
        <Route path="plans/:workObjectId" element={<PlanPage workObjects={data.workObjects} clusters={data.clusters} />} />
        <Route
          path="agents/:workObjectId"
          element={<AgentWorkspacePage workObjects={data.workObjects} clusters={data.clusters} onWorkObjectUpdated={onWorkObjectUpdated} />}
        />
        <Route path="review/:workObjectId" element={<ReviewPage />} />
        <Route path="releases" element={<ReleasesPage />} />
        <Route path="releases/:releaseId" element={<ReleaseDetailPage />} />
      </Route>
    </Routes>
  );
}

export function AppRouter({ data, onWorkObjectUpdated }: AppRoutesProps) {
  return (
    <BrowserRouter basename="/app">
      <AppRoutes data={data} onWorkObjectUpdated={onWorkObjectUpdated} />
    </BrowserRouter>
  );
}

export function AppMemoryRouter({ data, onWorkObjectUpdated, initialEntries = ["/"] }: AppRoutesProps & { initialEntries?: string[] }) {
  return (
    <MemoryRouter initialEntries={initialEntries}>
      <AppRoutes data={data} onWorkObjectUpdated={onWorkObjectUpdated} />
    </MemoryRouter>
  );
}
