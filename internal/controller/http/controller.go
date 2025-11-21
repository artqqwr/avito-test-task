package http

import (
	"avito-test-task/internal/domain"
	"avito-test-task/internal/service"
	"avito-test-task/pkg/api"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Controller struct {
	service service.Service
	Handler http.Handler
}

func NewController(s service.Service) *Controller {
	c := &Controller{
		service: s,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	api.HandlerFromMux(c, r)

	c.Handler = r
	return c
}

func (c *Controller) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	var body api.Team
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	members := make([]domain.User, len(body.Members))
	for i, m := range body.Members {
		members[i] = domain.User{
			ID:       m.UserId,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}
	team := domain.Team{
		Name:    body.TeamName,
		Members: members,
	}

	if err := c.service.CreateTeam(r.Context(), team); err != nil {
		c.respondError(w, err)
		return
	}

	response := struct {
		Team api.Team `json:"team"`
	}{Team: body}

	c.respondJSON(w, http.StatusCreated, response)
}

func (c *Controller) GetTeamGet(w http.ResponseWriter, r *http.Request, params api.GetTeamGetParams) {
	teamName := string(params.TeamName)

	team, err := c.service.GetTeam(r.Context(), teamName)
	if err != nil {
		c.respondError(w, err)
		return
	}

	members := make([]api.TeamMember, len(team.Members))
	for i, m := range team.Members {
		members[i] = api.TeamMember{
			UserId:   m.ID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	c.respondJSON(w, http.StatusOK, api.Team{
		TeamName: team.Name,
		Members:  members,
	})
}

func (c *Controller) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	var body api.PostUsersSetIsActiveJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	user, err := c.service.SetUserActive(r.Context(), body.UserId, body.IsActive)
	if err != nil {
		c.respondError(w, err)
		return
	}

	response := struct {
		User api.User `json:"user"`
	}{
		User: api.User{
			UserId:   user.ID,
			Username: user.Username,
			TeamName: user.TeamName,
			IsActive: user.IsActive,
		},
	}
	c.respondJSON(w, http.StatusOK, response)
}

func (c *Controller) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params api.GetUsersGetReviewParams) {
	userID := string(params.UserId)

	prs, err := c.service.GetUserReviews(r.Context(), userID)
	if err != nil {
		c.respondError(w, err)
		return
	}

	prShorts := make([]api.PullRequestShort, len(prs))
	for i, pr := range prs {
		prShorts[i] = api.PullRequestShort{
			PullRequestId:   pr.ID,
			PullRequestName: pr.Name,
			AuthorId:        pr.AuthorID,
			Status:          api.PullRequestShortStatus(pr.Status),
		}
	}

	response := struct {
		UserId       string                 `json:"user_id"`
		PullRequests []api.PullRequestShort `json:"pull_requests"`
	}{
		UserId:       userID,
		PullRequests: prShorts,
	}
	c.respondJSON(w, http.StatusOK, response)
}

func (c *Controller) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestCreateJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	req := domain.PullRequest{
		ID:       body.PullRequestId,
		Name:     body.PullRequestName,
		AuthorID: body.AuthorId,
	}

	pr, err := c.service.CreatePR(r.Context(), req)
	if err != nil {
		c.respondError(w, err)
		return
	}

	response := struct {
		Pr api.PullRequest `json:"pr"`
	}{
		Pr: c.mapDomainPRToAPI(pr),
	}
	c.respondJSON(w, http.StatusCreated, response)
}

func (c *Controller) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestMergeJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	pr, err := c.service.MergePR(r.Context(), body.PullRequestId)
	if err != nil {
		c.respondError(w, err)
		return
	}

	response := struct {
		Pr api.PullRequest `json:"pr"`
	}{
		Pr: c.mapDomainPRToAPI(pr),
	}
	c.respondJSON(w, http.StatusOK, response)
}

func (c *Controller) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestReassignJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	pr, newReviewerID, err := c.service.ReassignReviewer(r.Context(), body.PullRequestId, body.OldUserId)
	if err != nil {
		c.respondError(w, err)
		return
	}

	response := struct {
		Pr         api.PullRequest `json:"pr"`
		ReplacedBy string          `json:"replaced_by"`
	}{
		Pr:         c.mapDomainPRToAPI(pr),
		ReplacedBy: newReviewerID,
	}
	c.respondJSON(w, http.StatusOK, response)
}
