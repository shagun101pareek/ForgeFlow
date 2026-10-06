package generation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrNameTaken = errors.New("repository name is already taken")

type gitHubStatusError struct {
	Code int
}

func (e gitHubStatusError) Error() string {
	return fmt.Sprintf("github status %d", e.Code)
}

type RepositoryPublisher interface {
	Publish(ctx context.Context, token, name, description string, files map[string]string) (string, error)
}

type GitHubClient struct {
	API  string
	HTTP *http.Client
}

func NewGitHubClient() GitHubClient {
	return GitHubClient{
		API: "https://api.github.com",
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (c GitHubClient) Publish(ctx context.Context, token, name, description string, files map[string]string) (string, error) {
	var created struct {
		HTMLURL string `json:"html_url"`
		Name    string `json:"name"`
		Owner   struct {
			Login string `json:"login"`
		} `json:"owner"`
	}
	status, message, err := c.request(ctx, token, http.MethodPost, "/user/repos", map[string]any{
		"name":        name,
		"private":     true,
		"description": description,
		"auto_init":   false,
	}, &created)
	if err != nil {
		return "", err
	}
	if status == http.StatusUnprocessableEntity && strings.Contains(strings.ToLower(message), "already exists") {
		return "", ErrNameTaken
	}
	if status < 200 || status >= 300 {
		return "", gitHubStatusError{Code: status}
	}

	tree := make([]map[string]string, 0, len(files))
	for path, content := range files {
		var blob struct {
			SHA string `json:"sha"`
		}
		status, _, err = c.request(ctx, token, http.MethodPost, "/repos/"+created.Owner.Login+"/"+created.Name+"/git/blobs", map[string]string{
			"content":  content,
			"encoding": "utf-8",
		}, &blob)
		if err != nil {
			return "", err
		}
		if status < 200 || status >= 300 || blob.SHA == "" {
			return "", gitHubStatusError{Code: status}
		}
		tree = append(tree, map[string]string{
			"path": path,
			"mode": "100644",
			"type": "blob",
			"sha":  blob.SHA,
		})
	}

	var treeResult struct {
		SHA string `json:"sha"`
	}
	status, _, err = c.request(ctx, token, http.MethodPost, "/repos/"+created.Owner.Login+"/"+created.Name+"/git/trees", map[string]any{
		"tree": tree,
	}, &treeResult)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 || treeResult.SHA == "" {
		return "", gitHubStatusError{Code: status}
	}

	var commit struct {
		SHA string `json:"sha"`
	}
	status, _, err = c.request(ctx, token, http.MethodPost, "/repos/"+created.Owner.Login+"/"+created.Name+"/git/commits", map[string]any{
		"message": "Add the ForgeFlow generation",
		"tree":    treeResult.SHA,
		"parents": []string{},
	}, &commit)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 || commit.SHA == "" {
		return "", gitHubStatusError{Code: status}
	}

	status, _, err = c.request(ctx, token, http.MethodPost, "/repos/"+created.Owner.Login+"/"+created.Name+"/git/refs", map[string]string{
		"ref": "refs/heads/main",
		"sha": commit.SHA,
	}, nil)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		return "", gitHubStatusError{Code: status}
	}
	if created.HTMLURL == "" {
		return "", errors.New("github did not return a repository url")
	}
	return created.HTMLURL, nil
}

func (c GitHubClient) request(ctx context.Context, token, method, path string, body any, out any) (int, string, error) {
	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return 0, "", err
		}
		payload = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.API, "/")+path, payload)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "ForgeFlow")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return res.StatusCode, "", err
	}
	message := ""
	if out != nil && len(raw) > 0 {
		_ = json.Unmarshal(raw, out)
	}
	var apiErr struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(raw, &apiErr) == nil {
		message = apiErr.Message
	}
	return res.StatusCode, message, nil
}

func ValidRepositoryURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host != "github.com" {
		return false
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return false
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

func githubFailure(err error) (int, string) {
	if errors.Is(err, ErrNameTaken) {
		return http.StatusConflict, "a GitHub repository with that name already exists"
	}
	var status gitHubStatusError
	if errors.As(err, &status) {
		switch status.Code {
		case http.StatusUnauthorized:
			return http.StatusBadGateway, "GitHub rejected the token"
		case http.StatusForbidden:
			return http.StatusBadGateway, "GitHub token cannot create repositories"
		}
	}
	return http.StatusBadGateway, "could not publish to GitHub"
}
