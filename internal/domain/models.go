package domain

import (
	"time"
)

type Team struct {
	Name    string
	Members []User
}

type User struct {
	ID       string
	Username string
	TeamName string
	IsActive bool
}

type PullRequestStatus string

const (
	PRStatusOpen   PullRequestStatus = "OPEN"
	PRStatusMerged PullRequestStatus = "MERGED"
)

type PullRequest struct {
	ID        string
	Name      string
	AuthorID  string
	Status    PullRequestStatus
	CreatedAt time.Time
	MergedAt  *time.Time

	Reviewers []string
}
