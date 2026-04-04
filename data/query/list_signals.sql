SELECT id, COALESCE(project_id, ''), source_type, source_id, title, body, COALESCE(url, ''), metadata_json, created_at
FROM signals
ORDER BY created_at DESC;
