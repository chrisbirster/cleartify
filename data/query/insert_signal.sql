INSERT INTO signals (
  id,
  project_id,
  source_type,
  source_id,
  title,
  body,
  url,
  metadata_json,
  created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
