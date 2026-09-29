package generation

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func TestBuildZipPackagesTheGeneratedApp(t *testing.T) {
	body, filename, err := BuildZip("North Star", []File{
		{Path: "/App.js", Code: "export default function App() { return <main>North Star</main>; }\n"},
		{Path: "/styles.css", Code: "body { margin: 0; }\n"},
		{Path: "../secret.txt", Code: "nope"},
	})
	if err != nil {
		t.Fatalf("build zip: %v", err)
	}
	if filename != "north-star.zip" {
		t.Fatalf("filename = %s", filename)
	}

	reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	found := map[string]string{}
	for _, file := range reader.File {
		if strings.Contains(file.Name, "..") || strings.HasPrefix(file.Name, "/") {
			t.Fatalf("unsafe path %s", file.Name)
		}
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("open %s: %v", file.Name, err)
		}
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(rc)
		rc.Close()
		found[file.Name] = buf.String()
	}

	for _, name := range []string{"package.json", "index.html", "vite.config.js", "src/main.jsx", "src/App.jsx", "src/styles.css"} {
		if _, ok := found[name]; !ok {
			t.Fatalf("missing %s in %v", name, keys(found))
		}
	}
	if _, ok := found["secret.txt"]; ok || strings.Contains(strings.Join(keys(found), " "), "secret") {
		t.Fatal("unexpected file was packaged")
	}
	if !strings.Contains(found["src/App.jsx"], "North Star") {
		t.Fatal("app source was not packaged")
	}
	if !strings.Contains(found["package.json"], `"name": "north-star"`) {
		t.Fatalf("package name = %s", found["package.json"])
	}
	if !strings.Contains(found["index.html"], "<title>North Star</title>") {
		t.Fatal("index title was not escaped into the page")
	}
}

func TestBuildZipRejectsAnEmptyGeneration(t *testing.T) {
	if _, _, err := BuildZip("Notes", nil); err == nil {
		t.Fatal("expected an error")
	}
}

func keys(items map[string]string) []string {
	names := make([]string, 0, len(items))
	for name := range items {
		names = append(names, name)
	}
	return names
}
