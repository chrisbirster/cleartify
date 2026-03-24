import { Link } from "react-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { HeaderMenuItem } from "./header-menu-item";
import { useAppRepository } from "#/lib/app-repository";
import { appNavigation } from "#/lib/app-navigation";

export function HeaderMenuLinks() {
    const repository = useAppRepository();
    const queryClient = useQueryClient();
    const { data: session } = useQuery({
        queryKey: ["auth-session"],
        queryFn: () => repository.fetchSession(),
    });

    const logoutMutation = useMutation({
        mutationFn: () => repository.logout(),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["auth-session"] });
            queryClient.invalidateQueries({ queryKey: ["entitlements"] });
        },
    });

    const userLabel = session?.user?.displayName || session?.user?.email || "User";
    const userInitial = userLabel.trim().charAt(0).toUpperCase() || "U";

    return (
        <div className="flex w-full flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <Link to="/decks" className="flex items-center shrink-0">
                <div>
                    <h1 className="text-2xl font-bold text-[var(--app-text)]">Cleartify</h1>
                    <p className="text-xs uppercase tracking-[0.2em] text-[var(--app-muted)]">
                        AI governance system of record
                    </p>
                </div>
            </Link>
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
                <div className="flex gap-1 overflow-x-auto pb-1 sm:overflow-visible sm:pb-0">
                    {appNavigation.map((item) => (
                        <HeaderMenuItem key={item.to} item={{ path: item.to, label: item.label }} />
                    ))}
                </div>
                <div className="flex items-center gap-2">
                    {session?.entitlements?.plan && (
                        <span className="inline-flex items-center rounded-full border border-[var(--app-line)] bg-[var(--app-card)] px-3 py-1 text-xs font-medium text-[var(--app-text-soft)]">
                            {session.entitlements.plan.toUpperCase()}
                        </span>
                    )}
                    {session?.authenticated ? (
                        <>
                            <span className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-[var(--app-accent)] text-sm font-semibold text-[var(--app-accent-ink)]">
                                {userInitial}
                            </span>
                            <div className="hidden min-w-0 sm:block">
                                <p className="truncate text-sm font-medium text-[var(--app-text)]">{userLabel}</p>
                                <p className="text-xs uppercase tracking-[0.24em] text-[var(--app-muted)]">Signed in</p>
                            </div>
                            <button
                                type="button"
                                onClick={() => logoutMutation.mutate()}
                                className="rounded-md px-3 py-2 text-sm text-[var(--app-text-soft)] hover:bg-[var(--app-muted-surface)]"
                            >
                                Sign Out
                            </button>
                        </>
                    ) : (
                        <Link
                            to="/login"
                            className="rounded-md bg-[var(--app-accent)] px-3 py-2 text-sm font-medium text-[var(--app-accent-ink)]"
                        >
                            Log In
                        </Link>
                    )}
                </div>
            </div>
        </div>
    );
}
