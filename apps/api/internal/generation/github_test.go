package generation

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGitHubPublishCreatesAPrivateRepository(t *testing.T) {
	var sawAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/user/repos":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["private"] != true || body["name"] != "forgeflow-north-star" {
				t.Fatalf("repo body = %#v", body)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"html_url":"https://github.com/ada/forgeflow-north-star","name":"forgeflow-north-star","owner":{"login":"ada"}}`))
		case strings.HasSuffix(r.URL.Path, "/git/blobs"):
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"sha":"blob123"}`))
		case strings.HasSuffix(r.URL.Path, "/git/trees"):
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"sha":"tree123"}`))
		case strings.HasSuffix(r.URL.Path, "/git/commits"):
			raw, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(raw), `"parents":[]`) {
				t.Fatalf("commit = %s", raw)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"sha":"commit123"}`))
		case strings.HasSuffix(r.URL.Path, "/git/refs"):
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["ref"] != "refs/heads/main" || body["sha"] != "commit123" {
				t.Fatalf("ref = %#v", body)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"ref":"refs/heads/main"}`))
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := GitHubClient{API: server.URL, HTTP: server.Client()}
	url, err := client.Publish(context.Background(), "ghp_testtokenvalue123456", "forgeflow-north-star", "Generated with ForgeFlow", map[string]string{
		"src/App.jsx": "export default function App(){return null}",
	})
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if url != "https://github.com/ada/forgeflow-north-star" {
		t.Fatalf("url = %s", url)
	}
	if sawAuth != "Bearer ghp_testtokenvalue123456" {
		t.Fatalf("auth = %s", sawAuth)
	}
}

func TestGitHubPublishReportsATakenName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"Repository creation failed. Name already exists on this account"}`))
	}))
	defer server.Close()

	client := GitHubClient{API: server.URL, HTTP: server.Client()}
	_, err := client.Publish(context.Background(), "ghp_testtokenvalue123456", "forgeflow-notes", "Generated with ForgeFlow", map[string]string{
		"src/App.jsx": "x",
	})
	if err != ErrNameTaken {
		t.Fatalf("err = %v", err)
	}
	code, message := githubFailure(err)
	if code != http.StatusConflict || strings.Contains(message, "ghp_") {
		t.Fatalf("failure = %d %s", code, message)
	}
}
