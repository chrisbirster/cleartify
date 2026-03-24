export interface AppNavigationItem {
  to: string;
  label: string;
  description: string;
}

export const appNavigation: AppNavigationItem[] = [
  { to: "/", label: "Dashboard", description: "Governance overview" },
  { to: "/decks", label: "Org Map", description: "Teams, owners, and coverage" },
  { to: "/notes/view", label: "AI Systems", description: "Registry and access surfaces" },
  { to: "/stats", label: "Usage", description: "Spend, activity, and budgets" },
  { to: "/team", label: "Findings", description: "Remediation and review queue" },
];

export function pageTitleForPath(pathname: string): string {
  if (pathname === "/") return "Dashboard";
  if (pathname.startsWith("/onboarding/plan")) return "Choose Plan";
  if (pathname.startsWith("/stats")) return "Usage";
  if (pathname.startsWith("/settings")) return "User Settings";
  if (pathname.startsWith("/team")) return "Findings";
  if (pathname.startsWith("/notes/view")) return "AI Systems";
  if (pathname.startsWith("/notes/add")) return "Register System";
  if (pathname.startsWith("/marketplace")) return "Marketplace";
  if (pathname.startsWith("/templates")) return "Templates";
  if (pathname.startsWith("/decks")) return "Org Map";
  if (pathname.startsWith("/study-groups")) return "Study Groups";
  if (pathname.startsWith("/study/")) return "Study";
  if (pathname.startsWith("/tools/empty-cards")) return "Empty Cards";
  return "Cleartify";
}
