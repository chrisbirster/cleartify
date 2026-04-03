import { useEffect, useState } from "react";
import { Effect } from "effect";

import { completeOnboarding, loadAppData } from "./lib/api";
import type { AppData, WorkObject } from "./lib/types";
import { AppRouter } from "./router";
import { OnboardingWalkthrough } from "./components/OnboardingWalkthrough";

type AppState =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "ready"; data: AppData };

export function App() {
  const [state, setState] = useState<AppState>({ status: "loading" });
  const [onboardingState, setOnboardingState] = useState<{ pending: boolean; error?: string }>({ pending: false });

  const refreshData = () =>
    Effect.runPromise(loadAppData()).then((data) => {
      setState({ status: "ready", data });
      return data;
    });

  useEffect(() => {
    let isActive = true;

    refreshData()
      .then((data) => {
        if (!isActive) {
          return;
        }
        setState({ status: "ready", data });
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

  const handleWorkObjectUpdated = (updatedWorkObject: WorkObject) => {
    setState((currentState) => {
      if (currentState.status !== "ready") {
        return currentState;
      }

      return {
        status: "ready",
        data: {
          ...currentState.data,
          workObjects: currentState.data.workObjects.map((entry) =>
            entry.id === updatedWorkObject.id ? updatedWorkObject : entry,
          ),
        },
      };
    });
  };

  const handleOnboardingCompleted = () => {
    setOnboardingState({ pending: true });

    Effect.runPromise(completeOnboarding())
      .then((session) => {
        setState((currentState) => {
          if (currentState.status !== "ready") {
            return currentState;
          }

          return {
            status: "ready",
            data: {
              ...currentState.data,
              session,
            },
          };
        });
        setOnboardingState({ pending: false });
      })
      .catch(() => {
        setOnboardingState({ pending: false, error: "Unable to finish onboarding right now." });
      });
  };

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
    const loginHref = buildLoginHref("sign-in");
    const signupHref = buildLoginHref("sign-up");
    const authError = getAuthErrorMessage();

    return (
      <div className="flex min-h-screen items-center justify-center bg-background px-6 text-foreground">
        <div className="max-w-md rounded-3xl border border-border bg-card p-8 text-center shadow-[0_24px_80px_rgba(0,0,0,0.35)]">
          <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Sign in</p>
          <h1 className="mt-3 text-2xl font-semibold">Continue into Cleartify</h1>
          <p className="mt-3 text-sm leading-6 text-muted-foreground">
            Use hosted WorkOS AuthKit to enter the product workspace, then complete the first-run walkthrough before the
            app shell opens.
          </p>
          {authError ? <p className="mt-4 text-sm text-red-300">{authError}</p> : null}
          <div className="mt-6 flex flex-col gap-3">
            <a
              href={loginHref}
              className="rounded-full bg-foreground px-5 py-2 text-sm font-medium text-background transition hover:bg-foreground/90"
            >
              Sign in with WorkOS
            </a>
            <a
              href={signupHref}
              className="rounded-full border border-border px-5 py-2 text-sm font-medium text-foreground transition hover:border-border/70 hover:bg-card/60"
            >
              Create an account
            </a>
          </div>
        </div>
      </div>
    );
  }

  if (state.data.session.needsOnboarding) {
    return (
      <OnboardingWalkthrough
        workspaceName={state.data.session.workspace?.name ?? state.data.session.organization?.name}
        pending={onboardingState.pending}
        errorMessage={onboardingState.error}
        onComplete={handleOnboardingCompleted}
      />
    );
  }

  return <AppRouter data={state.data} onWorkObjectUpdated={handleWorkObjectUpdated} onDataChanged={refreshData} />;
}

function buildLoginHref(screenHint: "sign-in" | "sign-up") {
  const returnTo = `${window.location.pathname}${window.location.hash}`;
  const params = new URLSearchParams({
    screenHint,
    returnTo,
  });
  return `/login?${params.toString()}`;
}

function getAuthErrorMessage() {
  const authError = new URLSearchParams(window.location.search).get("authError");
  if (!authError) {
    return "";
  }
  switch (authError) {
    case "auth_failed":
      return "Sign-in did not complete successfully. Try again.";
    case "auth_unavailable":
      return "Hosted sign-in is not configured for this environment yet.";
    default:
      return "Sign-in is currently unavailable.";
  }
}

export default App;
