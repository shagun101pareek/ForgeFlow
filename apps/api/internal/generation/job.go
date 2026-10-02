package generation

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shagun101pareek/forgeflow/sql/generated"
)

func (h *Handler) runGeneration(id, projectID, userID uuid.UUID, projectName, prompt string) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	if err := h.queries.SetGenerationStatus(ctx, db.SetGenerationStatusParams{
		ID:     id,
		Status: "running",
	}); err != nil {
		h.failGeneration(ctx, id, "could not save generation")
		return
	}

	spec, err := h.generator.GenerateSpec(ctx, projectName, prompt)
	if err != nil {
		h.failGeneration(ctx, id, failureMessage(err))
		return
	}
	spec, err = Normalize(spec, projectName)
	if err != nil {
		h.failGeneration(ctx, id, "could not generate a UI specification")
		return
	}

	if imageURL := h.projectImageURL(ctx, projectID, userID); imageURL != "" {
		spec = applyProjectImage(spec, imageURL)
	}
	files := GenerateFiles(spec)
	spec = stripImages(spec)
	specJSON, err := json.Marshal(spec)
	if err != nil {
		h.failGeneration(ctx, id, "could not save generation")
		return
	}
	filesJSON, err := json.Marshal(files)
	if err != nil {
		h.failGeneration(ctx, id, "could not save generation")
		return
	}

	_ = h.queries.CompleteGeneration(ctx, db.CompleteGenerationParams{
		ID:            id,
		Specification: specJSON,
		Files:         filesJSON,
	})
}

func (h *Handler) projectImageURL(ctx context.Context, projectID, userID uuid.UUID) string {
	row, err := h.queries.GetProjectImageForUser(ctx, db.GetProjectImageForUserParams{
		ID:     projectID,
		UserID: userID,
	})
	if err != nil || len(row.Image) == 0 {
		return ""
	}
	return imageDataURL(row.ImageType, row.Image)
}

func (h *Handler) failGeneration(ctx context.Context, id uuid.UUID, message string) {
	_ = h.queries.FailGeneration(ctx, db.FailGenerationParams{
		ID:           id,
		ErrorMessage: message,
	})
}

func failureMessage(err error) string {
	var modelErr *ModelError
	if errors.As(err, &modelErr) {
		return modelErr.Message
	}
	if strings.Contains(strings.ToLower(err.Error()), "api key") {
		return "OpenAI API key is not configured"
	}
	return "could not generate a UI specification"
}
