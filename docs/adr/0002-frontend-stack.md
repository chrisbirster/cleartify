# ADR-0002: SolidJS 2 SPA with TypeScript, Solid Router and StyleX

- **Status:** Accepted
- **Date:** 2026-10-04

## Context

Cleartify needs a lightweight public/consumer web surface for the landing page, waitlist, beta activation, profiles, protocol discovery, creator studio and account settings. The primary assistant experience remains messaging, so the web stack should stay fast and relatively simple.

## Decision

Use:

- SolidJS 2 release-candidate line until final 2.0 is available;
- TypeScript;
- Solid Router;
- StyleX;
- Vite;
- SPA architecture for M0/MVP.

The initial dependency versions are deliberately pinned/ranged in `package.json`. Upgrade from Solid 2 RC to final 2.0 is an explicit maintenance task rather than a reason to fall back to Solid 1.x.

## Consequences

### Positive

- small, reactive client;
- consistent TypeScript codebase;
- route-level code splitting available as surfaces grow;
- StyleX gives build-time atomic CSS and shared design primitives;
- Vite provides a minimal development/build pipeline.

### Negative

- Solid 2 is pre-final at the time of this ADR and may require migration changes;
- StyleX + Solid is less conventional than StyleX + React, so integration should stay covered by build tests;
- SPA deployments require history fallback to `index.html` for client routes.

## Follow-up

Add CI that performs clean install, typecheck and production build before M0 merges into `dev`.
