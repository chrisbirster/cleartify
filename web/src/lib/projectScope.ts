import { matchPath, useLocation, useNavigate } from "react-router-dom";

import type { Cluster, Project, Signal, WorkObject } from "./types";

function normalizeProjectId(value: string | null | undefined) {
  return value?.trim() ?? "";
}

export function filterSignalsByProject(signals: Signal[], projectId?: string) {
  const normalizedProjectId = normalizeProjectId(projectId);
  if (!normalizedProjectId) {
    return signals;
  }
  return signals.filter((signal) => normalizeProjectId(signal.projectId) === normalizedProjectId);
}

export function filterClustersByProject(clusters: Cluster[], projectId?: string) {
  const normalizedProjectId = normalizeProjectId(projectId);
  if (!normalizedProjectId) {
    return clusters;
  }
  return clusters.filter((cluster) => normalizeProjectId(cluster.projectId) === normalizedProjectId);
}

export function filterWorkObjectsByProject(workObjects: WorkObject[], projectId?: string) {
  const normalizedProjectId = normalizeProjectId(projectId);
  if (!normalizedProjectId) {
    return workObjects;
  }
  return workObjects.filter((workObject) => normalizeProjectId(workObject.projectId) === normalizedProjectId);
}

function buildProjectSearch(projectId?: string) {
  const normalizedProjectId = normalizeProjectId(projectId);
  if (!normalizedProjectId) {
    return "";
  }

  const params = new URLSearchParams({ projectId: normalizedProjectId });
  return `?${params.toString()}`;
}

export function useProjectScope(projects: Project[]) {
  const location = useLocation();
  const navigate = useNavigate();
  const projectRouteMatch = matchPath("/projects/:projectId", location.pathname);
  const searchParams = new URLSearchParams(location.search);
  const scopedProjectId = normalizeProjectId(projectRouteMatch?.params.projectId ?? searchParams.get("projectId"));
  const selectedProject = projects.find((project) => project.id === scopedProjectId);
  const selectedProjectId = selectedProject?.id ?? "";

  const linkTo = (path: string, projectId = selectedProjectId) => {
    if (path === "/projects") {
      return projectId ? `/projects/${encodeURIComponent(projectId)}` : "/projects";
    }
    return `${path}${buildProjectSearch(projectId)}`;
  };

  const setSelectedProjectId = (projectId: string) => {
    const normalizedProjectId = normalizeProjectId(projectId);
    if (location.pathname === "/projects" || projectRouteMatch) {
      navigate(normalizedProjectId ? `/projects/${encodeURIComponent(normalizedProjectId)}` : "/projects");
      return;
    }

    const nextSearchParams = new URLSearchParams(location.search);
    if (normalizedProjectId) {
      nextSearchParams.set("projectId", normalizedProjectId);
    } else {
      nextSearchParams.delete("projectId");
    }

    const nextSearch = nextSearchParams.toString();
    navigate({
      pathname: location.pathname,
      search: nextSearch ? `?${nextSearch}` : "",
      hash: location.hash,
    });
  };

  return {
    selectedProject,
    selectedProjectId,
    linkTo,
    setSelectedProjectId,
  };
}
