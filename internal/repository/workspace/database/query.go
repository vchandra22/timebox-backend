package database

const QueryCreateWorkspace = `
	INSERT INTO workspaces (name, slug, timezone, logo_url, owner_id, status, leaderboard_enabled, default_planner_interval_minutes)
	VALUES ($1, $2, $3, $4, $5, 'active', FALSE, 30)
	RETURNING id, name, slug, timezone, logo_url, owner_id, ''::text AS owner_name, 'owner'::text AS role, 1::int AS member_count, status, leaderboard_enabled, default_planner_interval_minutes, created_at, updated_at
`

const QueryCreateOwnerMember = `
	INSERT INTO workspace_members (workspace_id, user_id, role, status)
	VALUES ($1, $2, 'owner', 'active')
`

const QueryListWorkspaces = `
	SELECT w.id, w.name, w.slug, w.timezone, w.logo_url, w.owner_id, u.full_name AS owner_name, wm.role,
		COUNT(am.user_id)::int AS member_count, w.status, w.leaderboard_enabled, w.default_planner_interval_minutes, w.created_at, w.updated_at
	FROM workspaces w
	JOIN users u ON u.id = w.owner_id
	JOIN workspace_members wm ON wm.workspace_id = w.id AND wm.user_id = $1 AND wm.status = 'active'
	LEFT JOIN workspace_members am ON am.workspace_id = w.id AND am.status = 'active'
	WHERE ($2 = '' OR w.name ILIKE '%' || $2 || '%' OR w.slug ILIKE '%' || $2 || '%')
		AND ($3 = '' OR w.status = $3)
	GROUP BY w.id, u.full_name, wm.role
	ORDER BY w.created_at DESC
	LIMIT $4 OFFSET $5
`

const QueryCountWorkspaces = `
	SELECT COUNT(*)
	FROM workspaces w
	JOIN workspace_members wm ON wm.workspace_id = w.id AND wm.user_id = $1 AND wm.status = 'active'
	WHERE ($2 = '' OR w.name ILIKE '%' || $2 || '%' OR w.slug ILIKE '%' || $2 || '%')
		AND ($3 = '' OR w.status = $3)
`

const QueryFindWorkspaceByID = `
	SELECT w.id, w.name, w.slug, w.timezone, w.logo_url, w.owner_id, u.full_name AS owner_name, ''::text AS role,
		COUNT(am.user_id)::int AS member_count, w.status, w.leaderboard_enabled, w.default_planner_interval_minutes, w.created_at, w.updated_at
	FROM workspaces w
	JOIN users u ON u.id = w.owner_id
	LEFT JOIN workspace_members am ON am.workspace_id = w.id AND am.status = 'active'
	WHERE w.id = $1
	GROUP BY w.id, u.full_name
`

const QueryUpdateWorkspace = `
	WITH updated AS (
		UPDATE workspaces
		SET name = $2, slug = $3, timezone = $4, logo_url = $5, leaderboard_enabled = $6, default_planner_interval_minutes = $7, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, slug, timezone, logo_url, owner_id, status, leaderboard_enabled, default_planner_interval_minutes, created_at, updated_at
	)
	SELECT updated.id, updated.name, updated.slug, updated.timezone, updated.logo_url, updated.owner_id, ''::text AS owner_name, ''::text AS role,
		(SELECT COUNT(*)::int FROM workspace_members WHERE workspace_id = updated.id AND status = 'active') AS member_count,
		updated.status, updated.leaderboard_enabled, updated.default_planner_interval_minutes, updated.created_at, updated.updated_at
	FROM updated
`

const QueryFindMember = `
	SELECT wm.workspace_id, wm.user_id, u.full_name, u.email, u.avatar_url, wm.role, wm.status, wm.joined_at, wm.updated_at
	FROM workspace_members wm
	JOIN users u ON u.id = wm.user_id
	WHERE wm.workspace_id = $1 AND wm.user_id = $2
`

const QueryListMembers = `
	SELECT wm.workspace_id, wm.user_id, u.full_name, u.email, u.avatar_url, wm.role, wm.status, wm.joined_at, wm.updated_at
	FROM workspace_members wm
	JOIN users u ON u.id = wm.user_id
	WHERE wm.workspace_id = $1
		AND ($2 = '' OR wm.role = $2)
		AND ($3 = '' OR wm.status = $3)
		AND ($4 = '' OR u.full_name ILIKE '%' || $4 || '%' OR u.email ILIKE '%' || $4 || '%')
	ORDER BY wm.joined_at DESC
	LIMIT $5 OFFSET $6
`

const QueryCountMembers = `
	SELECT COUNT(*)
	FROM workspace_members wm
	JOIN users u ON u.id = wm.user_id
	WHERE wm.workspace_id = $1
		AND ($2 = '' OR wm.role = $2)
		AND ($3 = '' OR wm.status = $3)
		AND ($4 = '' OR u.full_name ILIKE '%' || $4 || '%' OR u.email ILIKE '%' || $4 || '%')
`

const QueryListMemberTeams = `
	SELECT tm.user_id, t.id, t.name
	FROM team_members tm
	JOIN teams t ON t.id = tm.team_id
	WHERE tm.workspace_id = $1 AND tm.user_id = $2 AND tm.deleted_at IS NULL AND t.deleted_at IS NULL
	ORDER BY t.name
`

const QueryCreateInvitation = `
	INSERT INTO workspace_invitations (workspace_id, email, role, status, expires_at)
	VALUES ($1, $2, $3, 'invited', $4)
	RETURNING id, workspace_id, email, role, status, expires_at
`

const QueryFindUserIDByEmail = `
	SELECT id
	FROM users
	WHERE email = $1
`

const QueryUpsertInvitedMember = `
	INSERT INTO workspace_members (workspace_id, user_id, role, status)
	VALUES ($1, $2, $3, 'invited')
	ON CONFLICT (workspace_id, user_id)
	DO UPDATE SET role = EXCLUDED.role, status = 'invited', updated_at = NOW()
`

const QueryUpdateMember = `
	UPDATE workspace_members
	SET role = $3, status = $4, updated_at = NOW()
	WHERE workspace_id = $1 AND user_id = $2
	RETURNING workspace_id, user_id, ''::text AS full_name, ''::text AS email, NULL::text AS avatar_url, role, status, joined_at, updated_at
`

const QuerySoftDeleteTeamMembersForUser = `
	UPDATE team_members
	SET deleted_at = NOW(), updated_at = NOW()
	WHERE workspace_id = $1 AND user_id = $2 AND deleted_at IS NULL
`

const QueryUpsertTeamMember = `
	INSERT INTO team_members (team_id, workspace_id, user_id, status, deleted_at)
	SELECT id, workspace_id, $2, 'active', NULL
	FROM teams
	WHERE id = $1 AND workspace_id = $3 AND deleted_at IS NULL
	ON CONFLICT (team_id, user_id)
	DO UPDATE SET status = 'active', deleted_at = NULL, updated_at = NOW()
`

const QueryListTeams = `
	SELECT t.id, t.workspace_id, t.name, t.description, COUNT(tm.user_id)::int AS member_count, t.created_at, t.updated_at
	FROM teams t
	LEFT JOIN team_members tm ON tm.team_id = t.id AND tm.deleted_at IS NULL
	WHERE t.workspace_id = $1 AND t.deleted_at IS NULL
	GROUP BY t.id
	ORDER BY t.created_at DESC
`

const QueryCreateTeam = `
	INSERT INTO teams (workspace_id, name, description)
	VALUES ($1, $2, $3)
	RETURNING id, workspace_id, name, description, 0::int AS member_count, created_at, updated_at
`

const QueryFindTeam = `
	SELECT t.id, t.workspace_id, t.name, t.description, COUNT(tm.user_id)::int AS member_count, t.created_at, t.updated_at
	FROM teams t
	LEFT JOIN team_members tm ON tm.team_id = t.id AND tm.deleted_at IS NULL
	WHERE t.id = $1 AND t.deleted_at IS NULL
	GROUP BY t.id
`

const QueryUpdateTeam = `
	UPDATE teams
	SET name = $2, description = $3, updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
	RETURNING id, workspace_id, name, description, 0::int AS member_count, created_at, updated_at
`

const QueryDeleteTeam = `
	UPDATE teams
	SET deleted_at = NOW(), updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
	RETURNING id
`

const QuerySoftDeleteTeamMembers = `
	UPDATE team_members
	SET deleted_at = NOW(), updated_at = NOW()
	WHERE team_id = $1 AND deleted_at IS NULL
`
