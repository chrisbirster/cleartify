import { NavLink } from 'react-router'

function navItemBase(isActive: boolean) {
  return [
    'flex min-h-11 flex-1 flex-col items-center justify-center rounded-2xl px-2 py-2 text-[11px] font-medium transition',
    isActive
      ? 'bg-[var(--app-accent)] text-[var(--app-accent-ink)]'
      : 'text-[var(--app-text-soft)] hover:bg-[var(--app-muted-surface)] hover:text-[var(--app-text)]',
  ].join(' ')
}

function NavGlyph({ path }: { path: string }) {
  return (
    <svg className="mb-1 h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d={path} />
    </svg>
  )
}

export function MobileBottomNav({ onOpenMore }: { onOpenMore: () => void }) {
  return (
    <nav
      className="fixed inset-x-0 bottom-0 z-40 border-t border-[var(--app-line)] bg-[color:var(--app-header)] px-3 py-3 backdrop-blur md:hidden"
      style={{ paddingBottom: 'calc(0.75rem + env(safe-area-inset-bottom))' }}
    >
      <div className="flex items-center gap-2">
        <NavLink to="/" end className={({ isActive }) => navItemBase(isActive)}>
          <NavGlyph path="M3 10.5 12 3l9 7.5V20a1 1 0 0 1-1 1h-5.5v-6h-5V21H4a1 1 0 0 1-1-1z" />
          Dashboard
        </NavLink>
        <NavLink to="/decks" className={({ isActive }) => navItemBase(isActive)}>
          <NavGlyph path="M20 10c0 5-5.54 10.2-7.4 11.8a1 1 0 0 1-1.2 0C9.54 20.2 4 15 4 10a8 8 0 0 1 16 0Z M12 10h.01" />
          Org Map
        </NavLink>
        <NavLink to="/notes/view" className={({ isActive }) => navItemBase(isActive)}>
          <NavGlyph path="M12 2 2 7l10 5 10-5-10-5Z M2 12l10 5 10-5 M2 17l10 5 10-5" />
          Systems
        </NavLink>
        <NavLink to="/stats" className={({ isActive }) => navItemBase(isActive)}>
          <NavGlyph path="M22 12h-2.5a2 2 0 0 0-1.9 1.45l-2.4 8.1a.25.25 0 0 1-.48 0L9.24 2.2a.25.25 0 0 0-.48 0l-2.35 8.35A2 2 0 0 1 4.5 12H2" />
          Usage
        </NavLink>
        <button
          type="button"
          onClick={onOpenMore}
          className={navItemBase(false)}
        >
          <NavGlyph path="M5 12h.01M12 12h.01M19 12h.01" />
          More
        </button>
      </div>
    </nav>
  )
}
