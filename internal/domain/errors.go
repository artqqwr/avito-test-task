package domain

import "errors"

var (
	ErrNotFound    = errors.New("resource not found")
	ErrTeamExists  = errors.New("team already exists")
	ErrPRExists    = errors.New("pull request already exists")
	ErrPRMerged    = errors.New("pull request is already merged")
	ErrNotAssigned = errors.New("user is not assigned as a reviewer")
	ErrNoCandidate = errors.New("no active candidates available for review")
)
