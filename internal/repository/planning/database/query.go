package database

const QueryListCategories = `
	SELECT id, workspace_id, name, color, is_default, created_at, updated_at
	FROM categories
	WHERE workspace_id = $1 AND deleted_at IS NULL
	ORDER BY is_default DESC, name ASC
`

const QueryCreateCategory = `
	INSERT INTO categories (workspace_id, name, color)
	VALUES ($1, $2, $3)
	RETURNING id, workspace_id, name, color, is_default, created_at, updated_at
`

const QueryFindCategory = `
	SELECT id, workspace_id, name, color, is_default, created_at, updated_at
	FROM categories
	WHERE id = $1 AND deleted_at IS NULL
`

const QueryUpdateCategory = `
	UPDATE categories
	SET name = $2, color = $3, updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
	RETURNING id, workspace_id, name, color, is_default, created_at, updated_at
`

const QueryDeleteCategory = `
	UPDATE categories
	SET deleted_at = NOW(), updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
	RETURNING id
`

const QueryListTags = `
	SELECT id, workspace_id, name, created_at, updated_at
	FROM tags
	WHERE workspace_id = $1 AND deleted_at IS NULL
		AND ($2 = '' OR name ILIKE '%' || $2 || '%')
	ORDER BY name ASC
`

const QueryCreateTag = `
	INSERT INTO tags (workspace_id, name)
	VALUES ($1, $2)
	RETURNING id, workspace_id, name, created_at, updated_at
`

const QueryFindTag = `
	SELECT id, workspace_id, name, created_at, updated_at
	FROM tags
	WHERE id = $1 AND deleted_at IS NULL
`

const QueryUpdateTag = `
	UPDATE tags
	SET name = $2, updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
	RETURNING id, workspace_id, name, created_at, updated_at
`

const QueryDeleteTag = `
	UPDATE tags
	SET deleted_at = NOW(), updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
	RETURNING id
`

const QueryCountGoals = `
	SELECT COUNT(*)
	FROM goals g
	WHERE g.workspace_id = $1
		AND ($2 = '' OR g.title ILIKE '%' || $2 || '%' OR g.description ILIKE '%' || $2 || '%')
		AND ($3 = '' OR g.status = $3)
		AND ($4 = FALSE OR g.is_pinned = $5)
`

const QueryListGoals = `
	SELECT g.id, g.workspace_id, g.created_by, u.full_name AS created_by_name, g.title, g.description, g.target_date, g.status, g.is_pinned,
		0::int AS planned_minutes, 0::int AS actual_minutes, 0::int AS completed_blocks, 0::float AS progress_percent,
		g.created_at, g.updated_at
	FROM goals g
	JOIN users u ON u.id = g.created_by
	WHERE g.workspace_id = $1
		AND ($2 = '' OR g.title ILIKE '%' || $2 || '%' OR g.description ILIKE '%' || $2 || '%')
		AND ($3 = '' OR g.status = $3)
		AND ($4 = FALSE OR g.is_pinned = $5)
	ORDER BY g.is_pinned DESC, g.created_at DESC
	LIMIT $6 OFFSET $7
`

const QueryCreateGoal = `
	INSERT INTO goals (workspace_id, created_by, title, description, target_date, is_pinned)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, workspace_id, created_by, ''::text AS created_by_name, title, description, target_date, status, is_pinned,
		0::int AS planned_minutes, 0::int AS actual_minutes, 0::int AS completed_blocks, 0::float AS progress_percent,
		created_at, updated_at
`

const QueryFindGoal = `
	SELECT g.id, g.workspace_id, g.created_by, u.full_name AS created_by_name, g.title, g.description, g.target_date, g.status, g.is_pinned,
		0::int AS planned_minutes, 0::int AS actual_minutes, 0::int AS completed_blocks, 0::float AS progress_percent,
		g.created_at, g.updated_at
	FROM goals g
	JOIN users u ON u.id = g.created_by
	WHERE g.id = $1
`

const QueryUpdateGoal = `
	UPDATE goals
	SET title = $2, description = $3, target_date = $4, status = $5, is_pinned = $6, updated_at = NOW()
	WHERE id = $1
	RETURNING id, workspace_id, created_by, ''::text AS created_by_name, title, description, target_date, status, is_pinned,
		0::int AS planned_minutes, 0::int AS actual_minutes, 0::int AS completed_blocks, 0::float AS progress_percent,
		created_at, updated_at
`

const QueryArchiveGoal = `
	UPDATE goals
	SET status = 'archived', updated_at = NOW()
	WHERE id = $1
	RETURNING id
`

const QueryCountTasks = `
	SELECT COUNT(DISTINCT t.id)
	FROM tasks t
	LEFT JOIN task_tags tt ON tt.task_id = t.id
	WHERE t.workspace_id = $1 AND t.deleted_at IS NULL
		AND ($2 = '' OR t.title ILIKE '%' || $2 || '%' OR t.description ILIKE '%' || $2 || '%')
		AND ($3 = '' OR t.status = $3)
		AND ($4 = '' OR t.priority = $4)
		AND ($5 = '' OR t.goal_id::text = $5)
		AND ($6 = '' OR t.assignee_id::text = $6)
		AND ($7 = '' OR t.category_id::text = $7)
		AND ($8 = TRUE OR t.status <> 'done')
		AND ($9 = '' OR tt.tag_id::text = ANY(string_to_array($9, ',')))
`

const QueryListTasks = `
	SELECT DISTINCT t.id, t.workspace_id, t.goal_id::text AS goal_id, t.assignee_id::text AS assignee_id, t.category_id::text AS category_id, t.created_by, t.title, t.description,
		t.status, t.priority, t.estimated_minutes, t.position, u.full_name AS assignee_name, u.avatar_url AS assignee_avatar,
		g.title AS goal_title, 0::int AS timeboxes_count, t.created_at, t.updated_at
	FROM tasks t
	LEFT JOIN users u ON u.id = t.assignee_id
	LEFT JOIN goals g ON g.id = t.goal_id
	LEFT JOIN task_tags tt ON tt.task_id = t.id
	WHERE t.workspace_id = $1 AND t.deleted_at IS NULL
		AND ($2 = '' OR t.title ILIKE '%' || $2 || '%' OR t.description ILIKE '%' || $2 || '%')
		AND ($3 = '' OR t.status = $3)
		AND ($4 = '' OR t.priority = $4)
		AND ($5 = '' OR t.goal_id::text = $5)
		AND ($6 = '' OR t.assignee_id::text = $6)
		AND ($7 = '' OR t.category_id::text = $7)
		AND ($8 = TRUE OR t.status <> 'done')
		AND ($9 = '' OR tt.tag_id::text = ANY(string_to_array($9, ',')))
	ORDER BY t.position ASC, t.created_at DESC
	LIMIT $10 OFFSET $11
`

const QueryCreateTask = `
	INSERT INTO tasks (workspace_id, goal_id, assignee_id, category_id, created_by, title, description, status, priority, estimated_minutes, position)
	SELECT $1, $2, $3, $4, $5, $6, $7, 'backlog', $8, $9,
		COALESCE((SELECT MAX(position) + 1000 FROM tasks WHERE workspace_id = $1 AND deleted_at IS NULL), 1000)
	WHERE ($2::uuid IS NULL OR EXISTS (SELECT 1 FROM goals WHERE id = $2 AND workspace_id = $1))
		AND ($3::uuid IS NULL OR EXISTS (SELECT 1 FROM workspace_members WHERE workspace_id = $1 AND user_id = $3 AND status = 'active'))
		AND ($4::uuid IS NULL OR EXISTS (SELECT 1 FROM categories WHERE id = $4 AND workspace_id = $1 AND deleted_at IS NULL))
	RETURNING id, workspace_id, goal_id::text AS goal_id, assignee_id::text AS assignee_id, category_id::text AS category_id, created_by, title, description, status, priority, estimated_minutes,
		position, NULL::text AS assignee_name, NULL::text AS assignee_avatar, NULL::text AS goal_title, 0::int AS timeboxes_count, created_at, updated_at
`

const QueryFindTask = `
	SELECT t.id, t.workspace_id, t.goal_id::text AS goal_id, t.assignee_id::text AS assignee_id, t.category_id::text AS category_id, t.created_by, t.title, t.description,
		t.status, t.priority, t.estimated_minutes, t.position, u.full_name AS assignee_name, u.avatar_url AS assignee_avatar,
		g.title AS goal_title, 0::int AS timeboxes_count, t.created_at, t.updated_at
	FROM tasks t
	LEFT JOIN users u ON u.id = t.assignee_id
	LEFT JOIN goals g ON g.id = t.goal_id
	WHERE t.id = $1 AND t.deleted_at IS NULL
`

const QueryUpdateTask = `
	UPDATE tasks
	SET goal_id = $2, assignee_id = $3, category_id = $4, title = $5, description = $6, status = $7, priority = $8, estimated_minutes = $9, updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
		AND ($2::uuid IS NULL OR EXISTS (SELECT 1 FROM goals WHERE id = $2 AND workspace_id = tasks.workspace_id))
		AND ($3::uuid IS NULL OR EXISTS (SELECT 1 FROM workspace_members WHERE workspace_id = tasks.workspace_id AND user_id = $3 AND status = 'active'))
		AND ($4::uuid IS NULL OR EXISTS (SELECT 1 FROM categories WHERE id = $4 AND workspace_id = tasks.workspace_id AND deleted_at IS NULL))
	RETURNING id, workspace_id, goal_id::text AS goal_id, assignee_id::text AS assignee_id, category_id::text AS category_id, created_by, title, description, status, priority, estimated_minutes,
		position, NULL::text AS assignee_name, NULL::text AS assignee_avatar, NULL::text AS goal_title, 0::int AS timeboxes_count, created_at, updated_at
`

const QueryDeleteTask = `
	UPDATE tasks
	SET deleted_at = NOW(), updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
	RETURNING id
`

const QueryMoveTask = `
	WITH current AS (
		SELECT id, workspace_id, status
		FROM tasks
		WHERE id = $1 AND deleted_at IS NULL
	), updated AS (
		UPDATE tasks
		SET status = $2, position = $3, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, workspace_id, status, position, updated_at
	)
	SELECT updated.id, updated.workspace_id, current.status AS from_status, updated.status AS to_status, updated.position, updated.updated_at
	FROM updated
	JOIN current ON current.id = updated.id
`

const QueryDeleteTaskTags = `
	DELETE FROM task_tags
	WHERE task_id = $1
`

const QueryInsertTaskTag = `
	INSERT INTO task_tags (task_id, tag_id, workspace_id)
	SELECT $1, id, workspace_id
	FROM tags
	WHERE id = $2 AND workspace_id = $3 AND deleted_at IS NULL
	ON CONFLICT DO NOTHING
`

const QueryInsertChecklist = `
	INSERT INTO task_checklists (task_id, title, position)
	VALUES ($1, $2, $3)
`

const QueryListTaskTags = `
	SELECT tags.id, tags.workspace_id, tags.name, tags.created_at, tags.updated_at
	FROM task_tags
	JOIN tags ON tags.id = task_tags.tag_id
	WHERE task_tags.task_id = $1 AND tags.deleted_at IS NULL
	ORDER BY tags.name
`

const QueryListTaskChecklist = `
	SELECT id, title, is_done, position
	FROM task_checklists
	WHERE task_id = $1
	ORDER BY position ASC
`
