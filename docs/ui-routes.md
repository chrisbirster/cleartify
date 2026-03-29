# Cleartify Routes

## Marketing
- / -> marketing site

## App
- /app -> home / today
- /app/inbox -> context inbox
- /app/inbox/:clusterId -> cluster detail
- /app/work -> work objects list
- /app/work/:workObjectId -> work object detail
- /app/prototypes -> prototype list
- /app/prototypes/new -> prototype import / attach
- /app/plans/:workObjectId -> generated plan view
- /app/agents/:workObjectId -> agent workspace
- /app/review/:workObjectId -> review / diff / approval
- /app/releases -> release list
- /app/releases/:releaseId -> release detail

## API
- GET /api/signals
- GET /api/clusters
- POST /api/clusters/:id/create-work-object
- GET /api/work-objects
- POST /api/work-objects
- GET /api/work-objects/:id
- POST /api/work-objects/:id/generate-plan
- POST /api/work-objects/:id/generate-artifacts
- GET /api/work-objects/:id/artifacts
- POST /api/prototypes