# Live Workspace V1 Plan

## Purpose

This plan turns Cleartify from a polished demo workspace into a believable live product for real teams.

The current app already proves the concept:
- Context Inbox
- Work Objects
- Prototype-to-Plan
- Agent Workspace
- Review and Releases
- `/me`, `/team`, Cluster resolve, and CRUD surfaces for Signals, Decisions, and Artifacts

What still feels wrong is the product model:
- demo/seeded data is still too central
- there is no first-class Project object to organize work
- onboarding explains the app instead of configuring it
- integrations are not the source of truth yet
- Skills / playbooks are not first-class operational tools

Live Workspace V1 fixes that by introducing a new setup and operating model:

**Workspace → Projects → Work Objects → Plan → Agent Actions → Review → Release**

---

## Product North Star

Cleartify should feel like the product system where a real team:
1. creates a workspace
2. creates or imports a project
3. connects source systems and AI providers
4. routes product context into project-scoped work
5. turns work into plans and outputs
6. reviews and releases work with shared context intact

---

## Non-goals for Live Workspace V1

These are intentionally deferred:
- full enterprise RBAC
- tenant subdomains
- billing
- invite approval workflows
- deep two-way sync for every integration
- internal docs/wiki replacement
- agent orchestration engine
- background job platform
- skills marketplace
- multi-project portfolio analytics

---

## Recommended Epic Order

1. **Epic 1 — Demo Mode vs Live Workspace Mode**
2. **Epic 2 — Projects as a first-class entity**
3. **Epic 3 — Real setup onboarding wizard**
4. **Epic 4 — Connectors and provider setup**
5. **Epic 5 — Skills / Playbooks**

This order matters. Do not start with Skills or integrations before the product has a live workspace model and a project model.

---

# Epic 1 — Demo Mode vs Live Workspace Mode

## Why this epic exists

Right now the app still feels like a seeded demo even when a real user signs in. That hurts trust immediately. A new user should choose between a sample workspace and a real workspace instead of being dropped into demo-looking data by default.

## Goal

Split the product into two explicit modes:
- **Demo Workspace Mode** for evaluation, sales, and local review
- **Live Workspace Mode** for real users and their own connected data

## User experience

When a signed-in user enters the app for the first time, they should see a workspace mode chooser:
- Start with a sample workspace
- Set up a real workspace

If they choose:
- **Sample workspace**: they get the current polished hero flow with seeded data
- **Real workspace**: they start with a clean workspace and setup guidance

## In scope

- Add a workspace mode concept
- Add first-run mode selection for authenticated users
- Keep demo/sample data isolated to sample mode only
- Ensure the shell and routes behave correctly in both modes
- Add clear visual labeling when the user is in a sample workspace
- Ensure seeded data does not appear in live workspaces unless explicitly imported

## Out of scope

- project templates beyond one or two basic setup presets
- deep data import
- billing or plans
- advanced workspace settings

## Required product behavior

- A first-time signed-in user must choose sample vs live setup
- Sample mode should clearly say it is a sample workspace
- Live mode should start with either:
  - no projects yet, or
  - a setup wizard that immediately leads into project creation
- Existing hero flow remains available in sample mode for demos

## Data/model changes

Add minimal workspace-level state:
- `workspace_mode`: `sample | live`
- `sample_initialized_at` or equivalent if needed
- `live_initialized_at` or equivalent if needed

If current user/team/workspace records already exist, extend them rather than creating a parallel concept.

## Backend work

- Add workspace mode persistence
- Add endpoint(s) to:
  - get current workspace mode
  - set workspace mode on first-run
- Update seed/bootstrap behavior so sample data is loaded only for sample workspaces or demo mode
- Prevent live workspaces from implicitly booting into seeded hero data

## Frontend work

- Add a first-run workspace mode chooser screen
- Add clear UI labeling for sample workspace mode
- Route users correctly after choosing sample vs live
- Update dashboard / Today page to avoid showing sample content in live mode

## Validation and tests

### Backend
- workspace mode persists correctly
- sample initialization does not run for live workspaces
- live workspace starts empty or setup-ready

### Frontend
- first-run chooser renders correctly
- sample mode lands in sample hero flow
- live mode lands in setup flow
- sample mode is visibly labeled

### Manual verification
- sign in with a fresh user
- choose sample workspace and confirm hero path is present
- sign in with another fresh user, choose live workspace, confirm no seeded demo data is shown by default

## Definition of done

Epic 1 is done when:
- a new authenticated user explicitly chooses sample vs live workspace mode
- sample workspaces still support the hero demo flow
- live workspaces do not show seeded demo content by default
- the UI clearly indicates when the user is inside a sample workspace
- all routes still behave correctly in both modes
- all validation checks pass

## Codex prompt for Epic 1

```text
Read AGENTS.md, README.md, phase.md, and live-workspace-v1-plan.md first.

Implement Epic 1 only: Demo Mode vs Live Workspace Mode.

Goal:
Split Cleartify into explicit sample-workspace and live-workspace modes so real users are not dropped into seeded demo data by default.

Scope:
- add a persisted workspace mode concept
- add first-run mode selection for authenticated users
- keep sample hero flow available only in sample mode
- ensure live mode starts clean/setup-ready and does not show seeded demo data by default
- add visible UI labeling when a user is in sample mode

Requirements:
- keep changes small and clean
- preserve existing hero flow in sample mode
- do not start Projects, connectors, or Skills yet
- do not add billing or advanced settings

Return:
1. implementation summary
2. files changed
3. checks run
4. manual verification
5. whether Epic 1 is done
6. remaining items, if any
```

---

# Epic 2 — Projects as a First-Class Entity

## Why this epic exists

The current dashboard feels flat because everything is in one workspace bucket. Real teams need to organize work by project. This is the missing structural layer that makes the app feel like a product system instead of a seeded flow.

## Goal

Introduce **Project** as the main organizing unit for work.

## User experience

A user should be able to:
- create a project
- view a project dashboard
- filter inbox/work objects/releases by project
- understand which work belongs to which project
- switch between projects without losing context

## In scope

- add `Project` as a first-class entity
- create at least:
  - project list
n  - project creation flow
  - project detail/dashboard
- associate existing core objects with a project where appropriate:
  - Signals
  - Clusters
  - Work Objects
  - Plans
  - Agent outputs / Artifacts
  - Releases
- add project filtering/scoping to the main navigation flow

## Out of scope

- portfolio-level analytics
- complex roadmaps
- issue hierarchy like epic/story/subtask trees
- advanced permissions per project
- cross-project automation

## Data/model changes

Add a `projects` entity with a minimal shape:
- `id`
- `workspace_id`
- `name`
- `slug`
- `summary`
- `status`
- `created_at`
- `updated_at`

Add `project_id` to key entities where needed.

Start simple. If a Signal does not naturally belong to a project yet, allow it to be unassigned until triaged.

## Backend work

- create project persistence and routes
- add project lookup/list/create/update
- link Work Objects to projects
- update existing list/detail APIs to accept project scoping where needed
- ensure the existing hero flow can be mapped cleanly to one project

## Frontend work

- add Projects to the main navigation
- add project list page
- add create project flow
- add project dashboard page
- add project context in Work Objects, Inbox, Plans, Review, Releases
- add project switching/filtering in a minimal way

## Dashboard correction

Replace the vague current “Today” experience with something project-aware.
A user should be able to see:
- all projects
- or one selected project
- open cluster count
- active work objects
- release-ready work objects
- recent outputs

## Validation and tests

### Backend
- project CRUD works for MVP scope
- project associations persist correctly
- project filtering does not break existing hero flow

### Frontend
- project creation works
- project dashboard loads cleanly
- inbox/work/release views can be scoped to a project
- project navigation is obvious

### Manual verification
- create a new project
- assign or create work in that project
- confirm related pages reflect project context
- verify the sample hero flow works inside its own sample project

## Definition of done

Epic 2 is done when:
- Projects exist as a first-class entity
- a user can create and open a project
- Work Objects and the main flow are clearly scoped to a project
- dashboard and navigation feel project-aware
- the app no longer feels like one flat workspace bucket
- all validation checks pass

## Codex prompt for Epic 2

```text
Read AGENTS.md, README.md, phase.md, and live-workspace-v1-plan.md first.

Implement Epic 2 only: Projects as a first-class entity.

Goal:
Introduce Project as the main organizing unit so the workspace no longer feels flat.

Scope:
- add project persistence and minimal CRUD
- add project list/create/detail/dashboard surfaces
- associate the main product flow with project context
- make inbox/work/release views project-aware

Requirements:
- keep the model minimal and MVP-focused
- do not build portfolio analytics or advanced permissions
- preserve the hero flow by mapping it to a project
- keep changes small and clean

Return:
1. implementation summary
2. files changed
3. checks run
4. manual verification
5. whether Epic 2 is done
6. remaining items, if any
```

---

# Epic 3 — Real Setup Onboarding Wizard

## Why this epic exists

The current onboarding is informational, not operational. Real onboarding should configure the workspace instead of showing a brochure.

## Goal

Turn onboarding into a setup wizard that gets a new workspace ready for real use.

## User experience

For a live workspace, onboarding should guide the user through:
1. workspace/team name
2. avatar/logo (optional)
3. first project creation
4. template choice (blank or one/two simple templates)
5. optional teammate invite step
6. connector/provider setup handoff
7. finish into the actual project dashboard

## In scope

- replace the current passive walkthrough for live workspaces
- retain or adapt sample walkthrough separately for sample mode if useful
- persist onboarding progress/completion
- support a multi-step setup flow

## Out of scope

- advanced avatar/media pipeline
- complex invite approval flows
- multi-project bootstrapping
- deep template marketplace

## Setup wizard steps

### Step 1 — Name workspace
- workspace/team name
- optional short description

### Step 2 — Branding
- optional avatar/logo upload or generated placeholder

### Step 3 — Create first project
- project name
- short summary
- optional status

### Step 4 — Choose starting point
- blank project
- sample project structure
- one or two templates only

### Step 5 — Invite teammates (optional)
- optional emails
- or skip for now

### Step 6 — Connect tools / AI providers
- handoff to connectors screen or embedded mini-step

### Step 7 — Finish
- land in project dashboard

## Backend work

- persist onboarding progress and completion state
- persist workspace/profile setup fields
- persist first project creation through the onboarding flow
- optionally store deferred invite placeholders if included

## Frontend work

- build the multi-step wizard
- add progress indicator
- support skip/optional actions where appropriate
- handle validation cleanly
- land the user in a real destination at the end

## Relationship to sample mode

- sample workspace mode can keep a lighter educational walkthrough
- live workspace mode must use the setup wizard

## Validation and tests

### Backend
- onboarding state persists between steps
- first project creation via onboarding persists correctly
- completion state flips only when required steps are done

### Frontend
- wizard step progression works
- optional steps can be skipped safely
- completed onboarding lands the user in the right place

### Manual verification
- start with a fresh live workspace
- complete the wizard
- confirm workspace/profile/project state is created
- confirm user lands in a usable project view

## Definition of done

Epic 3 is done when:
- live workspace onboarding is a setup wizard, not a brochure
- it persists real workspace configuration
- it creates the first project
- it supports optional invite/branding steps without blocking progress
- completion lands the user in a project-ready experience
- all validation checks pass

## Codex prompt for Epic 3

```text
Read AGENTS.md, README.md, phase.md, and live-workspace-v1-plan.md first.

Implement Epic 3 only: Real setup onboarding wizard.

Goal:
Replace the current informational onboarding for live workspaces with a real setup wizard that configures the workspace and creates the first project.

Scope:
- workspace/team naming
- optional branding step
- first project creation
- simple starting-point/template selection
- optional teammate invite step
- finish into a real project dashboard

Requirements:
- keep the wizard minimal and practical
- preserve sample-mode walkthrough separately if needed
- do not build a large template system
- keep changes small and clean

Return:
1. implementation summary
2. files changed
3. checks run
4. manual verification
5. whether Epic 3 is done
6. remaining items, if any
```

---

# Epic 4 — Connectors and Provider Setup

## Why this epic exists

The app still feels fake if the primary context is seeded/demo content. The fastest way to make it feel real is to let the user connect the systems they already use and configure an AI provider.

## Goal

Create a simple connectors and provider setup experience that makes Cleartify operational.

## First connectors to support

Start narrow:
- **GitHub**
- **One docs source**: Notion *or* Confluence
- **One AI provider**: OpenAI *or* Anthropic

Slack can come later.
Linear/Jira can come later unless one is strategically required.

## In scope

- connectors/settings page
- ability to add provider/config credentials
- basic connector status display
- project/workspace-level configuration as appropriate
- validated external link/source typing improvements where needed

## Out of scope

- deep two-way sync
- background ingestion platform
- all provider options at once
- internal docs suite
- consumer-chat subscription integrations

## Product decision

Use:
- **OAuth** where natural for SaaS connectors
- **API keys** or admin-managed provider secrets for AI providers

Do not rely on personal consumer subscriptions.

## User experience

A user should be able to:
- open a Connectors / AI Providers settings area
- connect GitHub
- connect Notion or Confluence
- provide an AI provider key/config
- see connector status
- understand which project/workspace uses which integrations

## Backend work

- add connector/provider configuration models
- add secure storage or placeholder secure storage boundaries for secrets
- add validation and health-check behavior where practical
- expose configuration status to the frontend

## Frontend work

- add Connectors / Providers settings page
- show connected/disconnected/error states
- support configuration for the first provider set
- show connector setup as part of live onboarding handoff where helpful

## How this connects to agent actions

Agent actions should later use configured providers, not demo-only deterministic generators.
This epic does not need to fully replace deterministic actions yet, but it should establish the setup surface and provider model.

## Validation and tests

### Backend
- connector/provider configs persist correctly
- invalid credentials/configs are rejected clearly where validation exists
- status payloads are stable

### Frontend
- connect/configure flows render correctly
- connected/disconnected states are obvious
- AI provider config is understandable

### Manual verification
- configure GitHub
- configure one docs source
- configure one AI provider
- confirm the settings surface reflects state clearly

## Definition of done

Epic 4 is done when:
- the app has a real connectors/provider setup surface
- at least GitHub, one docs source, and one AI provider are supported in MVP form
- users can configure those integrations without editing the database manually
- the app feels setup-ready for real context instead of seeded-only
- all validation checks pass

## Codex prompt for Epic 4

```text
Read AGENTS.md, README.md, phase.md, and live-workspace-v1-plan.md first.

Implement Epic 4 only: Connectors and provider setup.

Goal:
Add the minimum connector/provider setup surfaces needed so Cleartify can operate on real user context instead of seeded-only data.

Scope:
- connectors/settings page
- GitHub support
- one docs source (Notion or Confluence)
- one AI provider (OpenAI or Anthropic)
- configuration and status display

Requirements:
- keep the implementation minimal and setup-oriented
- do not build deep sync jobs yet
- do not build a full secrets platform
- do not integrate consumer chat subscriptions
- keep changes small and clean

Return:
1. implementation summary
2. files changed
3. checks run
4. manual verification
5. whether Epic 4 is done
6. remaining items, if any
```

---

# Epic 5 — Skills / Playbooks

## Why this epic exists

The app needs reusable operational behavior. Teams often have recurring checklists or playbooks they want attached to certain work. This is the right way to make “skills” real before building a complicated agent platform.

## Goal

Introduce **Skills** as reusable checklists/playbooks that can be attached to projects or work types and run manually first.

## Product naming recommendation

Internally you can keep “Skill,” but in the UI consider calling them:
- Skills
- Playbooks
- Checklists

For MVP, the simplest understandable label is probably:
**Skills (checklists for repeated work)**

## In scope

- create a Skill entity
- define checklist steps/items
- attach a Skill to a Project or Work Object type/category
- run a Skill manually on a Work Object
- track completion of checklist items

## Out of scope

- autonomous agent execution
- marketplace
- versioned skill engine
- conditional branching workflow system
- external automation platform

## Example skills

- Release readiness checklist
- Bug triage checklist
- PR review checklist
- Design handoff checklist
- QA regression checklist

## User experience

A user should be able to:
- create a Skill with a name and checklist items
- attach it to a project or a type of work
- run it from a Work Object
- see checklist progress
- mark items done

## Data/model changes

Minimal MVP shape:
- `skills`
- `skill_items`
- `work_object_skill_runs`
- `work_object_skill_run_items`

Keep it simple. Do not build a generic workflow engine.

## Backend work

- CRUD for Skills
- attach/detach Skill to Project or Work Object type
- create a Skill run on a Work Object
- persist checklist item completion

## Frontend work

- Skills list/detail/editor
- attach Skill within project settings or work setup
- Work Object skill run view
- completion UX for checklist items

## How this connects to agent actions

Later, Skills can become input to agent actions.
For V1, they are human-driven operational checklists.

## Validation and tests

### Backend
- Skill CRUD works
- Skill attachment persists
- Work Object skill runs persist correctly
- item completion persists

### Frontend
- Skill creation/edit works
- Skill can be attached to a project or work pattern
- Work Object checklist progress works end to end

### Manual verification
- create a release readiness skill
- attach it to a project
- run it on a Work Object
- complete checklist items
- reload and confirm progress persists

## Definition of done

Epic 5 is done when:
- Skills exist as reusable checklists/playbooks
- a Skill can be created and attached in a meaningful MVP way
- a Work Object can run a Skill manually
- checklist progress persists and is understandable
- the app feels more operational, not just descriptive
- all validation checks pass

## Codex prompt for Epic 5

```text
Read AGENTS.md, README.md, phase.md, and live-workspace-v1-plan.md first.

Implement Epic 5 only: Skills / Playbooks.

Goal:
Add reusable checklist-style Skills that can be attached to Projects or Work Objects and run manually.

Scope:
- Skill entity and checklist items
- Skill CRUD
- attach Skill in an MVP-meaningful way
- manual Skill run on a Work Object
- persisted checklist progress

Requirements:
- keep this checklist/playbook-focused
- do not build autonomous agent execution
- do not build a workflow engine or marketplace
- keep changes small and clean

Return:
1. implementation summary
2. files changed
3. checks run
4. manual verification
5. whether Epic 5 is done
6. remaining items, if any
```

---

## Suggested overall success criteria for Live Workspace V1

Live Workspace V1 is complete when:
- real users can choose sample vs live mode
- live workspaces do not depend on seeded demo data
- projects organize the main workflow
- onboarding configures the workspace instead of explaining it
- basic connectors/provider setup exists
- skills/checklists make repeated work operational
- the product feels believable for a small real team

---

## Final implementation note

Work through these epics one at a time.
Do not let Codex combine them into one giant implementation batch.
At the start of each epic, restate the Definition of Done and require Codex to explicitly list remaining items instead of saying the epic is complete too early.
