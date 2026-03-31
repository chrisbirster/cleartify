SELECT id, work_object_id, title, summary, rationale, created_at
FROM decisions
WHERE work_object_id = ?
ORDER BY created_at ASC;
