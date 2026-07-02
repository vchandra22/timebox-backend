package database

const QueryCreateUser = `
	INSERT INTO users (name, email)
	VALUES ($1, $2)
	RETURNING id, name, email, created_at, updated_at
`

const QueryFindAllUsers = `
	SELECT id, name, email, created_at, updated_at
	FROM users
	ORDER BY created_at DESC
	LIMIT $1 OFFSET $2
`

const QueryCountUsers = `
	SELECT COUNT(*)
	FROM users
`

const QueryFindUserByID = `
	SELECT id, name, email, created_at, updated_at
	FROM users
	WHERE id = $1
`

const QueryUpdateUser = `
	UPDATE users
	SET name = $2, email = $3, updated_at = NOW()
	WHERE id = $1
	RETURNING id, name, email, created_at, updated_at
`

const QueryDeleteUser = `
	DELETE FROM users
	WHERE id = $1
	RETURNING id
`
