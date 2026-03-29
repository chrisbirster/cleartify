import { useEffect, useState } from "react";
import { Effect } from "effect";

import { loadAppData } from "./lib/api";
import type { AppData } from "./lib/types";
import { AppRouter } from "./router";

type AppState =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "ready"; data: AppData };

export function App() {
  const [state, setState] = useState<AppState>({ status: "loading" });

  useEffect(() => {
    let isActive = true;

    Effect.runPromise(loadAppData())
      .then((data) => {
        if (isActive) {
          setState({ status: "ready", data });
        }
      })
      .catch(() => {
        if (isActive) {
          setState({ status: "error", message: "Unable to load app data." });
        }
      });

    return () => {
      isActive = false;
    };
  }, []);

  if (state.status === "loading") {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background text-foreground">
        <p role="status" className="text-sm text-muted-foreground">
          Loading workspace
        </p>
      </div>
    );
  }

  if (state.status === "error") {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background px-6 text-foreground">
        <div className="max-w-md rounded-3xl border border-border bg-card p-8 text-center">
          <h1 className="text-2xl font-semibold">Unable to load Cleartify</h1>
          <p className="mt-3 text-sm text-muted-foreground">{state.message}</p>
        </div>
      </div>
    );
  }

  if (!state.data.session.authenticated) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background px-6 text-foreground">
        <div className="max-w-md rounded-3xl border border-border bg-card p-8 text-center">
          <h1 className="text-2xl font-semibold">Authentication required</h1>
          <p className="mt-3 text-sm text-muted-foreground">
            WorkOS AuthKit is reserved behind the Phase 0 auth boundary. Use the demo session locally or finish auth wiring next.
          </p>
        </div>
      </div>
    );
  }

  return <AppRouter data={state.data} />;
}

export default App;
