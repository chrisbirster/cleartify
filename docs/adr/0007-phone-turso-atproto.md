# ADR-0007: Phone-first Clear, Turso private state, optional ATProto social identity

- **Status:** Accepted
- **Date:** 2026-10-08

## Context

Most Cleartify users want one thing: save Clear as a contact and text it. Requiring a Bluesky/ATProto account during normal onboarding would expose infrastructure concepts that do not help that job.

At the same time, Cleartify's social/protocol marketplace benefits from portable public identity and user-owned public records. ATProto is a good fit for that public layer, while private assistant state must not be written to a public PDS.

Cleartify also needs a simple private database that works locally, in Docker, and with managed cloud durability.

## Decision

Cleartify separates the product into two data/identity planes.

### Clear: private plane

Regular users activate with a verified mobile number and an internal Cleartify user identifier.

Private and operational data lives in SQLite-compatible storage using Turso:

- phone/email identity mappings;
- conversations and messages;
- memories;
- reminders/responsibilities;
- private protocol installations and run state;
- approvals;
- tool credentials/grants;
- billing and entitlements;
- waitlist/invitation state.

The initial production topology uses the official Turso Go driver with a local database file and optional Turso Cloud sync. M1 runs a single writer process. We may later move to Turso's serverless driver or database-per-user isolation when scale or privacy requirements justify it.

### Cleartify Social: public plane

ATProto is optional and introduced only when a user wants public/social capabilities such as:

- a public profile;
- following creators;
- publishing protocols;
- creator verification/claims;
- public protocol collections;
- selling creator protocols.

Existing ATProto users bring their existing DID/PDS. New social users may use Bluesky-hosted ATProto account/PDS onboarding so Cleartify does not operate PDS infrastructure.

Public portable Cleartify records may live in the user's PDS. Cleartify can maintain a disposable AppView/search index in Turso for performance.

### Data rule

> If it is private, secret, transactional, or runtime state, it belongs in Cleartify's private Turso plane.

> If it should be public, portable, and owned by the creator/user, it may belong in ATProto.

ATProto/PDS records are never treated as a private memory store.

## Consequences

### Positive

- normal users never need to understand ATProto;
- the core onboarding remains phone-first;
- private assistant state stays private;
- creator/public data can be portable;
- Cleartify does not need to run PDS infrastructure;
- the Go backend can run locally against SQLite-compatible storage and sync to Turso Cloud;
- the social index can be rebuilt from canonical public records.

### Negative

- one person may have both an internal Cleartify ID and an optional DID;
- identity linking must be explicit and reversible;
- M1's sync topology intentionally starts with one writer process;
- premium protocol payloads cannot simply be published in a public PDS;
- AppView/indexing infrastructure is still required for fast social discovery.

## Product shorthand

**Clear is the private layer. Cleartify is the social layer.**
