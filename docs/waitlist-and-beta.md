# Waitlist and Beta Access

## Goals

The launch funnel has three distinct jobs:

1. collect an interested person's contact details;
2. obtain clear marketing consent for newsletters/product updates;
3. later convert selected people into verified beta users.

Do not conflate marketing consent with permission for operational/assistant messages.

## Waitlist form

Required fields:

- `name`
- `email`
- `mobile_phone` normalized to E.164

Optional/explicit fields:

- `email_marketing_opt_in`
- referral/source metadata

### Suggested records

```text
waitlist_entries
  id
  name
  email_normalized
  phone_e164
  status                # waiting | invited | activated | declined
  source
  referrer_profile_id
  created_at
  updated_at

consents
  id
  subject_type
  subject_id
  channel               # email_marketing | sms_marketing | transactional
  granted_at
  revoked_at
  policy_version
  evidence_json
```

Newsletter unsubscribes update consent; they do not delete the beta application or operational account identity.

## Beta invitation flow

A five-digit code alone is too small to function as the primary bearer secret. The invitation therefore combines a high-entropy one-time email link with the requested five-digit code.

The cleanest flow uses both contact channels collected at waitlist time:

```mermaid
sequenceDiagram
    participant A as Admin
    participant C as Cleartify
    participant E as Email
    participant U as User
    participant S as SMS

    A->>C: Invite waitlist entry
    C->>C: Create high-entropy one-time token
    C->>E: Send beta invitation link
    E-->>U: "You're in — activate beta"
    U->>C: Open signed one-time link
    C->>C: Validate link, create 5-digit challenge
    C->>S: Send 5-digit code to stored mobile
    S-->>U: "Your Cleartify code is 48217"
    U->>C: Enter 48217
    C->>C: Rate-limit + verify + consume invite
    C-->>U: Account activated; save Clear contact
```

This gives the 5-digit code a real job: it confirms possession of the phone number that will later identify the user when they text Clear.

## Invite record

```text
beta_invites
  id
  waitlist_entry_id
  email_token_hash
  email_token_expires_at
  email_token_used_at
  challenge_code_hash
  challenge_expires_at
  challenge_attempts
  max_challenge_attempts
  activated_user_id
  created_at
  activated_at
```

Rules:

- never store raw invitation tokens or raw codes;
- invitation link is single-use or bound to an activation session;
- 5-digit challenge expires quickly (e.g. ~10 minutes);
- strict attempt limit and resend cooldown;
- log security events without logging secrets;
- successful activation consumes both challenge and invite;
- reissuing an invite invalidates previous active secrets.

## Newsletter and updates

Separate delivery categories:

- **marketing:** newsletter, launch notes, creator/protocol highlights;
- **transactional:** waitlist confirmation, security messages, beta invitation;
- **assistant/service:** messages created by the user's use of Clear.

A user who unsubscribes from marketing should still receive transactional security messages required to use the beta.

## Activation UX

The activation web page should be intentionally tiny:

1. “You're invited.”
2. confirm partially masked email/phone;
3. “We texted your 5-digit beta code.”
4. enter code;
5. accept current terms/privacy;
6. choose display name/handle if desired;
7. “Add Clear to your contacts” / send the first text.

The user should reach their messaging app as quickly as possible.

## After activation

Default setup:

- starter protocols installed;
- personal profile private;
- Chris followed as the first friend;
- no public running-protocol activity enabled;
- initial conversation contains a short welcome from Clear;
- Clear explains permissions only when relevant rather than forcing a giant setup wizard.
