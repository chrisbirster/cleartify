# Cleartify MVP Demo Script

## Goal

Show that Cleartify keeps product context attached to the work all the way from execution planning to release readiness.

The strongest demo path is the hero Work Object:

- `Guided onboarding refresh`

Use the seeded demo data and walk through the product in this order:

1. Work Object
2. Plan
3. Agent Workspace
4. Review
5. Releases

## Setup

Use a fresh seeded SQLite database so the flow starts clean.

Suggested local boot:

```bash
PORT=8001 \
APP_PUBLIC_URL=http://127.0.0.1:8001 \
APP_MARKETING_URL=http://127.0.0.1:4173 \
DATABASE_PATH=.tasktmp/demo.sqlite.db \
SEED_DEMO=true \
go run ./cmd/cleartify
```

Then open:

- `http://127.0.0.1:8001/work/wo_onboarding_refresh`

## Demo Narrative

### 1. Work Object

Page:

- `/work/wo_onboarding_refresh`

Talk track:

- This is the core execution object.
- It already carries the title, summary, goal, priority, and status.
- The linked `Cluster` shows the product context that created the work.
- The prototype link and prototype note keep the visual reference attached to execution.
- `Decision`s and `Artifact`s show judgment and output on the same object instead of scattering them across tools.

Call out:

- `Guided onboarding refresh`
- linked `Cluster`: `Onboarding activation clarity`
- prototype note covering the five-screen slice
- seeded `Decision`: `Ship the five-screen onboarding slice first`
- seeded `Artifact`: `Onboarding scope review notes`

### 2. Plan

Page:

- `/plans/wo_onboarding_refresh`

Action:

- Click `Generate draft plan` if needed

Talk track:

- Prototype-to-Plan turns the current Work Object and prototype context into a persisted draft.
- The plan is deterministic in the MVP, but the product shape is already clear.
- Humans can refine the plan without losing the chain back to the Work Object.

Call out:

- overview
- implementation steps
- edge cases
- open questions

### 3. Agent Workspace

Page:

- `/agents/wo_onboarding_refresh`

Action:

- Click `Generate implementation brief`
- Optionally click `Generate task breakdown`

Talk track:

- Agent Workspace reads persisted product context instead of starting from an empty chat box.
- The MVP uses deterministic actions, but the important thing is the shape: context in, output out, both attached to the Work Object.

Call out:

- persisted draft plan summary
- generated `implementation_brief`
- optional `task_breakdown`

### 4. Review

Page:

- `/review/wo_onboarding_refresh`

Action:

- Click `Mark release-ready`

Talk track:

- Review is the final human checkpoint.
- The page brings together the Work Object, linked Cluster context, draft plan summary, `Decision`s, and `Artifact`s in one place.
- This is where the team confirms the execution story is coherent before moving toward release.

Call out:

- current status and priority
- draft plan summary
- seeded `Decision` and `Artifact`
- persisted `release_ready_at`

### 5. Releases

Page:

- `/releases`

Talk track:

- Once a Work Object is marked release-ready, it moves into the simple release queue.
- The MVP does not overbuild this into a release engine.
- It proves the context chain can reach a real release-facing checkpoint.

Call out:

- `Guided onboarding refresh`
- `Onboarding activation clarity`
- release-ready timestamp

## Suggested Short Pitch

> Cleartify is the system of record for product execution context.  
> It turns messy input into structured work, structured plans, agent outputs, and a final review-ready release flow without losing the why behind the work.

## Suggested Live Demo Timing

- Work Object: 60-90 seconds
- Plan: 45-60 seconds
- Agent Workspace: 45-60 seconds
- Review: 45-60 seconds
- Releases: 20-30 seconds

Total:

- about 4-5 minutes

## Demo Tips

- Use a fresh seeded database before each demo.
- Stick to the hero path unless someone asks to branch out.
- Prefer one generated agent output live; generate the second only if there is time.
- If questions go deep on AI, anchor back to the product thesis:
  - context is the source of truth
  - agent actions operate on persisted context
