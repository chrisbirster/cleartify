# ADR-0003: Declarative protocols with hidden specialist agents

- **Status:** Accepted
- **Date:** 2026-10-04

## Context

Cleartify needs repeatable, shareable ways of getting things done. Users should be able to install a workflow from another person without granting arbitrary code execution or learning how to manage an agent graph.

## Decision

Protocols are versioned **declarative definitions**, not arbitrary uploaded programs.

A protocol declares:

- purpose;
- triggers;
- inputs;
- constrained steps;
- capability/permission manifest;
- approval behavior;
- version/changelog.

Clear executes the protocol using internal specialist agents/models/tools. Those internals are hidden from normal users.

Published protocol versions are immutable. New permissions require renewed approval.

## Consequences

### Positive

- protocol installation has an inspectable security boundary;
- marketplace content can be reviewed/scanned;
- model/provider implementation can change without changing the protocol identity;
- creators publish behavior, not infrastructure configuration;
- the same protocol can personalize itself using each user's context and grants.

### Negative

- the schema/engine must grow deliberately as creators request expressiveness;
- arbitrary programming-language flexibility is intentionally sacrificed;
- migration/version compatibility becomes a product responsibility.

## Non-goal

The consumer product will not expose a default “create five agents and wire them together” UI.
