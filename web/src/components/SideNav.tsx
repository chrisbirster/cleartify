import type { WorkspaceMode } from "../lib/types";
import type { Project } from "../lib/types";
import { NavLink } from "react-router-dom";

import { useProjectScope } from "../lib/projectScope";

type NavItem = {
  label: string;
  to: string;
};

type SideNavProps = {
  items: NavItem[];
  projects?: Project[];
  workspaceMode?: WorkspaceMode;
};

export function SideNav({ items, projects = [], workspaceMode }: SideNavProps) {
  const isSampleWorkspace = workspaceMode === "sample";
  const { linkTo } = useProjectScope(projects);

  return (
    <aside className="border-r border-border bg-card/80 px-4 py-6">
      <div className="mb-8 px-3">
        <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Workspace</p>
        <h1 className="mt-2 text-xl font-semibold text-foreground">Cleartify</h1>
        {isSampleWorkspace ? (
          <p className="mt-2 text-xs font-semibold uppercase tracking-[0.18em] text-amber-200">Sample workspace</p>
        ) : null}
        <p className="mt-2 text-xs text-muted-foreground">
          {projects.length} {projects.length === 1 ? "Project" : "Projects"}
        </p>
      </div>
      <nav className="space-y-1">
        {items.map((item) => (
          <NavLink
            key={item.to}
            to={linkTo(item.to)}
            className={({ isActive }) =>
              [
                "block rounded-2xl px-3 py-2 text-sm transition-colors",
                isActive ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:bg-muted hover:text-foreground",
              ].join(" ")
            }
            end={item.to === "/"}
          >
            {item.label}
          </NavLink>
        ))}
      </nav>
    </aside>
  );
}
