package service

import (
	"avito-test-task/internal/domain"
	"avito-test-task/internal/repository"
	"context"
)

type Service interface {
	CreateTeam(ctx context.Context, team domain.Team) error
	GetTeam(ctx context.Context, name string) (domain.Team, error)
	SetUserActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
	GetUserReviews(ctx context.Context, userID string) ([]domain.PullRequest, error)

	CreatePR(ctx context.Context, req domain.PullRequest) (domain.PullRequest, error)
	MergePR(ctx context.Context, prID string) (domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (domain.PullRequest, string, error)
}

type service struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
	prRepo   repository.PullRequestRepository
}

var _ Service = (*service)(nil)

func NewService(
	t repository.TeamRepository,
	u repository.UserRepository,
	p repository.PullRequestRepository,
) *service {
	return &service{
		teamRepo: t,
		userRepo: u,
		prRepo:   p,
	}
}

func (s *service) CreateTeam(ctx context.Context, team domain.Team) error {
	return s.teamRepo.CreateTeamWithMembers(ctx, team)
}

func (s *service) GetTeam(ctx context.Context, name string) (domain.Team, error) {
	return s.teamRepo.GetTeamByName(ctx, name)
}

func (s *service) SetUserActive(ctx context.Context, userID string, isActive bool) (domain.User, error) {
	return s.userRepo.SetIsActive(ctx, userID, isActive)
}

func (s *service) GetUserReviews(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return nil, err
	}

	return s.prRepo.GetByReviewerID(ctx, userID)
}
