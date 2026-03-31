INSERT INTO draft_plans (work_object_id, overview, implementation_steps_json, edge_cases_json, open_questions_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(work_object_id) DO UPDATE SET
  overview = excluded.overview,
  implementation_steps_json = excluded.implementation_steps_json,
  edge_cases_json = excluded.edge_cases_json,
  open_questions_json = excluded.open_questions_json,
  updated_at = excluded.updated_at;
