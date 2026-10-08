# Cleartify

**Text Clear. It'll figure it out.**

Cleartify is a text-first personal assistant. The product should feel less like software you manage and more like a capable friend already in your contacts.

> “Can you handle this?”  
> **Clearly.**

## Product thesis

- **Text-first:** Clear is a contact, not another app destination.
- **Relationship:** persistent context, memory, follow-through and useful check-ins.
- **Execution:** Clear keeps moving an outcome forward instead of stopping at an answer.
- **Protocols:** reusable, installable ways of getting things done.
- **Profiles:** people share the protocols they choose to make public; creators can publish free or premium protocols.
- **Hidden agents:** models, tools and specialist agents stay behind the scenes; the user talks to one assistant.
- **Trust:** private by default, scoped permissions, explicit approval before consequential actions.

## Web

The new SPA uses:

- SolidJS 2 RC
- TypeScript
- Solid Router
- StyleX
- Vite

```bash
npm install
npm run dev
```

## Go API

The M1 backend is a Go modular monolith using SQLite-compatible Turso storage.

```bash
cp .env.example .env
# export variables from .env in your shell, then:
go run ./cmd/cleartify
```

Endpoints:

- `GET /healthz`
- `POST /api/waitlist`

Local development uses `DATABASE_PATH`. Set `TURSO_REMOTE_URL` and `TURSO_AUTH_TOKEN` to sync the database to Turso Cloud. M1 intentionally runs one backend writer while using sync mode.

Current routes:

- `/` — meme-friendly consumer landing page + waitlist UI
- `/@chris` — first-friend / creator protocol profile preview

The landing waitlist now calls the real `POST /api/waitlist` endpoint. Set `VITE_API_BASE_URL` when the API is on a different origin.

## Docs

Start with [`docs/README.md`](./docs/README.md).

Key documents:

- [`docs/mvp.md`](./docs/mvp.md) — phased product plan
- [`docs/architecture.md`](./docs/architecture.md) — messaging/runtime/backend architecture + Mermaid diagrams
- [`docs/protocols.md`](./docs/protocols.md) — protocol engine and creator model
- [`docs/profiles-and-marketplace.md`](./docs/profiles-and-marketplace.md) — social profiles, claims and premium protocol marketplace
- [`docs/waitlist-and-beta.md`](./docs/waitlist-and-beta.md) — signup/newsletter/invite/5-digit activation flow
- [`docs/brand.md`](./docs/brand.md) — Clear voice and the **Clearly.** device
- [`docs/deployment.md`](./docs/deployment.md) — Cloudflare + Docker/Fly deployment shape
- [`docs/adr/`](./docs/adr/) — accepted architecture/product decisions

## Branching

Use the project workflow:

```text
feature/* → dev → main → release tag
```

`main` is releasable. `dev` is the integration branch.

This repository was intentionally reset. The previous implementation remains recoverable through Git history.
