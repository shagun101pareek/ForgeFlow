package generation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shagun101pareek/forgeflow/internal/auth"
	"github.com/shagun101pareek/forgeflow/internal/respond"
	"github.com/shagun101pareek/forgeflow/sql/generated"
)

type Handler struct {
	queries   *db.Queries
	generator GenerationProvider
}

func NewHandler(queries *db.Queries, generator GenerationProvider) *Handler {
	return &Handler{queries: queries, generator: generator}
}

type generateRequest struct {
	ProjectID string `json:"projectId"`
	Prompt    string `json:"prompt"`
}

func (h *Handler) Create(c *fiber.Ctx) error {
	if h.queries == nil {
		return respond.Error(c, fiber.StatusServiceUnavailable, "database unavailable")
	}
	userID, ok := auth.UserIDFrom(c)
	if !ok {
		return respond.Error(c, fiber.StatusUnauthorized, "missing or invalid token")
	}

	var body generateRequest
	if err := c.BodyParser(&body); err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid request body")
	}
	prompt := strings.TrimSpace(body.Prompt)
	if prompt == "" || utf8.RuneCountInString(prompt) > 8000 {
		return respond.Error(c, fiber.StatusBadRequest, "prompt must be between 1 and 8000 characters")
	}
	projectID, err := uuid.Parse(body.ProjectID)
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid project id")
	}

	project, err := h.queries.GetProjectByIDForUser(c.Context(), db.GetProjectByIDForUserParams{
		ID:     projectID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return respond.Error(c, fiber.StatusNotFound, "project not found")
		}
		return respond.Error(c, fiber.StatusInternalServerError, "could not load project")
	}

	row, err := h.queries.CreateGeneration(c.Context(), db.CreateGenerationParams{
		ProjectID: project.ID,
		Prompt:    prompt,
	})
	if err != nil {
		return respond.Error(c, fiber.StatusInternalServerError, "could not save generation")
	}

	go h.runGeneration(row.ID, project.ID, userID, project.Name, prompt)

	return c.Status(fiber.StatusAccepted).JSON(generationFromRow(row))
}

func (h *Handler) List(c *fiber.Ctx) error {
	queries, userID, projectID, ok := h.ownedProject(c)
	if !ok {
		return nil
	}
	_ = queries.FailStaleGenerations(c.Context(), projectID)

	rows, err := queries.ListGenerationsForUserProject(c.Context(), db.ListGenerationsForUserProjectParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		return respond.Error(c, fiber.StatusInternalServerError, "could not list generations")
	}

	items := make([]generationSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, generationSummary{
			ID:        row.ID.String(),
			Prompt:    row.Prompt,
			Status:    row.Status,
			CreatedAt: row.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return c.JSON(items)
}

func (h *Handler) Latest(c *fiber.Ctx) error {
	queries, userID, projectID, ok := h.ownedProject(c)
	if !ok {
		return nil
	}
	_ = queries.FailStaleGenerations(c.Context(), projectID)

	row, err := queries.GetLatestGenerationForUserProject(c.Context(), db.GetLatestGenerationForUserProjectParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return respond.Error(c, fiber.StatusNotFound, "generation not found")
		}
		return respond.Error(c, fiber.StatusInternalServerError, "could not load generation")
	}
	return c.JSON(generationFromRow(row))
}

func (h *Handler) Get(c *fiber.Ctx) error {
	queries, userID, projectID, ok := h.ownedProject(c)
	if !ok {
		return nil
	}
	_ = queries.FailStaleGenerations(c.Context(), projectID)
	generationID, err := uuid.Parse(c.Params("generationId"))
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid generation id")
	}

	row, err := queries.GetGenerationForUser(c.Context(), db.GetGenerationForUserParams{
		ID:        generationID,
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return respond.Error(c, fiber.StatusNotFound, "generation not found")
		}
		return respond.Error(c, fiber.StatusInternalServerError, "could not load generation")
	}
	return c.JSON(generationFromRow(row))
}

type saveRequest struct {
	GenerationID string `json:"generationId"`
	Files        []File `json:"files"`
}

func (h *Handler) Save(c *fiber.Ctx) error {
	queries, userID, projectID, ok := h.ownedProject(c)
	if !ok {
		return nil
	}
	var body saveRequest
	if err := c.BodyParser(&body); err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid request body")
	}
	generationID, err := uuid.Parse(body.GenerationID)
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid generation id")
	}
	row, err := queries.GetGenerationForUser(c.Context(), db.GetGenerationForUserParams{
		ID:        generationID,
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return respond.Error(c, fiber.StatusNotFound, "generation not found")
		}
		return respond.Error(c, fiber.StatusInternalServerError, "could not load generation")
	}
	if row.Status != "completed" {
		return respond.Error(c, fiber.StatusConflict, "generation is not ready to edit")
	}
	current := generationFromRow(row)
	files, err := applyEdits(current.Files, body.Files)
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, err.Error())
	}
	filesJSON, err := json.Marshal(files)
	if err != nil {
		return respond.Error(c, fiber.StatusInternalServerError, "could not save edit")
	}
	saved, err := queries.SaveGeneration(c.Context(), db.SaveGenerationParams{
		ProjectID:     projectID,
		Prompt:        editPrompt(row.Prompt),
		Specification: row.Specification,
		Files:         filesJSON,
	})
	if err != nil {
		return respond.Error(c, fiber.StatusInternalServerError, "could not save edit")
	}
	return c.Status(fiber.StatusCreated).JSON(generationFromRow(saved))
}

func (h *Handler) Export(c *fiber.Ctx) error {
	queries, userID, projectID, ok := h.ownedProject(c)
	if !ok {
		return nil
	}
	generationID, err := uuid.Parse(c.Params("generationId"))
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid generation id")
	}

	row, err := queries.GetGenerationForUser(c.Context(), db.GetGenerationForUserParams{
		ID:        generationID,
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return respond.Error(c, fiber.StatusNotFound, "generation not found")
		}
		return respond.Error(c, fiber.StatusInternalServerError, "could not load generation")
	}
	if row.Status != "completed" {
		return respond.Error(c, fiber.StatusConflict, "generation is not ready to export")
	}

	record := generationFromRow(row)
	body, filename, err := BuildZip(record.Specification.Project.Name, record.Files)
	if err != nil {
		return respond.Error(c, fiber.StatusConflict, "generation is not ready to export")
	}

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	return c.Send(body)
}

type publishRequest struct {
	Token string `json:"token"`
}

func (h *Handler) Publish(c *fiber.Ctx) error {
	queries, userID, projectID, ok := h.ownedProject(c)
	if !ok {
		return nil
	}
	generationID, err := uuid.Parse(c.Params("generationId"))
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid generation id")
	}
	var body publishRequest
	if err := c.BodyParser(&body); err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid request body")
	}
	token := strings.TrimSpace(body.Token)
	if !validGitHubToken(token) {
		return respond.Error(c, fiber.StatusBadRequest, "a GitHub personal access token is required")
	}
	row, err := queries.GetGenerationForUser(c.Context(), db.GetGenerationForUserParams{
		ID:        generationID,
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return respond.Error(c, fiber.StatusNotFound, "generation not found")
		}
		return respond.Error(c, fiber.StatusInternalServerError, "could not load generation")
	}
	if row.Status != "completed" {
		return respond.Error(c, fiber.StatusConflict, "generation is not ready to publish")
	}
	record := generationFromRow(row)
	slug, files, err := projectBundle(record.Specification.Project.Name, record.Files)
	if err != nil {
		return respond.Error(c, fiber.StatusConflict, "generation is not ready to publish")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 45*time.Second)
	defer cancel()
	client := NewGitHubClient()
	name := "forgeflow-" + slug
	description := "Generated with ForgeFlow"
	url, err := client.Publish(ctx, token, name, description, files)
	if errors.Is(err, ErrNameTaken) {
		suffix := strings.ReplaceAll(generationID.String(), "-", "")
		if len(suffix) > 8 {
			suffix = suffix[:8]
		}
		url, err = client.Publish(ctx, token, name+"-"+suffix, description, files)
	}
	if err != nil {
		status, message := githubFailure(err)
		return respond.Error(c, status, message)
	}
	if !validRepositoryURL(url) {
		return respond.Error(c, fiber.StatusBadGateway, "GitHub did not return a repository url")
	}
	if err := queries.SetGenerationGitHubURL(c.Context(), db.SetGenerationGitHubURLParams{
		ID:        generationID,
		GithubUrl: url,
	}); err != nil {
		return respond.Error(c, fiber.StatusInternalServerError, "could not save the repository link")
	}
	return c.JSON(fiber.Map{"url": url})
}

func validGitHubToken(token string) bool {
	if len(token) < 20 || len(token) > 300 || strings.ContainsAny(token, " \t\r\n") {
		return false
	}
	return true
}

func (h *Handler) ownedProject(c *fiber.Ctx) (*db.Queries, uuid.UUID, uuid.UUID, bool) {
	if h.queries == nil {
		_ = respond.Error(c, fiber.StatusServiceUnavailable, "database unavailable")
		return nil, uuid.Nil, uuid.Nil, false
	}
	userID, ok := auth.UserIDFrom(c)
	if !ok {
		_ = respond.Error(c, fiber.StatusUnauthorized, "missing or invalid token")
		return nil, uuid.Nil, uuid.Nil, false
	}
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		_ = respond.Error(c, fiber.StatusBadRequest, "invalid project id")
		return nil, uuid.Nil, uuid.Nil, false
	}
	if _, err := h.queries.GetProjectByIDForUser(c.Context(), db.GetProjectByIDForUserParams{
		ID:     projectID,
		UserID: userID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_ = respond.Error(c, fiber.StatusNotFound, "project not found")
			return nil, uuid.Nil, uuid.Nil, false
		}
		_ = respond.Error(c, fiber.StatusInternalServerError, "could not load project")
		return nil, uuid.Nil, uuid.Nil, false
	}
	return h.queries, userID, projectID, true
}

type generationSummary struct {
	ID        string `json:"id"`
	Prompt    string `json:"prompt"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type generationResponse struct {
	ID            string `json:"id"`
	ProjectID     string `json:"projectId"`
	Prompt        string `json:"prompt"`
	Status        string `json:"status"`
	Error         string `json:"error"`
	Specification Spec   `json:"specification"`
	Files         []File `json:"files"`
	GitHubURL     string `json:"githubUrl"`
	CreatedAt     string `json:"createdAt"`
}

func generationFromRow(row db.Generation) generationResponse {
	var spec Spec
	_ = json.Unmarshal(row.Specification, &spec)
	if spec.Pages == nil {
		spec.Pages = []Page{}
	}
	var files []File
	_ = json.Unmarshal(row.Files, &files)
	if files == nil {
		files = []File{}
	}
	return generationResponse{
		ID:            row.ID.String(),
		ProjectID:     row.ProjectID.String(),
		Prompt:        row.Prompt,
		Status:        row.Status,
		Error:         row.ErrorMessage,
		Specification: spec,
		Files:         files,
		GitHubURL:     row.GithubUrl,
		CreatedAt:     row.CreatedAt.UTC().Format(time.RFC3339),
	}
}
