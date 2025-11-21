package http

import (
	"avito-test-task/internal/domain"
	"avito-test-task/pkg/api"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

func (c *Controller) respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			log.Printf("Failed to encode response: %v\n", err)
		}
	}
}

func (c *Controller) mapDomainPRToAPI(pr domain.PullRequest) api.PullRequest {
	return api.PullRequest{
		PullRequestId:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorId:          pr.AuthorID,
		Status:            api.PullRequestStatus(pr.Status),
		AssignedReviewers: pr.Reviewers,
		CreatedAt:         &pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

func (c *Controller) respondError(w http.ResponseWriter, err error) {
	var code api.ErrorResponseErrorCode
	var status int

	switch {
	case errors.Is(err, domain.ErrNotFound):
		code, status = api.NOTFOUND, http.StatusNotFound
	case errors.Is(err, domain.ErrTeamExists):
		code, status = api.TEAMEXISTS, http.StatusBadRequest
	case errors.Is(err, domain.ErrPRExists):
		code, status = api.PREXISTS, http.StatusConflict
	case errors.Is(err, domain.ErrPRMerged):
		code, status = api.PRMERGED, http.StatusConflict
	case errors.Is(err, domain.ErrNotAssigned):
		code, status = api.NOTASSIGNED, http.StatusConflict
	case errors.Is(err, domain.ErrNoCandidate):
		code, status = api.NOCANDIDATE, http.StatusConflict
	default:
		code, status = "INTERNAL_ERROR", http.StatusInternalServerError
	}

	resp := api.ErrorResponse{
		Error: struct {
			Code    api.ErrorResponseErrorCode `json:"code"`
			Message string                     `json:"message"`
		}{
			Code:    code,
			Message: err.Error(),
		},
	}
	c.respondJSON(w, status, resp)
}
