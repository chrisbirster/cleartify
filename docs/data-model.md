# Cleartify Data Model

## Core entities

### Signal
Raw product input from a source.
Fields:
- id
- source_type
- source_id
- title
- body
- url
- metadata_json
- created_at

### Cluster
A grouped theme of related signals.
Fields:
- id
- title
- summary
- severity
- status
- created_at
- updated_at

### WorkObject
The main execution object.
Fields:
- id
- title
- summary
- goal
- status
- priority
- owner_type
- owner_id
- prototype_url
- created_at
- updated_at

### Decision
Recorded judgment tied to a WorkObject.
Fields:
- id
- work_object_id
- title
- summary
- rationale
- created_at

### Artifact
Generated or attached execution artifact.
Fields:
- id
- work_object_id
- artifact_type
- title
- body
- metadata_json
- created_at

## Relationships
- many Signals belong to many Clusters
- one Cluster can spawn zero or more WorkObjects
- one WorkObject has many Decisions
- one WorkObject has many Artifacts