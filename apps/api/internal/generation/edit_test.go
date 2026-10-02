package generation

import "testing"

func TestEditPromptPrefixesOnce(t *testing.T) {
	if got := editPrompt("Build a landing page"); got != "Edited: Build a landing page" {
		t.Fatalf("prompt = %q", got)
	}
	if got := editPrompt("Edited: Build a landing page"); got != "Edited: Build a landing page" {
		t.Fatalf("second edit = %q", got)
	}
}

func TestApplyEditsKeepsExistingPaths(t *testing.T) {
	existing := []File{
		{Path: "/App.js", Code: "original"},
		{Path: "/styles.css", Code: "body{}"},
	}
	updated, err := applyEdits(existing, []File{
		{Path: "/styles.css", Code: "body{color:red}"},
		{Path: "/App.js", Code: "export default function App(){return null}"},
	})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if updated[0].Code != "export default function App(){return null}" || updated[1].Path != "/styles.css" {
		t.Fatalf("files = %#v", updated)
	}
	if _, err := applyEdits(existing, []File{{Path: "/secret.js", Code: "nope"}, {Path: "/App.js", Code: "x"}}); err == nil {
		t.Fatal("expected unknown path to be rejected")
	}
	if _, err := applyEdits(existing, []File{{Path: "/App.js", Code: "  "}, {Path: "/styles.css", Code: ""}}); err == nil {
		t.Fatal("expected empty App.js to be rejected")
	}
}
