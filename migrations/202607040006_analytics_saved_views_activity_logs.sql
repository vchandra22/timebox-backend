CREATE TABLE IF NOT EXISTS saved_views (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	workspace_id UUID NOT NULL REFERENCES workspaces(id),
	user_id UUID NOT NULL REFERENCES users(id),
	name VARCHAR(120) NOT NULL,
	resource_type VARCHAR(32) NOT NULL,
	filter_json JSONB NOT NULL DEFAULT '{}'::jsonb,
	shared BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS activity_logs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	workspace_id UUID REFERENCES workspaces(id),
	actor_id UUID REFERENCES users(id),
	action VARCHAR(120) NOT NULL,
	resource_type VARCHAR(32) NOT NULL,
	resource_id UUID,
	old_value JSONB NOT NULL DEFAULT '{}'::jsonb,
	new_value JSONB NOT NULL DEFAULT '{}'::jsonb,
	ip_address TEXT,
	user_agent TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_saved_views_workspace_resource ON saved_views(workspace_id, resource_type, deleted_at);
CREATE INDEX IF NOT EXISTS idx_activity_logs_workspace_created ON activity_logs(workspace_id, created_at DESC);
