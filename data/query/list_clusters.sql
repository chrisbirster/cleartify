SELECT id, title, summary, severity, status, created_at, updated_at
FROM clusters
WHERE status != 'resolved'
ORDER BY created_at ASC;
