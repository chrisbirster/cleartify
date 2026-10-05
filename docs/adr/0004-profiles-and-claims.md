# ADR-0004: People-first profiles, seeded profiles and claims

- **Status:** Accepted
- **Date:** 2026-10-04

## Context

A cold protocol marketplace filled with anonymous workflows is difficult to understand and distribute. People already care about how trusted creators, experts and friends operate. Cleartify can make those operating systems installable.

The product also needs a way to have useful public discovery before every notable creator registers.

## Decision

Cleartify uses a **people-first social graph**:

- profiles own/publish protocols;
- users follow profiles;
- profiles are a primary protocol-discovery surface;
- Chris is the default first follow/friend for new accounts;
- ordinary users are private by default;
- creators explicitly choose what protocol/activity information is public.

A profile may exist with no `owner_user_id`. Such seeded profiles are allowed only when based on sourced public information and must be visibly labeled **Unclaimed**.

Unclaimed profiles:

- must not imply the person joined or endorsed Cleartify;
- keep provenance for sourced claims/protocol descriptions;
- cannot expose private/current activity claims such as “running now”;
- cannot monetize as that person;
- support correction/removal/claim requests.

Claiming binds the verified person/account to the existing profile rather than creating a duplicate.

## Consequences

### Positive

- marketplace discovery maps to human trust and existing audiences;
- profile links become organic distribution;
- creators have a clear home for free and paid protocols;
- seeded, sourced profiles improve cold-start usefulness;
- users can understand “what this person uses” without understanding workflow technology.

### Negative

- identity verification and provenance become important product infrastructure;
- incorrect attribution can damage trust and must be easy to correct;
- public-profile seeding requires conservative policies and operational review;
- we must prevent unclaimed pages from looking like endorsements.

## Monetization rule

Only claimed/verified creator profiles may sell protocols or receive creator revenue.
