SELECT id, workspace_id, name, slug, summary, status, created_at, updated_at
FROM projects
WHERE id = ?;
