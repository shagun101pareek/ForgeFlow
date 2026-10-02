package projects

import (
	"errors"
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shagun101pareek/forgeflow/internal/generation"
	"github.com/shagun101pareek/forgeflow/internal/respond"
	"github.com/shagun101pareek/forgeflow/sql/generated"
)

const maxProjectImageBytes = 1_000_000

func (h *Handler) UploadImage(c *fiber.Ctx) error {
	queries, userID, ok := h.authorized(c)
	if !ok {
		return nil
	}
	projectID, err := parseProjectID(c.Params("id"))
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid project id")
	}
	project, ok := h.ownedProject(c, queries, projectID, userID)
	if !ok {
		return nil
	}
	file, err := c.FormFile("image")
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "an image file is required")
	}
	if file.Size > maxProjectImageBytes {
		return respond.Error(c, fiber.StatusBadRequest, "image must be at most 1 MB")
	}
	opened, err := file.Open()
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "an image file is required")
	}
	defer opened.Close()
	data, err := io.ReadAll(io.LimitReader(opened, maxProjectImageBytes+1))
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "an image file is required")
	}
	if len(data) == 0 || len(data) > maxProjectImageBytes {
		return respond.Error(c, fiber.StatusBadRequest, "image must be at most 1 MB")
	}
	contentType := generation.DetectImageType(data)
	if contentType == "" {
		return respond.Error(c, fiber.StatusBadRequest, "image must be a PNG, JPEG, GIF, or WebP")
	}
	if err := queries.SetProjectImage(c.Context(), db.SetProjectImageParams{
		ID:        projectID,
		UserID:    userID,
		Image:     data,
		ImageType: contentType,
	}); err != nil {
		return respond.Error(c, fiber.StatusInternalServerError, "could not save image")
	}
	project.HasImage = true
	return c.JSON(project)
}

func (h *Handler) GetImage(c *fiber.Ctx) error {
	queries, userID, ok := h.authorized(c)
	if !ok {
		return nil
	}
	projectID, err := parseProjectID(c.Params("id"))
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid project id")
	}
	row, err := queries.GetProjectImageForUser(c.Context(), db.GetProjectImageForUserParams{
		ID:     projectID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return respond.Error(c, fiber.StatusNotFound, "project not found")
		}
		return respond.Error(c, fiber.StatusInternalServerError, "could not load image")
	}
	if len(row.Image) == 0 || row.ImageType == "" {
		return respond.Error(c, fiber.StatusNotFound, "image not found")
	}
	c.Set("Content-Type", row.ImageType)
	c.Set("Cache-Control", "private, max-age=60")
	return c.Send(row.Image)
}

func (h *Handler) ClearImage(c *fiber.Ctx) error {
	queries, userID, ok := h.authorized(c)
	if !ok {
		return nil
	}
	projectID, err := parseProjectID(c.Params("id"))
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid project id")
	}
	project, ok := h.ownedProject(c, queries, projectID, userID)
	if !ok {
		return nil
	}
	if _, err := queries.ClearProjectImage(c.Context(), db.ClearProjectImageParams{
		ID:     projectID,
		UserID: userID,
	}); err != nil {
		return respond.Error(c, fiber.StatusInternalServerError, "could not remove image")
	}
	project.HasImage = false
	return c.JSON(project)
}

func (h *Handler) ownedProject(c *fiber.Ctx, queries *db.Queries, projectID, userID uuid.UUID) (projectResponse, bool) {
	project, err := queries.GetProjectByIDForUser(c.Context(), db.GetProjectByIDForUserParams{
		ID:     projectID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_ = respond.Error(c, fiber.StatusNotFound, "project not found")
			return projectResponse{}, false
		}
		_ = respond.Error(c, fiber.StatusInternalServerError, "could not load project")
		return projectResponse{}, false
	}
	return projectJSON(project.ID, project.Name, project.Description, project.HasImage, project.CreatedAt, project.UpdatedAt), true
}
