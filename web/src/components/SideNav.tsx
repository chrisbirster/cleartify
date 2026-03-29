import { NavLink } from "react-router-dom";

type NavItem = {
  label: string;
  to: string;
};

type SideNavProps = {
  items: NavItem[];
};

export function SideNav({ items }: SideNavProps) {
  return (
    <aside className="border-r border-border bg-card/80 px-4 py-6">
      <div className="mb-8 px-3">
        <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Workspace</p>
        <h1 className="mt-2 text-xl font-semibold text-foreground">Cleartify</h1>
      </div>
      <nav className="space-y-1">
        {items.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
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
