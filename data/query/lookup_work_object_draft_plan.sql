SELECT work_object_id, overview, implementation_steps_json, edge_cases_json, open_questions_json, created_at, updated_at
FROM draft_plans
WHERE work_object_id = ?;
