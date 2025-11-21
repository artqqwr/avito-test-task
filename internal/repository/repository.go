package repository

import (
	"avito-test-task/internal/domain"
	"context"
)

type TeamRepository interface {
	CreateTeamWithMembers(ctx context.Context, team domain.Team) error
	GetTeamByName(ctx context.Context, name string) (domain.Team, error)
}

type UserRepository interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
	GetByID(ctx context.Context, userID string) (domain.User, error)
	GetActiveUsersByTeam(ctx context.Context, teamName string) ([]domain.User, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr domain.PullRequest) error
	GetByID(ctx context.Context, id string) (domain.PullRequest, error)

	Merge(ctx context.Context, id string) (domain.PullRequest, error)

	UpdateReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error

	GetByReviewerID(ctx context.Context, reviewerID string) ([]domain.PullRequest, error)
}
