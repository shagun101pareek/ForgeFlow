package generation

import (
	"archive/zip"
	"bytes"
	"fmt"
	"html"
	"strings"
	"unicode"
)

func BuildZip(projectName string, files []File) ([]byte, string, error) {
	byPath := map[string]string{}
	for _, file := range files {
		byPath[file.Path] = file.Code
	}
	app, ok := byPath["/App.js"]
	if !ok || strings.TrimSpace(app) == "" {
		return nil, "", fmt.Errorf("generation has no app file")
	}
	css := byPath["/styles.css"]

	slug := projectSlug(projectName)
	title := strings.TrimSpace(projectName)
	if title == "" {
		title = "ForgeFlow"
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	entries := map[string]string{
		"package.json":   packageJSON(slug),
		"vite.config.js": viteConfig,
		"index.html":     indexHTML(title),
		"src/main.jsx":   mainJSX,
		"src/App.jsx":    app,
		"src/styles.css": css,
	}
	for name, contents := range entries {
		writer, err := zw.Create(name)
		if err != nil {
			return nil, "", err
		}
		if _, err := writer.Write([]byte(contents)); err != nil {
			return nil, "", err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), slug + ".zip", nil
}

func projectSlug(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			dash = false
			continue
		}
		if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "forgeflow"
	}
	if len(slug) > 40 {
		slug = strings.Trim(slug[:40], "-")
	}
	if slug == "" {
		return "forgeflow"
	}
	return slug
}

func packageJSON(slug string) string {
	return fmt.Sprintf(`{
  "name": %q,
  "private": true,
  "scripts": {
    "dev": "vite",
    "build": "vite build"
  },
  "dependencies": {
    "react": "^19.0.0",
    "react-dom": "^19.0.0"
  },
  "devDependencies": {
    "@vitejs/plugin-react": "^4.3.4",
    "vite": "^6.2.0"
  }
}
`, slug)
}

func indexHTML(title string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>%s</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.jsx"></script>
  </body>
</html>
`, html.EscapeString(title))
}

const viteConfig = `import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
});
`

const mainJSX = `import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App.jsx";

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
`
