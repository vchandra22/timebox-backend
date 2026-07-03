CREATE TABLE IF NOT EXISTS timeboxes (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	workspace_id UUID NOT NULL REFERENCES workspaces(id),
	task_id UUID REFERENCES tasks(id),
	owner_id UUID NOT NULL REFERENCES users(id),
	category_id UUID REFERENCES categories(id),
	title VARCHAR(200) NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	scheduled_start TIMESTAMPTZ NOT NULL,
	scheduled_end TIMESTAMPTZ NOT NULL,
	planned_minutes INT NOT NULL,
	actual_minutes INT NOT NULL DEFAULT 0,
	status VARCHAR(24) NOT NULL DEFAULT 'planned',
	is_buffer BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS time_logs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	timebox_id UUID NOT NULL REFERENCES timeboxes(id),
	workspace_id UUID NOT NULL REFERENCES workspaces(id),
	started_at TIMESTAMPTZ NOT NULL,
	ended_at TIMESTAMPTZ,
	duration_seconds INT NOT NULL DEFAULT 0,
	source VARCHAR(16) NOT NULL,
	note TEXT,
	created_by UUID NOT NULL REFERENCES users(id),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_timeboxes_workspace_schedule ON timeboxes(workspace_id, owner_id, scheduled_start, scheduled_end);
CREATE INDEX IF NOT EXISTS idx_timeboxes_workspace_status ON timeboxes(workspace_id, status, deleted_at);
CREATE INDEX IF NOT EXISTS idx_time_logs_timebox ON time_logs(timebox_id, started_at);
