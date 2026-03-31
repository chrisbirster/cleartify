SELECT s.id, s.source_type, s.source_id, s.title, s.body, COALESCE(s.url, ''), s.metadata_json, s.created_at
FROM cluster_signals cs
JOIN signals s ON s.id = cs.signal_id
WHERE cs.cluster_id = ?
ORDER BY s.created_at DESC;
