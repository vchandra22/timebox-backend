package database

const timeboxSelect = `
	SELECT tb.id, tb.workspace_id, tb.task_id::text AS task_id, tb.owner_id::text AS owner_id, tb.category_id::text AS category_id,
		tb.title, tb.description, tb.scheduled_start, tb.scheduled_end, tb.planned_minutes, tb.actual_minutes, tb.status, tb.is_buffer,
		u.full_name AS owner_name, u.avatar_url AS owner_avatar, c.name AS category_name, c.color AS category_color, t.title AS task_title,
		tb.created_at, tb.updated_at
	FROM timeboxes tb
	JOIN users u ON u.id = tb.owner_id
	LEFT JOIN categories c ON c.id = tb.category_id
	LEFT JOIN tasks t ON t.id = tb.task_id
`

const QueryListTimeboxes = timeboxSelect + `
	WHERE tb.workspace_id = $1 AND tb.deleted_at IS NULL
		AND ($2::timestamptz IS NULL OR (tb.scheduled_start >= $2 AND tb.scheduled_start < $2 + INTERVAL '1 day'))
		AND ($3::timestamptz IS NULL OR tb.scheduled_start >= $3)
		AND ($4::timestamptz IS NULL OR tb.scheduled_start <= $4)
		AND ($5 = '' OR tb.owner_id::text = $5)
		AND ($6 = '' OR tb.status = $6)
		AND ($7 = '' OR tb.category_id::text = $7)
		AND ($8 = '' OR t.goal_id::text = $8)
	ORDER BY tb.scheduled_start ASC
`

const QueryCreateTimebox = `
	INSERT INTO timeboxes (workspace_id, task_id, owner_id, category_id, title, description, scheduled_start, scheduled_end, planned_minutes, is_buffer)
	SELECT $1, $2, $3, $4, COALESCE(NULLIF($5, ''), t.title), $6, $7, $8, EXTRACT(EPOCH FROM ($8 - $7))::int / 60, $9
	FROM (SELECT 1) seed
	LEFT JOIN tasks t ON t.id = $2 AND t.workspace_id = $1
	WHERE EXISTS (SELECT 1 FROM workspace_members WHERE workspace_id = $1 AND user_id = $3 AND status = 'active')
		AND ($2::uuid IS NULL OR t.id IS NOT NULL)
		AND ($4::uuid IS NULL OR EXISTS (SELECT 1 FROM categories WHERE id = $4 AND workspace_id = $1 AND deleted_at IS NULL))
	RETURNING id
`

const QueryFindTimebox = timeboxSelect + `
	WHERE tb.id = $1 AND tb.deleted_at IS NULL
`

const QueryUpdateTimebox = `
	UPDATE timeboxes
	SET task_id = $2, owner_id = $3, category_id = $4, title = $5, description = $6, scheduled_start = $7, scheduled_end = $8,
		planned_minutes = EXTRACT(EPOCH FROM ($8 - $7))::int / 60, status = $9, is_buffer = $10, updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
		AND EXISTS (SELECT 1 FROM workspace_members WHERE workspace_id = timeboxes.workspace_id AND user_id = $3 AND status = 'active')
		AND ($2::uuid IS NULL OR EXISTS (SELECT 1 FROM tasks WHERE id = $2 AND workspace_id = timeboxes.workspace_id))
		AND ($4::uuid IS NULL OR EXISTS (SELECT 1 FROM categories WHERE id = $4 AND workspace_id = timeboxes.workspace_id AND deleted_at IS NULL))
	RETURNING id
`

const QueryDeleteTimebox = `
	UPDATE timeboxes
	SET deleted_at = NOW(), status = 'cancelled', updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
	RETURNING id
`

const QueryUpdateTimeboxStatus = `
	UPDATE timeboxes
	SET status = $2, actual_minutes = $3, updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
	RETURNING id
`

const QueryCreateTimeLog = `
	INSERT INTO time_logs (timebox_id, workspace_id, started_at, ended_at, duration_seconds, source, note, created_by)
	SELECT $1, workspace_id, $2, $3, GREATEST(EXTRACT(EPOCH FROM (COALESCE($3, $2) - $2))::int, 0), $4, $5, $6
	FROM timeboxes
	WHERE id = $1 AND deleted_at IS NULL
	RETURNING id, timebox_id, workspace_id, started_at, ended_at, duration_seconds, source, note, created_by
`

const QueryCloseRunningLog = `
	UPDATE time_logs
	SET ended_at = $2, duration_seconds = GREATEST(EXTRACT(EPOCH FROM ($2 - started_at))::int, 0), note = COALESCE($3, note)
	WHERE timebox_id = $1 AND ended_at IS NULL
`

const QueryListTimeLogs = `
	SELECT id, timebox_id, workspace_id, started_at, ended_at, duration_seconds, source, note, created_by
	FROM time_logs
	WHERE timebox_id = $1
	ORDER BY started_at ASC
`

const QuerySumTimeLogSeconds = `
	SELECT COALESCE(SUM(duration_seconds), 0)
	FROM time_logs
	WHERE timebox_id = $1
`
