SELECT id, COALESCE(cluster_id, ''), title, summary, goal, status, priority, owner_type, owner_id, COALESCE(prototype_url, ''), COALESCE(prototype_note, ''), created_at, updated_at
FROM work_objects
ORDER BY created_at ASC;
