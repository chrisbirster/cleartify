import type { ReactNode } from "react";

import type { AuthSessionResponse, Project } from "../lib/types";
import { SideNav } from "../components/SideNav";
import { TopBar } from "../components/TopBar";

type AppShellProps = {
  session: AuthSessionResponse;
  projects?: Project[];
  navItems: Array<{ label: string; to: string }>;
  children: ReactNode;
};

export function AppShell({ session, projects = [], navItems, children }: AppShellProps) {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <div className="grid min-h-screen lg:grid-cols-[260px_1fr]">
        <SideNav items={navItems} projects={projects} workspaceMode={session.workspace?.mode} />
        <div className="flex min-h-screen flex-col">
          <TopBar projects={projects} session={session} />
          <main className="flex-1 px-6 py-8 lg:px-10">{children}</main>
        </div>
      </div>
    </div>
  );
}
