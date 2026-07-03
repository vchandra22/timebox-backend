package database

const QueryCreateUser = `
	INSERT INTO users (full_name, email, password_hash, timezone)
	VALUES ($1, $2, $3, $4)
	RETURNING id, full_name, email, password_hash, timezone, avatar_url, email_verified_at, created_at, updated_at
`

const QueryFindAllUsers = `
	SELECT id, full_name, email, password_hash, timezone, avatar_url, email_verified_at, created_at, updated_at
	FROM users
	ORDER BY created_at DESC
	LIMIT $1 OFFSET $2
`

const QueryCountUsers = `
	SELECT COUNT(*)
	FROM users
`

const QueryFindUserByID = `
	SELECT id, full_name, email, password_hash, timezone, avatar_url, email_verified_at, created_at, updated_at
	FROM users
	WHERE id = $1
`

const QueryFindUserByEmail = `
	SELECT id, full_name, email, password_hash, timezone, avatar_url, email_verified_at, created_at, updated_at
	FROM users
	WHERE email = $1
`

const QueryUpdateUser = `
	UPDATE users
	SET full_name = $2, email = $3, timezone = $4, avatar_url = $5, updated_at = NOW()
	WHERE id = $1
	RETURNING id, full_name, email, password_hash, timezone, avatar_url, email_verified_at, created_at, updated_at
`

const QueryDeleteUser = `
	DELETE FROM users
	WHERE id = $1
	RETURNING id
`
