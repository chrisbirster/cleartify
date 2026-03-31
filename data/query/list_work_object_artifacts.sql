SELECT id, work_object_id, artifact_type, title, body, metadata_json, created_at
FROM artifacts
WHERE work_object_id = ?
ORDER BY created_at ASC;
