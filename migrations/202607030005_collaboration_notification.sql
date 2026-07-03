CREATE TABLE IF NOT EXISTS comments (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	workspace_id UUID NOT NULL REFERENCES workspaces(id),
	resource_type VARCHAR(32) NOT NULL,
	resource_id UUID NOT NULL,
	parent_id UUID REFERENCES comments(id),
	body TEXT NOT NULL,
	author_id UUID NOT NULL REFERENCES users(id),
	edited_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS comment_mentions (
	comment_id UUID NOT NULL REFERENCES comments(id),
	user_id UUID NOT NULL REFERENCES users(id),
	PRIMARY KEY (comment_id, user_id)
);

CREATE TABLE IF NOT EXISTS attachments (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	workspace_id UUID NOT NULL REFERENCES workspaces(id),
	resource_type VARCHAR(32) NOT NULL,
	resource_id UUID NOT NULL,
	cloudinary_public_id TEXT NOT NULL,
	url TEXT NOT NULL,
	file_name TEXT NOT NULL,
	file_type TEXT NOT NULL,
	file_size BIGINT NOT NULL,
	uploaded_by UUID NOT NULL REFERENCES users(id),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS notifications (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users(id),
	workspace_id UUID REFERENCES workspaces(id),
	type VARCHAR(64) NOT NULL,
	title TEXT NOT NULL,
	body TEXT NOT NULL,
	payload JSONB NOT NULL DEFAULT '{}'::jsonb,
	read_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notification_preferences (
	user_id UUID NOT NULL REFERENCES users(id),
	trigger_type VARCHAR(64) NOT NULL,
	in_app_enabled BOOLEAN NOT NULL DEFAULT TRUE,
	email_enabled BOOLEAN NOT NULL DEFAULT FALSE,
	telegram_enabled BOOLEAN NOT NULL DEFAULT FALSE,
	reminder_minutes_before INT,
	send_time VARCHAR(8),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (user_id, trigger_type)
);

CREATE INDEX IF NOT EXISTS idx_comments_resource ON comments(resource_type, resource_id, parent_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_attachments_resource ON attachments(resource_type, resource_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id, read_at, created_at);
