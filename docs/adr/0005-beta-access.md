# ADR-0005: Waitlist and two-channel beta activation

- **Status:** Accepted
- **Date:** 2026-10-04

## Context

The public landing page needs to collect name, email and mobile number, send optional newsletter/product updates, and later invite selected people into a limited beta.

A five-digit code is desirable as a memorable beta-gate interaction, but a five-digit secret by itself has insufficient entropy to be the sole invitation credential.

## Decision

Waitlist signup collects:

- name;
- email;
- mobile phone;
- explicit email-marketing consent.

Beta activation uses two channels:

1. Cleartify sends a transactional invitation email containing a high-entropy one-time activation link.
2. After the user opens the valid link, Cleartify sends a five-digit challenge to the waitlisted mobile number.
3. The user enters the challenge on the activation page.
4. Successful verification consumes the invitation and creates/activates the account.

Raw invite tokens and challenge codes are not stored; store hashes, expiry and attempt state.

## Consequences

### Positive

- preserves the desired five-digit beta-code experience;
- validates the mobile identity that will later text Clear;
- a forwarded email alone does not activate another person's phone identity;
- marketing consent remains separate from required transactional/service messaging.

### Negative

- activation depends on both email and SMS deliverability;
- phone-number changes need a recovery/manual-review path;
- invite/code abuse requires rate limits and resend controls.

## Privacy/consent rule

Newsletter opt-out does not disable security/transactional messages needed for an active beta account. SMS marketing is not implied by providing the phone number for activation or assistant messaging.
