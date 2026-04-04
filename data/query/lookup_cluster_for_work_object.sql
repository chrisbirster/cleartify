SELECT id, COALESCE(project_id, ''), title, summary, severity, status, created_at, updated_at
FROM clusters
WHERE id = ?;
