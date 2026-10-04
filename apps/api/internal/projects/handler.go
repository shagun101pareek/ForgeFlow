package projects

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shagun101pareek/forgeflow/internal/auth"
	"github.com/shagun101pareek/forgeflow/internal/respond"
	"github.com/shagun101pareek/forgeflow/sql/generated"
)

type Handler struct {
	queries *db.Queries
}

func NewHandler(queries *db.Queries) *Handler {
	return &Handler{queries: queries}
}

type projectResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	HasImage     bool   `json:"hasImage"`
	LatestPrompt string `json:"latestPrompt,omitempty"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type createProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (h *Handler) List(c *fiber.Ctx) error {
	queries, userID, ok := h.authorized(c)
	if !ok {
		return nil
	}

	rows, err := queries.ListProjectsByUser(c.Context(), userID)
	if err != nil {
		return respond.Error(c, fiber.StatusInternalServerError, "could not list projects")
	}

	projects := make([]projectResponse, 0, len(rows))
	for _, row := range rows {
		projects = append(projects, projectFromList(row))
	}
	return c.JSON(projects)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	queries, userID, ok := h.authorized(c)
	if !ok {
		return nil
	}

	var body createProjectRequest
	if err := c.BodyParser(&body); err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	name, err := validateProjectName(body.Name)
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, err.Error())
	}
	description, err := validateDescription(body.Description)
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, err.Error())
	}

	project, err := queries.CreateProject(c.Context(), db.CreateProjectParams{
		UserID:      userID,
		Name:        name,
		Description: description,
	})
	if err != nil {
		return respond.Error(c, fiber.StatusInternalServerError, "could not create project")
	}

	return c.Status(fiber.StatusCreated).JSON(projectJSON(project.ID, project.Name, project.Description, project.HasImage, project.CreatedAt, project.UpdatedAt))
}

func (h *Handler) Get(c *fiber.Ctx) error {
	queries, userID, ok := h.authorized(c)
	if !ok {
		return nil
	}

	projectID, err := parseProjectID(c.Params("id"))
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid project id")
	}

	project, err := queries.GetProjectByIDForUser(c.Context(), db.GetProjectByIDForUserParams{
		ID:     projectID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return respond.Error(c, fiber.StatusNotFound, "project not found")
		}
		return respond.Error(c, fiber.StatusInternalServerError, "could not load project")
	}

	return c.JSON(projectJSON(project.ID, project.Name, project.Description, project.HasImage, project.CreatedAt, project.UpdatedAt))
}

func (h *Handler) Update(c *fiber.Ctx) error {
	queries, userID, ok := h.authorized(c)
	if !ok {
		return nil
	}

	projectID, err := parseProjectID(c.Params("id"))
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid project id")
	}

	var body updateProjectRequest
	if err := c.BodyParser(&body); err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid request body")
	}
	if body.Name == nil && body.Description == nil {
		return respond.Error(c, fiber.StatusBadRequest, "name or description is required")
	}

	params := db.UpdateProjectParams{
		ID:     projectID,
		UserID: userID,
	}
	if body.Name != nil {
		name, err := validateProjectName(*body.Name)
		if err != nil {
			return respond.Error(c, fiber.StatusBadRequest, err.Error())
		}
		params.Name = pgtype.Text{String: name, Valid: true}
	}
	if body.Description != nil {
		description, err := validateDescription(*body.Description)
		if err != nil {
			return respond.Error(c, fiber.StatusBadRequest, err.Error())
		}
		params.Description = pgtype.Text{String: description, Valid: true}
	}

	project, err := queries.UpdateProject(c.Context(), params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return respond.Error(c, fiber.StatusNotFound, "project not found")
		}
		return respond.Error(c, fiber.StatusInternalServerError, "could not update project")
	}

	return c.JSON(projectJSON(project.ID, project.Name, project.Description, project.HasImage, project.CreatedAt, project.UpdatedAt))
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	queries, userID, ok := h.authorized(c)
	if !ok {
		return nil
	}

	projectID, err := parseProjectID(c.Params("id"))
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid project id")
	}

	deleted, err := queries.DeleteProject(c.Context(), db.DeleteProjectParams{
		ID:     projectID,
		UserID: userID,
	})
	if err != nil {
		return respond.Error(c, fiber.StatusInternalServerError, "could not delete project")
	}
	if deleted == 0 {
		return respond.Error(c, fiber.StatusNotFound, "project not found")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) authorized(c *fiber.Ctx) (*db.Queries, uuid.UUID, bool) {
	if h.queries == nil {
		_ = respond.Error(c, fiber.StatusServiceUnavailable, "database unavailable")
		return nil, uuid.Nil, false
	}
	userID, ok := auth.UserIDFrom(c)
	if !ok {
		_ = respond.Error(c, fiber.StatusUnauthorized, "missing or invalid token")
		return nil, uuid.Nil, false
	}
	return h.queries, userID, true
}

func projectFromList(row db.ListProjectsByUserRow) projectResponse {
	project := projectJSON(row.ID, row.Name, row.Description, row.HasImage, row.CreatedAt, row.UpdatedAt)
	project.LatestPrompt = row.LatestPrompt
	return project
}

func projectJSON(id uuid.UUID, name, description string, hasImage bool, createdAt, updatedAt time.Time) projectResponse {
	return projectResponse{
		ID:          id.String(),
		Name:        name,
		Description: description,
		HasImage:    hasImage,
		CreatedAt:   createdAt.UTC().Format(time.RFC3339),
		UpdatedAt:   updatedAt.UTC().Format(time.RFC3339),
	}
}

func parseProjectID(raw string) (uuid.UUID, error) {
	return uuid.Parse(raw)
}

func validateProjectName(name string) (string, error) {
	name = strings.TrimSpace(name)
	count := utf8.RuneCountInString(name)
	if count < 1 || count > 120 {
		return "", errors.New("name must be between 1 and 120 characters")
	}
	return name, nil
}

func validateDescription(description string) (string, error) {
	description = strings.TrimSpace(description)
	if utf8.RuneCountInString(description) > 2000 {
		return "", errors.New("description must be at most 2000 characters")
	}
	return description, nil
}
