CREATE TABLE IF NOT EXISTS categories (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	workspace_id UUID NOT NULL REFERENCES workspaces(id),
	name VARCHAR(120) NOT NULL,
	color VARCHAR(16) NOT NULL,
	is_default BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS tags (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	workspace_id UUID NOT NULL REFERENCES workspaces(id),
	name VARCHAR(80) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ,
	UNIQUE (workspace_id, name)
);

CREATE TABLE IF NOT EXISTS goals (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	workspace_id UUID NOT NULL REFERENCES workspaces(id),
	created_by UUID NOT NULL REFERENCES users(id),
	title VARCHAR(200) NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	target_date DATE,
	status VARCHAR(24) NOT NULL DEFAULT 'active',
	is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tasks (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	workspace_id UUID NOT NULL REFERENCES workspaces(id),
	goal_id UUID REFERENCES goals(id),
	assignee_id UUID REFERENCES users(id),
	category_id UUID REFERENCES categories(id),
	created_by UUID NOT NULL REFERENCES users(id),
	title VARCHAR(200) NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status VARCHAR(24) NOT NULL DEFAULT 'backlog',
	priority VARCHAR(16) NOT NULL DEFAULT 'medium',
	estimated_minutes INT,
	position INT NOT NULL DEFAULT 1000,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS task_tags (
	task_id UUID NOT NULL REFERENCES tasks(id),
	tag_id UUID NOT NULL REFERENCES tags(id),
	workspace_id UUID NOT NULL REFERENCES workspaces(id),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (task_id, tag_id)
);

CREATE TABLE IF NOT EXISTS task_checklists (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	task_id UUID NOT NULL REFERENCES tasks(id),
	title VARCHAR(200) NOT NULL,
	is_done BOOLEAN NOT NULL DEFAULT FALSE,
	position INT NOT NULL DEFAULT 1000,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_categories_workspace ON categories(workspace_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_tags_workspace ON tags(workspace_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_goals_workspace ON goals(workspace_id, status, is_pinned);
CREATE INDEX IF NOT EXISTS idx_tasks_workspace_status ON tasks(workspace_id, status, position);
CREATE INDEX IF NOT EXISTS idx_tasks_workspace_assignee ON tasks(workspace_id, assignee_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_task_tags_workspace ON task_tags(workspace_id, tag_id);
