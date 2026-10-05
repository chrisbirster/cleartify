# ADR-0001: Clear is the single user-facing assistant

- **Status:** Accepted
- **Date:** 2026-10-04

## Context

Cleartify may use multiple models, tools, specialist agents, protocol steps, queues and integrations. Exposing all of that to normal users creates configuration burden and makes the product feel like software that must itself be managed.

The intended experience is closer to texting a capable friend/assistant.

## Decision

The normal user interacts with one identity: **Clear**.

Primary interaction is text messaging. The web app is secondary and exists for onboarding, permissions, profile/marketplace discovery, creator tooling, billing and history.

Specialist agents/models/tools are implementation details behind Clear and are not represented as coworkers the user must configure.

## Consequences

### Positive

- extremely low interaction friction;
- stable relationship/personality even when model providers change;
- Clear can route work internally without teaching users agent concepts;
- messaging channel becomes a durable habit rather than another app destination.

### Negative

- orchestration quality becomes our responsibility;
- debugging must expose sufficient internal traces to operators without leaking complexity into the consumer UX;
- power users may eventually want advanced controls, which should live behind an explicit expert surface rather than the default experience.

## Brand implication

The assistant is **Clear**. The recurring punchline is **“Clearly.”**
