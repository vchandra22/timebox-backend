package database

const QueryResourceWorkspace = `
	SELECT workspace_id::text FROM goals WHERE $1 = 'goal' AND id = $2
	UNION ALL
	SELECT workspace_id::text FROM tasks WHERE $1 = 'task' AND id = $2 AND deleted_at IS NULL
	UNION ALL
	SELECT workspace_id::text FROM timeboxes WHERE $1 = 'timebox' AND id = $2 AND deleted_at IS NULL
	LIMIT 1
`

const QueryListComments = `
	SELECT c.id, c.workspace_id, c.resource_type, c.resource_id::text, c.parent_id::text, c.body, c.author_id::text, u.full_name AS author_name,
		c.edited_at, c.created_at, c.updated_at
	FROM comments c
	JOIN users u ON u.id = c.author_id
	WHERE c.resource_type = $1 AND c.resource_id = $2 AND c.deleted_at IS NULL
		AND (($3::uuid IS NULL AND c.parent_id IS NULL) OR c.parent_id = $3)
	ORDER BY c.created_at ASC
`

const QueryCreateComment = `
	INSERT INTO comments (workspace_id, resource_type, resource_id, parent_id, body, author_id)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, workspace_id, resource_type, resource_id::text, parent_id::text, body, author_id::text, ''::text AS author_name, edited_at, created_at, updated_at
`

const QueryFindComment = `
	SELECT c.id, c.workspace_id, c.resource_type, c.resource_id::text, c.parent_id::text, c.body, c.author_id::text, u.full_name AS author_name,
		c.edited_at, c.created_at, c.updated_at
	FROM comments c
	JOIN users u ON u.id = c.author_id
	WHERE c.id = $1 AND c.deleted_at IS NULL
`

const QueryUpdateComment = `
	UPDATE comments
	SET body = $2, edited_at = NOW(), updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
	RETURNING id, workspace_id, resource_type, resource_id::text, parent_id::text, body, author_id::text, ''::text AS author_name, edited_at, created_at, updated_at
`

const QueryDeleteComment = `
	UPDATE comments
	SET deleted_at = NOW(), updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
	RETURNING id
`

const QueryInsertMention = `
	INSERT INTO comment_mentions (comment_id, user_id)
	SELECT $1, user_id
	FROM workspace_members
	WHERE workspace_id = $3 AND user_id = $2 AND status = 'active'
	ON CONFLICT DO NOTHING
`

const QueryListMentions = `
	SELECT u.id::text AS user_id, u.full_name
	FROM comment_mentions cm
	JOIN users u ON u.id = cm.user_id
	WHERE cm.comment_id = $1
	ORDER BY u.full_name
`

const QueryMoveAttachmentsToComment = `
	UPDATE attachments
	SET resource_type = 'comment', resource_id = $2
	WHERE id = $1 AND workspace_id = $3 AND uploaded_by = $4 AND deleted_at IS NULL
`

const QueryCreateMentionNotification = `
	INSERT INTO notifications (user_id, workspace_id, type, title, body, payload)
	SELECT cm.user_id, c.workspace_id, 'mention', 'You were mentioned', c.body,
		jsonb_build_object('comment_id', c.id, 'resource_type', c.resource_type, 'resource_id', c.resource_id)
	FROM comment_mentions cm
	JOIN comments c ON c.id = cm.comment_id
	LEFT JOIN notification_preferences np ON np.user_id = cm.user_id AND np.trigger_type = 'mention'
	WHERE cm.comment_id = $1 AND COALESCE(np.in_app_enabled, TRUE) = TRUE
`

const QueryCreateAttachment = `
	INSERT INTO attachments (workspace_id, resource_type, resource_id, cloudinary_public_id, url, file_name, file_type, file_size, uploaded_by)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id, workspace_id, resource_type, resource_id::text, cloudinary_public_id, url, file_name, file_type, file_size, uploaded_by::text, ''::text AS uploaded_by_name, created_at
`

const QueryListAttachments = `
	SELECT a.id, a.workspace_id, a.resource_type, a.resource_id::text, a.cloudinary_public_id, a.url, a.file_name, a.file_type, a.file_size,
		a.uploaded_by::text, u.full_name AS uploaded_by_name, a.created_at
	FROM attachments a
	JOIN users u ON u.id = a.uploaded_by
	WHERE a.resource_type = $1 AND a.resource_id = $2 AND a.deleted_at IS NULL
	ORDER BY a.created_at DESC
`

const QueryFindAttachment = `
	SELECT a.id, a.workspace_id, a.resource_type, a.resource_id::text, a.cloudinary_public_id, a.url, a.file_name, a.file_type, a.file_size,
		a.uploaded_by::text, u.full_name AS uploaded_by_name, a.created_at
	FROM attachments a
	JOIN users u ON u.id = a.uploaded_by
	WHERE a.id = $1 AND a.deleted_at IS NULL
`

const QueryDeleteAttachment = `
	UPDATE attachments
	SET deleted_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
	RETURNING id
`

const QueryCountNotifications = `
	SELECT COUNT(*)
	FROM notifications
	WHERE user_id = $1
		AND ($2 = 'all' OR ($2 = 'unread' AND read_at IS NULL) OR ($2 = 'read' AND read_at IS NOT NULL))
		AND ($3 = '' OR type = $3)
`

const QueryListNotifications = `
	SELECT id, user_id::text, workspace_id::text, type, title, body, payload, read_at, created_at
	FROM notifications
	WHERE user_id = $1
		AND ($2 = 'all' OR ($2 = 'unread' AND read_at IS NULL) OR ($2 = 'read' AND read_at IS NOT NULL))
		AND ($3 = '' OR type = $3)
	ORDER BY created_at DESC
	LIMIT $4 OFFSET $5
`

const QueryMarkNotificationRead = `
	UPDATE notifications
	SET read_at = CASE WHEN $3 THEN NOW() ELSE NULL END
	WHERE id = $2 AND user_id = $1
	RETURNING id, user_id::text, workspace_id::text, type, title, body, payload, read_at, created_at
`

const QueryMarkAllNotificationsRead = `
	UPDATE notifications
	SET read_at = NOW()
	WHERE user_id = $1 AND read_at IS NULL AND ($2::text IS NULL OR type = $2)
`

const QueryGetPreferences = `
	SELECT trigger_type, in_app_enabled, email_enabled, telegram_enabled, reminder_minutes_before, send_time
	FROM notification_preferences
	WHERE user_id = $1
	ORDER BY trigger_type
`

const QueryUpsertPreference = `
	INSERT INTO notification_preferences (user_id, trigger_type, in_app_enabled, email_enabled, telegram_enabled, reminder_minutes_before, send_time)
	VALUES ($1, $2, $3, $4, FALSE, $5, $6)
	ON CONFLICT (user_id, trigger_type)
	DO UPDATE SET in_app_enabled = EXCLUDED.in_app_enabled, email_enabled = EXCLUDED.email_enabled, telegram_enabled = FALSE,
		reminder_minutes_before = EXCLUDED.reminder_minutes_before, send_time = EXCLUDED.send_time, updated_at = NOW()
`
