# System Architecture

## Architectural stance

Start as a **modular monolith plus background workers**, not microservices. The product is early; the hard problems are product behavior, trust, integrations, and durable execution — not service topology.

The web SPA and messaging runtime share one API/backend boundary, but texting Clear does **not** depend on the SPA being open.

### Initial stack

- **Web:** SolidJS 2, TypeScript, Solid Router, StyleX, Vite.
- **API/runtime:** Go service with clear internal modules.
- **Primary database:** PostgreSQL.
- **Background work:** durable PostgreSQL-backed job queue/worker abstraction initially.
- **Object storage:** optional for large artifacts/attachments.
- **Email provider:** adapter interface; use for waitlist, newsletters, transactional mail, and beta invitations.
- **Messaging provider:** adapter interface for SMS/RCS/other supported text channels.
- **Model gateway:** provider-neutral model interface; route by task/cost/quality.
- **Payments:** add behind an interface when paid protocols enter beta.

## Top-level view

```mermaid
flowchart LR
    U[User] -->|texts| MP[Messaging provider]
    MP -->|signed webhook| API[Cleartify API]
    WEB[Solid SPA] -->|HTTPS| API
    EMAIL[Email provider] <--> API

    API --> DB[(PostgreSQL)]
    API --> Q[Durable jobs]
    Q --> CLR[Clear runtime]

    CLR --> CTX[Context + memory]
    CLR --> PE[Protocol engine]
    CLR --> MG[Model gateway]
    CLR --> AG[Approval gate]
    CLR --> TOOLS[Tool adapters]

    TOOLS --> EXT[Calendar / Email / Web / Other services]
    CLR -->|outbound reply| MP
```

## Backend modules

Keep these as modules/packages in one Go codebase until scaling or reliability requirements justify separation.

```text
internal/
  accounts/          users, identities, sessions
  waitlist/          signup + consent + referrals
  invites/           beta invitation + 5-digit verification
  messaging/         inbound/outbound provider adapters
  conversations/     threads + message events
  clear/             assistant orchestration
  memory/            long-lived personal context
  protocols/         definitions, versions, installs, runs
  approvals/         consequential-action policy and grants
  tools/             email/calendar/browser/etc adapters
  profiles/          public/creator/unclaimed profiles
  marketplace/       discovery, entitlements, offers
  notifications/     transactional + newsletter sends
  billing/           future paid protocol/subscription support
  jobs/              durable asynchronous execution
```

## Text message lifecycle

```mermaid
sequenceDiagram
    participant U as User
    participant M as Messaging Provider
    participant W as Webhook/API
    participant J as Durable Job Queue
    participant C as Clear Runtime
    participant P as Protocol Engine
    participant T as Tools

    U->>M: "Can you plan my trip next month?"
    M->>W: signed inbound webhook
    W->>W: verify signature + normalize + dedupe
    W->>J: enqueue inbound-message job
    W-->>M: 2xx quickly

    J->>C: process message
    C->>C: load identity, context, permissions, active protocols
    C->>P: resolve matching protocol / next action
    P->>T: research / calendar / other tool calls
    T-->>P: results
    P-->>C: outcome + next state
    C->>M: send natural-language reply
    M-->>U: Clear responds
```

### Important property

Inbound webhooks acknowledge quickly. Long model/tool work happens asynchronously. Every externally visible message/action gets an idempotency key so retries do not duplicate sends or purchases.

## Clear runtime

Clear is an orchestrator, not one giant prompt.

```mermaid
flowchart TD
    MSG[Inbound message] --> ID[Resolve user + conversation]
    ID --> CONTEXT[Assemble relevant context]
    CONTEXT --> INTENT[Understand intent]
    INTENT --> PROTO{Protocol match?}
    PROTO -->|yes| RUN[Continue/start protocol run]
    PROTO -->|no| PLAN[Plan one-off response/action]
    RUN --> POLICY[Permission + approval policy]
    PLAN --> POLICY
    POLICY -->|allowed| EXEC[Execute tools / model work]
    POLICY -->|needs approval| ASK[Ask user]
    EXEC --> MEMORY[Write durable outcome/memory]
    MEMORY --> RESP[Compose concise reply]
    ASK --> RESP
```

### Runtime concepts

- **Conversation:** the natural text thread with Clear.
- **Memory:** durable facts/preferences/relationships learned with provenance and user controls.
- **Responsibility:** an ongoing outcome Clear owns, such as “keep chasing support until resolved.”
- **Protocol:** reusable executable playbook.
- **Protocol run:** one instance of a protocol for one user.
- **Approval:** explicit user authority for a consequential step.
- **Tool grant:** scoped permission to access an external service.

## Data model overview

```mermaid
erDiagram
    USER ||--o| PROFILE : may_own
    USER ||--o{ IDENTITY : has
    USER ||--o{ CONSENT : grants
    USER ||--o{ CONVERSATION : has
    USER ||--o{ PROTOCOL_INSTALL : installs
    USER ||--o{ FOLLOW : follows

    PROFILE ||--o{ PROTOCOL : publishes
    PROFILE ||--o{ PROFILE_CLAIM : claim_attempts
    PROFILE ||--o{ FOLLOW : followed

    PROTOCOL ||--o{ PROTOCOL_VERSION : versions
    PROTOCOL_VERSION ||--o{ PROTOCOL_INSTALL : installed_version
    PROTOCOL_INSTALL ||--o{ PROTOCOL_RUN : runs

    CONVERSATION ||--o{ MESSAGE : contains
    PROTOCOL_RUN ||--o{ RUN_EVENT : emits
    PROTOCOL_RUN ||--o{ APPROVAL : requires

    WAITLIST_ENTRY ||--o{ BETA_INVITE : receives
    USER ||--o{ TOOL_GRANT : authorizes
```

## Suggested core tables

### Identity and access

- `users`
- `identities` — phone/email/provider identity mappings
- `sessions`
- `tool_grants`
- `consents`

### Launch funnel

- `waitlist_entries`
- `beta_invites`
- `beta_code_attempts`
- `notification_deliveries`

### Messaging / assistant

- `conversations`
- `messages`
- `memories`
- `responsibilities`
- `approvals`
- `jobs`

### Protocols

- `protocols`
- `protocol_versions`
- `protocol_installs`
- `protocol_runs`
- `protocol_run_events`
- `protocol_offers`
- `entitlements`

### Social / creator layer

- `profiles`
- `profile_sources`
- `profile_claims`
- `follows`

## Deployment shape for MVP

```mermaid
flowchart TB
    DNS[cleartify.com] --> EDGE[HTTPS / reverse proxy]
    EDGE --> APP[Go API + embedded or separately served SPA]
    EDGE --> WEBHOOK[/webhooks/messaging]
    WEBHOOK --> APP
    APP --> PG[(PostgreSQL)]
    WORKER[Go worker process] --> PG
    APP --> PROVIDERS[Email + Messaging + Model + Tool providers]
    WORKER --> PROVIDERS
```

One repository can build two process modes from the same Go codebase: `api` and `worker`. The SPA can later be embedded into the Go binary for a single deployable artifact if desired.

## Scale triggers

Do not split services because something feels “enterprise.” Split only when a real boundary appears, e.g.:

- messaging webhook throughput needs independent scaling;
- browser/tool execution requires isolated sandbox infrastructure;
- marketplace payments need stronger operational isolation;
- protocol workers have materially different compute/runtime needs;
- compliance requires a separate trust boundary.
