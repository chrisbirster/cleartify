# Round 6: text-first follow-ups (draft, flag-gated)

This is an engineering preview, NOT a deployed or production-ready reminder product.

## Opt-in flags

- `CLEAR_SESSIONS_ENABLED=true` and `CLEAR_CONVERSATION_ENABLED=true` are both required. Both default OFF.
- In that configuration the app initializes persistent follow-up tables and starts a worker that ticks every 30 seconds; SIGINT/SIGTERM cancels the worker.
- `FOLLOWUP_DELIVERY_CHANNEL` defaults to `in_app`. In-app delivery is durable database storage, **not a push notification or SMS**.
- `FOLLOWUP_DELIVERY_CHANNEL=debug_sms` is allowed only with `NOTIFICATION_MODE=debug`, which is loopback-only. This uses the existing `notifications.Sender` Debug implementation and is never a production provider.
- There is no configured real SMS provider. If the sender fails, the job retries with exponential backoff, capped at five attempts. The fifth failure persists a terminal `dead` delivery state (dead-letter); it is not automatically sent again. Provider errors are stored as the generic `delivery_failed` marker, not logged verbatim.
- Enabling the preview does NOT enable other protocols, checkout or external execution.

## Text API (same-origin + private Clear session)

`POST /api/me/clear/messages`, JSON body `{"text":"remind me to call Alice at 2026-11-01T18:00:00Z"}` or `{"text":"remind me to call Alice","dueAt":"2026-11-01T18:00:00Z"}`. Both require RFC3339 with timezone. A missing due time defaults to 24 hours after request, explicitly reported as `dueTimeDefaulted:true`. Completion: `{"text":"done <followup ID>"}` (command casing is ignored).

Creation replies include a follow-up record, `deliveryChannel`, and `deliveryStatus:"pending"`. They **do not say the nudge has already been delivered**.

`GET /api/me/clear/followups` lists only the current user's follow-ups; `GET /api/me/clear/nudges` returns only that user's in-app nudge queue, including actual persisted delivery state. Marking an item done removes it from future eligible nudges and from the open inbox.

`POST /api/me/self-serve-guide-installs` can opt in with `createFollowups:true` and optionally `followupDueAt` in RFC3339; an individual guide step may provide `dueAt` instead (which takes precedence). Absent a due time, each follow-up defaults to 24 hours. Installation without this opt-in creates no follow-ups. Only permissionless, public, non-taken-down guides are eligible.

## Security and delivery contract

- New behavior disabled by default, no bypass or client-supplied owner ID, no executable guide operations.
- Creation limit: 20 follow-ups per user per rolling 24 hours; nudge admission cap: 30 per user per rolling 24 hours (the 31st is deferred; tested at the exact 30/31 boundary).
- An atomic unique followup/due key prevents duplicate in-app nudge creation after retries and restarts.
- The worker uses a database lease and backoff for failed dispatch. The in-app channel can mark delivery durably. External SMS cannot guarantee exactly-once delivery across a crash after send but before database commit without downstream provider idempotency; **do not represent the Debug SMS path as exactly-once**.
- In-app visibility is not proof that a user read the message. No push transport, read receipts, recurrence, rich natural-language time interpretation, SMS provider, or deployed scheduler is present.
- #100 separately addresses a uniqueness constraint/audit for the already merged self-service install code and safe failed-login delivery. #100 is not merged and must be reviewed before exposing install paths to concurrency.
