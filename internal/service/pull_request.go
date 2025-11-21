package service

import (
	"avito-test-task/internal/domain"
	"context"
	"math/rand"
	"time"
)

func (s *service) CreatePR(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error) {
	author, err := s.userRepo.GetByID(ctx, pr.AuthorID)
	if err != nil {
		return domain.PullRequest{}, err
	}

	// Get candidates, only active ones from the same team
	candidates, err := s.userRepo.GetActiveUsersByTeam(ctx, author.TeamName)
	if err != nil {
		return domain.PullRequest{}, err
	}

	// Filtering the author
	validCandidates := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if c.ID != pr.AuthorID {
			validCandidates = append(validCandidates, c.ID)
		}
	}

	if len(validCandidates) == 0 {
		return domain.PullRequest{}, domain.ErrNoCandidate
	}

	rand.Shuffle(len(validCandidates), func(i, j int) {
		validCandidates[i], validCandidates[j] = validCandidates[j], validCandidates[i]
	})

	assignCount := 2
	if len(validCandidates) < 2 {
		assignCount = len(validCandidates)
	}

	pr.Reviewers = validCandidates[:assignCount]
	pr.Status = domain.PRStatusOpen
	pr.CreatedAt = time.Now()

	if err := s.prRepo.Create(ctx, pr); err != nil {
		return domain.PullRequest{}, err
	}

	return pr, nil
}

func (s *service) MergePR(ctx context.Context, prID string) (domain.PullRequest, error) {
	return s.prRepo.Merge(ctx, prID)
}

func (s *service) ReassignReviewer(ctx context.Context, prID, oldUserID string) (domain.PullRequest, string, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, "", domain.ErrNotFound
	}

	if pr.Status == domain.PRStatusMerged {
		return domain.PullRequest{}, "", domain.ErrPRMerged
	}

	isAssigned := false
	for _, id := range pr.Reviewers {
		if id == oldUserID {
			isAssigned = true
			break
		}
	}
	if !isAssigned {
		return domain.PullRequest{}, "", domain.ErrNotAssigned
	}

	oldUser, err := s.userRepo.GetByID(ctx, oldUserID)
	if err != nil {
		return domain.PullRequest{}, "", domain.ErrNotFound
	}

	candidates, err := s.userRepo.GetActiveUsersByTeam(ctx, oldUser.TeamName)
	if err != nil {
		return domain.PullRequest{}, "", err
	}

	validCandidates := make([]string, 0)

	currentReviewersMap := make(map[string]bool)
	for _, r := range pr.Reviewers {
		currentReviewersMap[r] = true
	}

	for _, c := range candidates {
		if c.ID == pr.AuthorID {
			continue
		}
		if currentReviewersMap[c.ID] {
			continue
		}
		validCandidates = append(validCandidates, c.ID)
	}

	if len(validCandidates) == 0 {
		return domain.PullRequest{}, "", domain.ErrNoCandidate
	}

	newReviewerID := validCandidates[rand.Intn(len(validCandidates))]

	if err := s.prRepo.UpdateReviewer(ctx, prID, oldUserID, newReviewerID); err != nil {
		return domain.PullRequest{}, "", err
	}

	for i, r := range pr.Reviewers {
		if r == oldUserID {
			pr.Reviewers[i] = newReviewerID
			break
		}
	}

	return pr, newReviewerID, nil
}
