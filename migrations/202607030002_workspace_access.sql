CREATE TABLE IF NOT EXISTS workspaces (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(120) NOT NULL,
	slug VARCHAR(140) NOT NULL UNIQUE,
	timezone TEXT NOT NULL,
	logo_url TEXT,
	owner_id UUID NOT NULL REFERENCES users(id),
	status TEXT NOT NULL DEFAULT 'active',
	leaderboard_enabled BOOLEAN NOT NULL DEFAULT FALSE,
	default_planner_interval_minutes INT NOT NULL DEFAULT 30,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workspace_members (
	workspace_id UUID NOT NULL REFERENCES workspaces(id),
	user_id UUID NOT NULL REFERENCES users(id),
	role TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'active',
	joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (workspace_id, user_id)
);

CREATE TABLE IF NOT EXISTS workspace_invitations (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	workspace_id UUID NOT NULL REFERENCES workspaces(id),
	email VARCHAR(255) NOT NULL,
	role TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'invited',
	expires_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS teams (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	workspace_id UUID NOT NULL REFERENCES workspaces(id),
	name VARCHAR(120) NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS team_members (
	team_id UUID NOT NULL REFERENCES teams(id),
	workspace_id UUID NOT NULL REFERENCES workspaces(id),
	user_id UUID NOT NULL REFERENCES users(id),
	status TEXT NOT NULL DEFAULT 'active',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ,
	PRIMARY KEY (team_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_workspace_members_user ON workspace_members(user_id, status);
CREATE INDEX IF NOT EXISTS idx_workspace_members_workspace ON workspace_members(workspace_id, status, role);
CREATE INDEX IF NOT EXISTS idx_workspace_invitations_workspace ON workspace_invitations(workspace_id, status);
CREATE INDEX IF NOT EXISTS idx_teams_workspace ON teams(workspace_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_team_members_workspace_user ON team_members(workspace_id, user_id, deleted_at);
