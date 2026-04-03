import { Link } from "react-router-dom";

import type { AuthSessionResponse } from "../lib/types";

type TopBarProps = {
  session: AuthSessionResponse;
};

export function TopBar({ session }: TopBarProps) {
  return (
    <header className="flex items-center justify-between border-b border-border bg-black/20 px-6 py-4 backdrop-blur-sm">
      <div>
        <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Cleartify</p>
        <p className="text-sm font-medium text-foreground">AI-native product execution system</p>
      </div>
      <div className="text-right">
        <p className="text-xs text-muted-foreground">
          {session.authenticated ? (
            <Link className="transition hover:text-foreground" to="/team">
              {session.workspace?.name ?? session.organization?.name ?? "Current session"}
            </Link>
          ) : (
            session.workspace?.name ?? session.organization?.name ?? "Current session"
          )}
        </p>
        <p className="text-sm font-medium text-foreground">
          {session.authenticated ? (
            <Link className="transition hover:text-primary" to="/me">
              {session.user?.displayName ?? "Signed out"}
            </Link>
          ) : (
            session.user?.displayName ?? "Signed out"
          )}
        </p>
        {session.authenticated ? (
          <a href="/logout" className="mt-1 inline-block text-xs text-muted-foreground transition hover:text-foreground">
            Sign out
          </a>
        ) : null}
      </div>
    </header>
  );
}
