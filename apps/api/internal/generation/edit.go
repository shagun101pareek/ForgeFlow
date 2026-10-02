package generation

import (
	"errors"
	"strings"
	"unicode/utf8"
)

const maxEditedFileBytes = 200_000

func editPrompt(original string) string {
	const prefix = "Edited: "
	base := strings.TrimPrefix(strings.TrimSpace(original), prefix)
	prompt := prefix + base
	if utf8.RuneCountInString(prompt) <= 8000 {
		return prompt
	}
	return string([]rune(prompt)[:8000])
}

func applyEdits(existing, submitted []File) ([]File, error) {
	if len(existing) == 0 || len(submitted) != len(existing) {
		return nil, errors.New("files do not match this generation")
	}
	byPath := make(map[string]File, len(submitted))
	for _, file := range submitted {
		if _, ok := byPath[file.Path]; ok {
			return nil, errors.New("files do not match this generation")
		}
		byPath[file.Path] = file
	}
	out := make([]File, len(existing))
	for i, file := range existing {
		next, ok := byPath[file.Path]
		if !ok {
			return nil, errors.New("files do not match this generation")
		}
		if len(next.Code) > maxEditedFileBytes {
			return nil, errors.New("file is too large")
		}
		if file.Path == "/App.js" && strings.TrimSpace(next.Code) == "" {
			return nil, errors.New("App.js cannot be empty")
		}
		out[i] = File{Path: file.Path, Code: next.Code}
	}
	return out, nil
}
