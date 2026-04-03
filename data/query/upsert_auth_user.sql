INSERT INTO users (
	id,
	email,
	display_name,
	created_at
) VALUES (?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	email = excluded.email,
	display_name = excluded.display_name;
