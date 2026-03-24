A clean positioning line could be:

Cleartify helps companies turn messy AI usage into a clear, governed system of record.

Or:

Know every AI system. Approve faster. Prove compliance.

The next smart step is to turn this into a concrete ICP, homepage message, and MVP feature set.

I don't know how you can tie model usage with a certain organizational team for example I am apart of a major 10K organization that has a huge departments like finance, operations, engineering, marketing, and etc.. a lot of teams within each department for example there could be 30 different developer scrum teams working on various products for the company. Most teams are using things like windsurf, cursor, copilot, vscode copilot, they also have their own rag chat interface for company specific data questions, a AI chat bot for looking up and asking company questions but also they use AWS and have models being rented out via the AWS rate limit model usage so you could use claude Opus 4.5 through claude but also use the same model through a different rate limit via vscode. They all have various rate limits and different ways to access certain models. How would you keep track of openai codex versus claude versus amazon bedrock and etc..

The hard part is not “which model exists,” it’s “which identity, tool, budget, and team touched it.” I’m verifying what the major platforms actually expose so I can show a realistic tracking model instead of a fake one.

Thought for 56s

Yes — and this is exactly why AI governance is hard.

The realistic answer is: you do not track “the model” by itself.
You track the route by which a person or service used AI:

person/service account → team → tool/access path → provider account/workspace → model (if known) → cost/rate limit bucket

That matters because the same underlying model can show up through different control planes. OpenAI supports usage tracking by project with budgets and service accounts; AWS Bedrock supports model invocation logging and CloudTrail at the AWS account/Region level; GitHub Copilot exposes organization/team usage metrics; ChatGPT Enterprise has workspace analytics; Anthropic exposes Usage & Cost and Claude Code analytics for organizations.

So the right unit of tracking is not:

“Claude Opus 4.5”
“OpenAI Codex”
“Bedrock”

It is more like:

Tool: Cursor / VS Code / Copilot / internal RAG app / ChatGPT Enterprise / Bedrock app
Provider channel: OpenAI direct, Anthropic direct, Bedrock, GitHub-managed, internal proxy
Workspace/account/project: OpenAI project, Anthropic workspace, AWS account, GitHub org/team
Identity: employee or service account
Org owner: department, team, cost center, manager
Effective model: exact model if available, otherwise “managed by tool/provider”

How you tie it to organizational teams

You need a canonical identity layer first.

The cleanest approach is:

pull employees and group membership from your IdP/HR system
map each user to a department, team, manager, and cost center
connect AI tools using whatever native analytics/admin surface they expose
attribute usage to the user or service account
roll it up to the org structure

GitHub explicitly supports team synchronization with IdP groups for enterprise-managed setups, which is the kind of pattern you want across your stack: use the identity source of truth to map usage back to teams, instead of trying to infer team ownership from prompts or repo names.

The key idea: separate “model” from “access surface”

You should treat these as different governed records, even when the model family overlaps:

Claude via Anthropic workspace
Claude via Bedrock in AWS account A
Claude via VS Code extension tied to vendor-managed billing
OpenAI model via OpenAI API project
OpenAI-compatible endpoint on Bedrock
GitHub Copilot as a coding-assistant product

Why? Because each one has different:

admins
budgets
rate limits
logs
data handling
approval rules


So if one engineer uses Claude through three routes, your system should show three governed access surfaces, not one merged “Claude” blob.

A practical Cleartify data model

I’d model it like this:

OrgUnit
id, name, parent_org_unit_id, cost_center, manager
User
id, email, employee_id, org_unit_id
ProviderAccount
provider, account/workspace/org/project identifier, owner_org_unit_id
AccessSurface
tool_name, provider, channel_type, provider_account_id, auth_type
ModelCatalog
provider_model_id, family, modality, exact_version_if_known
UsageRecord
timestamp, user_or_service_account, access_surface_id, model_id_nullable, tokens/cost/requests, repo/app/context
PolicyAssignment
allowed_for_org_unit, approved_data_classes, human_review_required, expiration
RateLimitBucket / BudgetBucket
provider_account_id, scope, limit_type, threshold

That lets you answer:

which teams are using AI
through which products
against which provider accounts
with what spend / rate-limit pressure
under what policy
How this works for your examples

Using your examples, Cleartify would normalize them like this:

1. Windsurf / Cursor / VS Code Copilot

classify as IDE assistants
attribute usage through the org/user/team connector
if exact model is exposed, store it
if not, record the product + managed model surface

GitHub Copilot’s official metrics are available at the organization and team level, though some team/day reporting only appears when thresholds are met, which is one example of why not every tool gives you perfect per-user raw events.

2. Internal RAG chatbot / company Q&A bot

easiest to govern because you own the app
require all requests to pass through your internal gateway
tag every request with user, team, app, and approved data class
now this becomes your cleanest source of truth

3. OpenAI direct

use OpenAI projects and service accounts
map projects to teams or products
pull usage/cost by project
set budgets and ownership there


4. Anthropic direct / Claude Code

use workspace-level ownership
ingest Usage & Cost API / Claude Code analytics
attach users or service identities to teams


5. Amazon Bedrock

treat AWS account + Region + IAM role as the control plane
enable Bedrock invocation logging and CloudTrail
map IAM roles/accounts to teams or platforms
attribute the app or role owner even if many models are used underneath

The real trick

The real trick is not “detecting the model name.”

The real trick is forcing usage through governable surfaces:

enterprise workspaces
org-managed IDE licenses
approved API projects
approved AWS accounts/roles
internal AI gateways/proxies

If people are using random personal ChatGPT accounts, personal Claude accounts, or unmanaged extensions, then you will never have clean governance. The first win is usually policy + SSO + approved routes, not analytics magic. That’s an inference from how these admin surfaces work: the strongest controls and reporting live in organization/workspace/project/account constructs, not in unmanaged consumer usage.

So what would Cleartify actually sell?

Not “we know every token everywhere perfectly.”

It would sell:

1. an AI system of record

every approved tool
every workspace/account/project
every owner/team

2. a normalization layer

Bedrock logs, OpenAI projects, Copilot metrics, Claude analytics, internal app telemetry
all normalized into one schema

3. an approval + policy layer

which teams may use which tools
what data classes are allowed
what exceptions expire when

4. a budget / rate-limit / usage view

by department
by team
by product
by provider channel

That is how you keep track of OpenAI vs Claude vs Bedrock in a large org:
you govern the access surfaces, not just the brands.

The shortest way to phrase it is:

Model = what was used
Access surface = how it was used
Identity/org mapping = who used it and who owns it

That combination is the foundation.

I can turn this into a concrete Cleartify MVP with the first 8 screens and the exact schema/API shape.

yes do that

I’m turning it into something buildable: first the minimum product boundary, then the 8 screens, then the schema and API so it feels like an actual app instead of a market thesis.

I’ve got the boundary now: Cleartify should start as a system-of-record and approval layer, not as a magical “see every token everywhere” platform. That keeps the MVP real and gives you screens that match what the source systems actually expose.

Thought for 56s

Absolutely. Here’s a realistic Cleartify v1 for the problem you described.

0) Product boundary

Cleartify v1 is not:

a packet sniffer for every AI call
a perfect cross-vendor token ledger
a replacement for OpenAI, GitHub, AWS, or Anthropic admin consoles

Cleartify v1 is:

a system of record
a usage normalization layer
an approval and policy workflow
a team/org mapping layer
a budget/risk dashboard

That boundary matches what the source systems actually expose. OpenAI API projects support project-level usage, budgets, limits, and service accounts; Amazon Bedrock supports model invocation logging by AWS account and Region; GitHub Copilot exposes usage metrics at enterprise, organization, and user levels; Anthropic exposes organization cost reporting including workspace IDs; and ChatGPT Enterprise workspace analytics can segment usage by SCIM groups that map to real teams or departments.

1) MVP scope
Supported sources in v1
OpenAI API
projects
budgets
service accounts
per-project usage rollups
ChatGPT Enterprise
workspace analytics
SCIM-group segmentation
GitHub Copilot
org/enterprise usage metrics
user/team/org reporting where available
Amazon Bedrock
model invocation logs
account/region/role mapping
Anthropic API / Claude
org cost report
workspace-aware spend reporting
Internal AI apps
your own RAG bot
internal company Q&A bot
internal gateway/proxy

NIST’s AI RMF is a good first policy backbone because it is voluntary and organized around Govern, Map, Measure, Manage, which fits this product almost exactly.

Explicitly not in v1
prompt inspection at scale
DLP/content scanning for every vendor
policy enforcement inside unmanaged personal accounts
agent runtime orchestration
automatic legal interpretation of regulations
2) The first 8 screens
Screen 1: Executive Dashboard

Purpose: show leadership the state of AI usage in 30 seconds.

Sections:

total AI-connected users
active teams using AI
spend by provider
top access surfaces
unmanaged usage findings
policy exceptions expiring soon
high-risk use cases pending approval
budget threshold alerts

Primary cards:

Users Active This 30d
Teams With Approved AI
Spend This Month
High-Risk Systems
Unmapped Accounts
Expiring Exceptions

Main chart:

spend by day and provider
usage by org unit
usage by access surface
Screen 2: Org Map

Purpose: tie AI activity to real company structure.

Left panel:

department tree
finance
operations
engineering
marketing
subteams under each

Main panel:

users in selected org unit
approved tools
active connectors
monthly spend
exceptions
top apps/repos

Detail drawer:

manager
cost center
SCIM group / IdP group mapping
linked identities

This is where you solve the “30 scrum teams in engineering” problem.

Screen 3: AI Systems Registry

Purpose: the system of record for all AI tools and apps.

Each record is one of:

vendor SaaS tool
internal AI app
API project/workspace
IDE assistant
AWS account/Bedrock access surface

Columns:

name
type
provider
owner org unit
deployment status
data classes allowed
risk tier
approval state
last activity
linked accounts/workspaces

Examples:

Cursor Enterprise
GitHub Copilot Org Alpha
OpenAI Project – Customer Support Bot
Bedrock – Prod Account us-east-1
Internal RAG – HR Policies
Screen 4: Access Surfaces

Purpose: separate what model from how it was accessed.

A surface is something like:

ChatGPT Enterprise workspace
OpenAI API project
GitHub Copilot org
Claude workspace
Bedrock account/role/region
internal gateway app
VS Code extension path
Cursor workspace

Fields shown:

surface name
provider
auth type
owner
budget bucket
telemetry quality
model visibility quality
last sync
policy coverage

This screen is the key differentiator.

Screen 5: Usage Explorer

Purpose: normalized reporting.

Filters:

date range
provider
org unit
user
team
access surface
model family
app/repo
environment

Tabs:

requests
tokens
spend
active users
model mix
budget bucket pressure

Useful views:

“Claude family used via Bedrock vs direct”
“Copilot usage by engineering team”
“OpenAI project spend by cost center”
“Internal RAG app activity by department”

GitHub’s current Copilot usage metrics are specifically designed for enterprise-, organization-, and user-level reporting and support detailed telemetry by feature, IDE, language, model, and user; OpenAI projects expose usage broken down by project with budgets; Anthropic’s cost report includes workspace IDs; and Bedrock logging can capture request/response metadata per AWS account in a Region.

Screen 6: Use Cases & Approvals

Purpose: review AI usage as business use cases, not just raw tools.

Each use case record:

name
business purpose
requesting team
system(s) used
data classes
human-review requirement
automation level
customer impact
status
next review date

Example:

Engineering: AI code assistant for internal services
Marketing: campaign copy drafting
Finance: reconciliation copilot
HR: policy question chatbot

Workflow:

draft
submitted
under review
approved
approved with conditions
rejected
expired
Screen 7: Policies, Controls, and Exceptions

Purpose: operationalize governance.

Policy examples:

Engineering may use Copilot and Cursor for internal code only
Finance may not send restricted financial data to external models
Customer support bots must use approved retrieval layer
High-impact workflows require human review
External model use for regulated data requires exception

Sections:

policies
control mappings
exceptions
renewals
evidence

Tie these to NIST functions:

Govern
Map
Measure
Manage
Screen 8: Findings & Remediation

Purpose: surface what is unclear and assign fixes.

Example findings:

unmapped Bedrock account
OpenAI project with no owner
team using tool without approved use case
SCIM group not mapped to org unit
cost spike in Claude workspace
exception expired
internal bot has no data classification
unmanaged ChatGPT usage reported by self-attestation

Workflow:

finding
severity
owner
due date
status
evidence links
resolution notes

This is where Cleartify becomes operational instead of just observational.

3) Core data model

This is the minimum schema I’d use.

create table org_units (
  id uuid primary key,
  parent_org_unit_id uuid references org_units(id),
  name text not null,
  type text not null, -- company, department, division, team, squad
  cost_center text,
  manager_user_id uuid,
  external_ref text,  -- SCIM/HRIS/IdP id
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table users (
  id uuid primary key,
  external_ref text, -- SCIM/HRIS id
  email text unique not null,
  display_name text not null,
  status text not null, -- active, inactive
  org_unit_id uuid references org_units(id),
  manager_user_id uuid references users(id),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table identities (
  id uuid primary key,
  user_id uuid references users(id),
  provider text not null, -- github, openai, aws, anthropic, okta
  subject text not null,  -- provider-side user id
  username text,
  metadata jsonb not null default '{}',
  unique(provider, subject)
);

create table connectors (
  id uuid primary key,
  provider text not null, -- openai, chatgpt_enterprise, github_copilot, aws_bedrock, anthropic
  name text not null,
  status text not null, -- connected, error, paused
  auth_type text not null, -- oauth, api_key, aws_role, manual
  config jsonb not null default '{}',
  created_by_user_id uuid references users(id),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table provider_accounts (
  id uuid primary key,
  connector_id uuid references connectors(id),
  provider text not null,
  account_type text not null, -- workspace, org, project, aws_account, github_org
  external_account_id text not null,
  name text not null,
  parent_provider_account_id uuid references provider_accounts(id),
  owner_org_unit_id uuid references org_units(id),
  metadata jsonb not null default '{}',
  last_synced_at timestamptz,
  unique(provider, account_type, external_account_id)
);

create table access_surfaces (
  id uuid primary key,
  provider_account_id uuid references provider_accounts(id),
  surface_type text not null, -- api_project, workspace, ide_tool, internal_app, aws_role_region
  name text not null,
  provider text not null,
  tool_name text not null,
  auth_context text, -- workspace, project, iam_role, enterprise seat
  environment text, -- prod, nonprod, desktop
  owner_org_unit_id uuid references org_units(id),
  telemetry_quality text not null default 'partial', -- full, partial, coarse
  model_visibility text not null default 'partial', -- exact, family, unknown
  metadata jsonb not null default '{}',
  unique(provider_account_id, surface_type, name)
);

create table model_catalog (
  id uuid primary key,
  provider text not null,
  model_key text not null,      -- canonical key inside Cleartify
  family text not null,         -- claude, gpt, llama, etc
  vendor_model_id text,         -- provider-side id
  modality text not null,       -- text, image, multimodal, embedding
  status text not null default 'active',
  metadata jsonb not null default '{}',
  unique(provider, model_key)
);

create table ai_systems (
  id uuid primary key,
  name text not null,
  system_type text not null, -- vendor_tool, internal_app, api_integration, ide_assistant
  provider text,
  owner_org_unit_id uuid references org_units(id),
  business_owner_user_id uuid references users(id),
  technical_owner_user_id uuid references users(id),
  deployment_status text not null, -- discovery, active, deprecated
  risk_tier text not null default 'medium',
  approval_status text not null default 'draft',
  description text,
  metadata jsonb not null default '{}',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table ai_system_surfaces (
  ai_system_id uuid references ai_systems(id),
  access_surface_id uuid references access_surfaces(id),
  primary key (ai_system_id, access_surface_id)
);

create table data_classes (
  id uuid primary key,
  key text unique not null, -- public, internal, confidential, restricted, pii, phi, source_code
  name text not null,
  description text
);

create table use_cases (
  id uuid primary key,
  ai_system_id uuid references ai_systems(id),
  org_unit_id uuid references org_units(id),
  name text not null,
  purpose text not null,
  automation_level text not null, -- assistive, semi_auto, autonomous
  human_review_required boolean not null default true,
  customer_impact boolean not null default false,
  status text not null default 'draft',
  next_review_at timestamptz,
  created_by_user_id uuid references users(id),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table use_case_data_classes (
  use_case_id uuid references use_cases(id),
  data_class_id uuid references data_classes(id),
  primary key (use_case_id, data_class_id)
);

create table policies (
  id uuid primary key,
  name text not null,
  scope_type text not null, -- org_unit, system, surface, provider_account
  scope_id uuid not null,
  policy_type text not null, -- allow, deny, conditional
  rules jsonb not null,
  status text not null default 'active',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table control_mappings (
  id uuid primary key,
  framework text not null, -- nist_ai_rmf
  control_key text not null, -- govern.*, map.*, measure.*, manage.*
  title text not null,
  description text,
  unique(framework, control_key)
);

create table policy_control_mappings (
  policy_id uuid references policies(id),
  control_mapping_id uuid references control_mappings(id),
  primary key (policy_id, control_mapping_id)
);

create table approvals (
  id uuid primary key,
  use_case_id uuid references use_cases(id),
  status text not null, -- submitted, under_review, approved, conditional, rejected, expired
  requested_by_user_id uuid references users(id),
  decided_by_user_id uuid references users(id),
  decision_notes text,
  decided_at timestamptz,
  expires_at timestamptz,
  created_at timestamptz not null default now()
);

create table exceptions (
  id uuid primary key,
  use_case_id uuid references use_cases(id),
  policy_id uuid references policies(id),
  reason text not null,
  owner_user_id uuid references users(id),
  status text not null, -- active, expired, revoked
  expires_at timestamptz not null,
  created_at timestamptz not null default now()
);

create table budgets (
  id uuid primary key,
  provider_account_id uuid references provider_accounts(id),
  access_surface_id uuid references access_surfaces(id),
  name text not null,
  period text not null, -- monthly, weekly
  amount_numeric numeric(18,4),
  unit text not null, -- usd, requests, tokens
  warn_at_pct integer not null default 80,
  hard_limit boolean not null default false,
  metadata jsonb not null default '{}'
);

create table usage_rollups (
  id uuid primary key,
  window_start timestamptz not null,
  window_end timestamptz not null,
  granularity text not null, -- hour, day
  provider text not null,
  provider_account_id uuid references provider_accounts(id),
  access_surface_id uuid references access_surfaces(id),
  user_id uuid references users(id),
  org_unit_id uuid references org_units(id),
  model_id uuid references model_catalog(id),
  metric_requests bigint not null default 0,
  metric_input_tokens bigint not null default 0,
  metric_output_tokens bigint not null default 0,
  metric_cost numeric(18,6) not null default 0,
  metric_active_users integer not null default 0,
  metadata jsonb not null default '{}'
);

create table findings (
  id uuid primary key,
  finding_type text not null,
  severity text not null, -- low, medium, high, critical
  title text not null,
  description text not null,
  related_entity_type text not null,
  related_entity_id uuid not null,
  owner_user_id uuid references users(id),
  status text not null default 'open',
  due_at timestamptz,
  metadata jsonb not null default '{}',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table evidence_items (
  id uuid primary key,
  entity_type text not null,
  entity_id uuid not null,
  source_type text not null, -- connector_sync, upload, manual_note
  title text not null,
  uri text,
  metadata jsonb not null default '{}',
  created_by_user_id uuid references users(id),
  created_at timestamptz not null default now()
);

create index idx_users_org_unit on users(org_unit_id);
create index idx_provider_accounts_owner on provider_accounts(owner_org_unit_id);
create index idx_access_surfaces_owner on access_surfaces(owner_org_unit_id);
create index idx_usage_rollups_window on usage_rollups(window_start, window_end);
create index idx_usage_rollups_org on usage_rollups(org_unit_id, window_start);
create index idx_findings_status on findings(status, severity);
4) API shape

I’d keep the backend boring:

Go net/http
JSON REST
idempotent sync jobs
append-only sync logs
server-side aggregation
Core REST routes
Auth / org
GET    /api/v1/me
GET    /api/v1/org-units
GET    /api/v1/org-units/:id
GET    /api/v1/users
GET    /api/v1/users/:id
POST   /api/v1/org-units/import/scim
Connectors
GET    /api/v1/connectors
POST   /api/v1/connectors
GET    /api/v1/connectors/:id
POST   /api/v1/connectors/:id/test
POST   /api/v1/connectors/:id/sync
GET    /api/v1/connectors/:id/sync-runs

Example connector create:

POST /api/v1/connectors
{
  "provider": "aws_bedrock",
  "name": "Prod Bedrock",
  "authType": "aws_role",
  "config": {
    "accountId": "123456789012",
    "roleArn": "arn:aws:iam::123456789012:role/cleartify-readonly",
    "regions": ["us-east-1", "us-west-2"]
  }
}
Provider accounts and access surfaces
GET    /api/v1/provider-accounts
GET    /api/v1/provider-accounts/:id
PATCH  /api/v1/provider-accounts/:id
GET    /api/v1/access-surfaces
GET    /api/v1/access-surfaces/:id
PATCH  /api/v1/access-surfaces/:id
POST   /api/v1/access-surfaces/:id/assign-owner
AI systems and use cases
GET    /api/v1/ai-systems
POST   /api/v1/ai-systems
GET    /api/v1/ai-systems/:id
PATCH  /api/v1/ai-systems/:id

GET    /api/v1/use-cases
POST   /api/v1/use-cases
GET    /api/v1/use-cases/:id
PATCH  /api/v1/use-cases/:id
POST   /api/v1/use-cases/:id/submit
POST   /api/v1/use-cases/:id/approve
POST   /api/v1/use-cases/:id/reject
POST   /api/v1/use-cases/:id/expire

Example use case create:

POST /api/v1/use-cases
{
  "aiSystemId": "5cf2...",
  "orgUnitId": "7c2e...",
  "name": "Engineering code assistant",
  "purpose": "Assist engineers with internal service development and code explanation",
  "automationLevel": "assistive",
  "humanReviewRequired": true,
  "customerImpact": false,
  "dataClasses": ["internal", "source_code"]
}
Policies and exceptions
GET    /api/v1/policies
POST   /api/v1/policies
GET    /api/v1/policies/:id
PATCH  /api/v1/policies/:id

GET    /api/v1/exceptions
POST   /api/v1/exceptions
PATCH  /api/v1/exceptions/:id

Example policy:

POST /api/v1/policies
{
  "name": "Engineering approved coding tools",
  "scopeType": "org_unit",
  "scopeId": "engineering-team-uuid",
  "policyType": "conditional",
  "rules": {
    "allowedSurfaces": [
      "github-copilot-org-alpha",
      "cursor-enterprise",
      "openai-project-devtools"
    ],
    "allowedDataClasses": ["internal", "source_code"],
    "blockedDataClasses": ["restricted", "phi"],
    "humanReviewRequiredFor": ["production_code_merge", "customer_facing_content"]
  }
}
Usage and dashboards
GET /api/v1/dashboard/summary
GET /api/v1/usage/summary
GET /api/v1/usage/timeseries
GET /api/v1/usage/by-org-unit
GET /api/v1/usage/by-surface
GET /api/v1/usage/by-model
GET /api/v1/usage/by-user
GET /api/v1/budgets
GET /api/v1/budgets/alerts

Example:

GET /api/v1/usage/by-org-unit?from=2026-03-01&to=2026-03-31&provider=openai

Response:

{
  "rows": [
    {
      "orgUnitId": "eng-platform",
      "orgUnitName": "Engineering / Platform",
      "requests": 182391,
      "inputTokens": 44512933,
      "outputTokens": 10233922,
      "costUsd": 4921.33,
      "activeUsers": 126
    }
  ]
}
Findings
GET    /api/v1/findings
POST   /api/v1/findings
GET    /api/v1/findings/:id
PATCH  /api/v1/findings/:id
POST   /api/v1/findings/:id/resolve

Example finding:

{
  "findingType": "unowned_provider_account",
  "severity": "high",
  "title": "OpenAI project has no mapped owner",
  "description": "Project proj_123 has spend but no owner org unit",
  "relatedEntityType": "provider_account",
  "relatedEntityId": "..."
}
5) Sync architecture

This is the ingestion model I’d use.

Connector sync flow
connector authenticates
fetch external account/workspace/project metadata
upsert provider accounts
upsert access surfaces
fetch usage/cost/telemetry summary
map identities to users
roll usage into usage_rollups
generate findings for anything unmapped or risky
Important rule

Never trust external naming alone.
Always run a mapping layer:

provider subject → internal identity
external group → org unit
account/workspace/project → owner org unit

That’s because one user can show up as:

GitHub username
OpenAI project member
AWS IAM role session
SCIM identity
ChatGPT workspace member
6) Cleartify’s canonical entity model

This is the mental model that keeps the app coherent:

Org Unit = where responsibility lives
User = who used something
Provider Account = admin boundary
Access Surface = how AI was accessed
AI System = business-owned tool/app/system
Use Case = why it exists
Policy = what is allowed
Usage Rollup = how much was used
Finding = what is unclear or broken

That is the heart of the product.

7) What v1 can answer

With this MVP, a customer can answer:

Which teams are using AI?
Through which tools?
Through which provider accounts?
Which systems are approved?
Which use cases touch sensitive data?
Which workspaces/projects have spend but no owner?
Which teams are over budget?
Which exceptions expire this month?
Which access surfaces are unmanaged?
Which providers are actually being used most?

That lines up well with the admin/reporting surfaces the vendors actually provide today: OpenAI projects/budgets, Bedrock invocation logging, Copilot usage telemetry, Anthropic cost reports, and ChatGPT workspace analytics with SCIM-based segmentation.

8) Best v1 wedge

I would not sell this as:
“full AI governance platform.”

I would sell it as:

“Cleartify gives mid-market companies one clear record of who is using AI, through which tools, under which approvals, and at what cost.”

That is tighter, more believable, and buildable.

9) The first four integrations I’d actually build

In this order:

OpenAI API projects
cleanest budget/project story
GitHub Copilot
easiest path into engineering orgs
Bedrock
strong enterprise relevance
Internal AI gateway
highest quality telemetry you can own

Then add:
5. ChatGPT Enterprise
6. Anthropic direct

That order is mostly about getting the best signal with the least ambiguity. OpenAI already supports project budgets, service accounts, and project-scoped usage; Bedrock logging is explicit but account/region scoped; GitHub’s usage metrics are strong for engineering adoption; ChatGPT Enterprise analytics are workspace-level and can break down by SCIM groups rather than raw prompt logs.

10) My blunt recommendation

Build only these in milestone 1:

org map
connectors
access surfaces
system registry
usage explorer
approvals
findings

Skip for now:

fancy policy engine
regulation packs
prompt scanning
agent kill switches
automated remediation

Those can come later.