package entity

import "time"

type Category struct {
	ID          string
	WorkspaceID string
	Name        string
	Color       string
	IsDefault   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Tag struct {
	ID          string
	WorkspaceID string
	Name        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Goal struct {
	ID              string
	WorkspaceID     string
	CreatedBy       string
	CreatedByName   string
	Title           string
	Description     string
	TargetDate      *time.Time
	Status          string
	IsPinned        bool
	IsPinnedSet     bool
	PlannedMinutes  int
	ActualMinutes   int
	CompletedBlocks int
	ProgressPercent float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Task struct {
	ID               string
	WorkspaceID      string
	GoalID           *string
	AssigneeID       *string
	CategoryID       *string
	CreatedBy        string
	Title            string
	Description      string
	Status           string
	Priority         string
	EstimatedMinutes *int
	Position         int
	Assignee         *UserSummary
	Goal             *GoalSummary
	Tags             []Tag
	TagIDs           []string
	TagIDsSet        bool
	Checklist        []TaskChecklist
	TimeboxesCount   int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type UserSummary struct {
	ID        string
	FullName  string
	AvatarURL *string
}

type GoalSummary struct {
	ID    string
	Title string
}

type TaskChecklist struct {
	ID       string
	Title    string
	IsDone   bool
	Position int
}

type TaskMove struct {
	ID          string
	WorkspaceID string
	FromStatus  string
	ToStatus    string
	Position    int
	UpdatedAt   time.Time
}
