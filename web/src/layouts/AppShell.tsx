import type { ReactNode } from "react";

import type { AuthSessionResponse } from "../lib/types";
import { SideNav } from "../components/SideNav";
import { TopBar } from "../components/TopBar";

type AppShellProps = {
  session: AuthSessionResponse;
  navItems: Array<{ label: string; to: string }>;
  children: ReactNode;
};

export function AppShell({ session, navItems, children }: AppShellProps) {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <div className="grid min-h-screen lg:grid-cols-[260px_1fr]">
        <SideNav items={navItems} workspaceMode={session.workspace?.mode} />
        <div className="flex min-h-screen flex-col">
          <TopBar session={session} />
          <main className="flex-1 px-6 py-8 lg:px-10">{children}</main>
        </div>
      </div>
    </div>
  );
}
