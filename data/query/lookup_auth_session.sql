SELECT
	sessions.id,
	users.id,
	users.email,
	users.display_name,
	COALESCE(users.onboarding_completed_at, ''),
	COALESCE(sessions.refresh_token, ''),
	COALESCE(sessions.workos_session_id, ''),
	COALESCE(sessions.organization_id, ''),
	COALESCE(sessions.organization_name, ''),
	COALESCE(sessions.access_token_expires_at, '')
FROM sessions
JOIN users ON users.id = sessions.user_id
WHERE sessions.id = ?;
