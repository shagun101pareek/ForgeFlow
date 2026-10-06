package projects

import (
	"errors"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/shagun101pareek/forgeflow/internal/respond"
	"github.com/shagun101pareek/forgeflow/sql/generated"
)

func (h *Handler) Duplicate(c *fiber.Ctx) error {
	queries, userID, ok := h.authorized(c)
	if !ok {
		return nil
	}
	projectID, err := parseProjectID(c.Params("id"))
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid project id")
	}
	source, err := queries.GetProjectByIDForUser(c.Context(), db.GetProjectByIDForUserParams{
		ID:     projectID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return respond.Error(c, fiber.StatusNotFound, "project not found")
		}
		return respond.Error(c, fiber.StatusInternalServerError, "could not load project")
	}

	created, err := queries.CreateProject(c.Context(), db.CreateProjectParams{
		UserID:      userID,
		Name:        duplicateName(source.Name),
		Description: source.Description,
	})
	if err != nil {
		return respond.Error(c, fiber.StatusInternalServerError, "could not duplicate project")
	}

	hasImage := false
	image, err := queries.GetProjectImageForUser(c.Context(), db.GetProjectImageForUserParams{
		ID:     projectID,
		UserID: userID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		_, _ = queries.DeleteProject(c.Context(), db.DeleteProjectParams{ID: created.ID, UserID: userID})
		return respond.Error(c, fiber.StatusInternalServerError, "could not duplicate project")
	}
	if len(image.Image) > 0 && image.ImageType != "" {
		if err := queries.SetProjectImage(c.Context(), db.SetProjectImageParams{
			ID:        created.ID,
			UserID:    userID,
			Image:     image.Image,
			ImageType: image.ImageType,
		}); err != nil {
			_, _ = queries.DeleteProject(c.Context(), db.DeleteProjectParams{ID: created.ID, UserID: userID})
			return respond.Error(c, fiber.StatusInternalServerError, "could not duplicate project")
		}
		hasImage = true
	}

	generation, err := queries.GetLatestCompletedGenerationForUserProject(c.Context(), db.GetLatestCompletedGenerationForUserProjectParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		_, _ = queries.DeleteProject(c.Context(), db.DeleteProjectParams{ID: created.ID, UserID: userID})
		return respond.Error(c, fiber.StatusInternalServerError, "could not duplicate project")
	}
	if err == nil {
		if _, err := queries.SaveGeneration(c.Context(), db.SaveGenerationParams{
			ProjectID:     created.ID,
			Prompt:        generation.Prompt,
			Specification: generation.Specification,
			Files:         generation.Files,
		}); err != nil {
			_, _ = queries.DeleteProject(c.Context(), db.DeleteProjectParams{ID: created.ID, UserID: userID})
			return respond.Error(c, fiber.StatusInternalServerError, "could not duplicate project")
		}
	}

	return c.Status(fiber.StatusCreated).JSON(projectJSON(created.ID, created.Name, created.Description, hasImage, created.CreatedAt, created.UpdatedAt))
}

func duplicateName(name string) string {
	const suffix = " copy"
	const maxRunes = 120
	if utf8.RuneCountInString(name)+len(suffix) <= maxRunes {
		return name + suffix
	}
	runes := []rune(name)
	keep := maxRunes - len(suffix)
	return string(runes[:keep]) + suffix
}
