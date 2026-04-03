# Cleartify Routes

## Marketing
- / -> marketing site

## App
- / -> home / today
- /inbox -> context inbox
- /inbox/:clusterId -> cluster detail
- /work -> work objects list
- /work/:workObjectId -> work object detail
- /prototypes -> prototype list
- /prototypes/new -> prototype import / attach
- /plans/:workObjectId -> generated plan view
- /agents/:workObjectId -> agent workspace
- /review/:workObjectId -> review / diff / approval
- /releases -> release list
- /releases/:releaseId -> release detail

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
