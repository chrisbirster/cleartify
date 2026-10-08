# System Architecture

## Architectural stance

Start as a **modular Go monolith plus durable workers**, not microservices.

Cleartify has two product planes:

- **Clear** is the private assistant a normal user texts.
- **Cleartify Social** is the optional public layer for profiles, creators, protocols, discovery, and the marketplace.

The web SPA is useful for onboarding, permissions, discovery, creator tools, and settings. It is not required to keep a text conversation with Clear alive.

## Initial stack

- **Web:** SolidJS 2, TypeScript, Solid Router, StyleX, Vite.
- **Public web hosting:** Cloudflare Workers Static Assets.
- **API/runtime:** Go.
- **Private database:** SQLite-compatible Turso.
- **M1 database topology:** local Turso database file with optional Turso Cloud sync; one writer process.
- **Backend deployment:** Docker, initially suitable for Fly.io; no Fly-specific application code.
- **Public/social data:** optional ATProto records in a user's PDS.
- **Social search/AppView:** disposable Turso index derived from public ATProto records.
- **Messaging:** provider adapter for SMS/RCS/other supported text channels.
- **Models:** provider-neutral model gateway.
- **Tool interoperability:** MCP where available.
- **Agent interoperability:** A2A where available, with DEEZ/NUTS delegation semantics when Clear acts for a user.

## Product/data split

```mermaid
flowchart LR
    USER[Regular user] -->|texts| CLEAR[Clear]
    CLEAR --> PRIVATE[(Turso private state)]

    CREATOR[Social user / creator] --> WEB[Cleartify web]
    WEB --> SOCIAL[Cleartify Social]
    SOCIAL --> PDS[User's ATProto PDS]
    PDS --> INDEX[(Cleartify AppView index)]

    SOCIAL --> CLEAR
    INDEX --> WEB
```

### Clear: private plane

A regular user needs only a verified mobile identity plus an internal Cleartify user ID.

Private data includes:

- waitlist/invite state;
- phone and email mappings;
- conversations;
- memories;
- reminders and responsibilities;
- private protocol installations;
- protocol run state;
- approvals;
- credentials/tool grants;
- billing and entitlements.

This data stays in Cleartify's private Turso plane.

### Cleartify Social: optional public plane

ATProto is introduced only when a user wants public/social capabilities such as:

- a public profile;
- publishing protocols;
- following creators publicly;
- claiming a seeded profile;
- creator verification;
- public protocol collections.

Existing ATProto users bring their current DID/PDS. New social users should be able to use Bluesky-hosted ATProto account/PDS infrastructure so Cleartify does not operate PDS servers.

ATProto is not used as private assistant memory storage.

## M1 request path

```mermaid
sequenceDiagram
    participant U as Visitor
    participant W as Solid SPA / Cloudflare
    participant A as Go API
    participant D as Turso
    participant C as Turso Cloud

    U->>W: submit name/email/mobile
    W->>A: POST /api/waitlist
    A->>A: normalize + validate
    A->>D: upsert waitlist entry
    A->>D: append consent record
    D-->>A: committed
    A->>C: push when cloud sync is configured
    A-->>W: 201 joined
    W-->>U: You're on the list. Clearly.
```

The landing page must never fake a successful signup if the backend failed.

## Text message lifecycle

```mermaid
sequenceDiagram
    participant U as User
    participant M as Messaging provider
    participant A as Go API
    participant J as Durable jobs
    participant C as Clear runtime
    participant P as Protocol engine
    participant T as Tools / agents

    U->>M: text Clear
    M->>A: signed inbound webhook
    A->>A: verify, normalize, dedupe
    A->>J: persist/enqueue message work
    A-->>M: acknowledge quickly

    J->>C: process durable job
    C->>C: load memory, permissions, responsibilities
    C->>P: match/continue protocol if relevant
    P->>T: MCP / API / A2A work
    T-->>P: result/status
    P-->>C: outcome + next state
    C->>M: outbound reply
    M-->>U: Clear responds
```

Long-running work is represented durably in Turso. Goroutines may execute work, but an in-memory goroutine is never the source of truth for an ongoing responsibility.

## Go module boundaries

Keep these as packages in one codebase until a real scaling or security boundary appears.

```text
cmd/cleartify/

internal/
  api/
  config/
  database/
  waitlist/
  accounts/
  invites/
  messaging/
  conversations/
  clear/
  memory/
  protocols/
  approvals/
  tools/
  jobs/
  social/
  atproto/
  marketplace/
  billing/
  notifications/
```

## Private data model

The first schema starts with launch data:

- `waitlist_entries`
- `consents`

Later private tables include:

- `users`
- `phone_identities`
- `beta_invites`
- `sessions`
- `conversations`
- `messages`
- `memories`
- `responsibilities`
- `jobs`
- `approvals`
- `tool_grants`
- `protocol_installs`
- `protocol_runs`
- `entitlements`

Public creator/protocol records should not automatically be duplicated as canonical private records. Turso may hold a rebuildable index for search and ranking.

## Deployment shape

```mermaid
flowchart TB
    DNS[cleartify.com] --> CF[Cloudflare]
    CF --> SPA[Solid SPA static assets]
    SPA --> API[api.cleartify.com]

    SMS[Messaging provider] -->|webhook| API
    API --> GO[Go modular monolith]
    GO --> LOCAL[(Local Turso file)]
    LOCAL <-->|sync| TC[(Turso Cloud)]
    GO --> PROVIDERS[Models / email / tools / MCP / A2A]

    ATP[ATProto network] --> IDX[Social indexer]
    IDX --> INDEX[(Disposable AppView index)]
    INDEX --> API
```

For M1, run **one backend writer**. This keeps Turso local-first sync semantics simple. When horizontal API scaling becomes necessary, move the private store to a direct remote/serverless Turso topology or another supported multi-writer arrangement without changing the domain interfaces.

## Docker portability

The backend must remain runnable as:

```bash
docker run cleartify
```

Application code must not depend on Fly.io APIs. Fly can be the first container host, while Cloudflare handles DNS, TLS, CDN, and the SPA. Cloudflare Containers or another host can be evaluated later without rewriting Clear.

## Scale triggers

Split components only when a measured boundary appears, for example:

- messaging webhook volume needs independent scaling;
- browser execution needs an isolated sandbox;
- protocol jobs have materially different resource requirements;
- creator billing needs a stronger trust boundary;
- ATProto indexing becomes operationally independent;
- private database topology needs multi-region/multi-writer behavior.
