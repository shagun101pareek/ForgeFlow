package generation

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type File struct {
	Path string `json:"path"`
	Code string `json:"code"`
}

func GenerateFiles(spec Spec) []File {
	return []File{
		{Path: "/App.js", Code: generateApp(spec)},
		{Path: "/styles.css", Code: appCSS},
	}
}

func generateApp(spec Spec) string {
	var b strings.Builder
	b.WriteString(`import "./styles.css";
import { useState } from "react";

const INITIAL_ROUTE = "/";

export default function App() {
  const [route, setRoute] = useState(INITIAL_ROUTE);
  const [notice, setNotice] = useState("");

  return (
    <div className="app">
`)

	used := map[string]struct{}{"App": {}, "Nav": {}}
	names := make([]string, len(spec.Pages))
	for i, page := range spec.Pages {
		names[i] = uniqueComponent(page.Name, used)
		b.WriteString(fmt.Sprintf("      {route === %s ? <%s route={route} setRoute={setRoute} notice={notice} setNotice={setNotice} /> : null}\n", jsString(page.Route), names[i]))
	}

	b.WriteString(`    </div>
  );
}

`)

	brand := spec.Project.Name
	for _, page := range spec.Pages {
		for _, section := range page.Sections {
			if section.Type == "navbar" && section.Title != "" {
				brand = section.Title
			}
		}
	}

	b.WriteString("function Nav({ route, setRoute, setNotice }) {\n  const links = [\n")
	for _, page := range spec.Pages {
		b.WriteString(fmt.Sprintf("    { label: %s, route: %s },\n", jsString(page.Name), jsString(page.Route)))
	}
	b.WriteString(fmt.Sprintf(`  ];
  return (
    <header className="nav">
      <button type="button" className="brand" onClick={() => { setNotice(""); setRoute("/"); }}>%s</button>
      <nav className="links">
        {links.map((link) => (
          <button
            key={link.route}
            type="button"
            className={route === link.route ? "link active" : "link"}
            onClick={() => { setNotice(""); setRoute(link.route); }}
          >
            {link.label}
          </button>
        ))}
      </nav>
    </header>
  );
}

`, jsText(brand)))

	routes := make(map[string]struct{}, len(spec.Pages))
	for _, page := range spec.Pages {
		routes[page.Route] = struct{}{}
	}

	for i, page := range spec.Pages {
		b.WriteString(fmt.Sprintf("function %s({ route, setRoute, notice, setNotice }) {\n  return (\n    <>\n      <Nav route={route} setRoute={setRoute} setNotice={setNotice} />\n      <main>\n", names[i]))
		for _, section := range page.Sections {
			if section.Type == "navbar" {
				continue
			}
			b.WriteString(renderSection(section, routes))
		}
		b.WriteString("      </main>\n    </>\n  );\n}\n\n")
	}

	return b.String()
}

func renderSection(section Section, routes map[string]struct{}) string {
	switch section.Type {
	case "hero":
		return fmt.Sprintf(`        <section className="hero">
          %s
          <p className="eyebrow">Prototype</p>
          <h1>%s</h1>
          %s
          <div className="actions">
            %s
            %s
          </div>
        </section>
`, heroImage(section.Image), jsText(section.Title), paragraph(section.Subtitle), actionButton("btn", section.PrimaryLabel, section.PrimaryRoute, routes), actionButton("btn secondary", section.SecondaryLabel, section.SecondaryRoute, routes))
	case "features":
		return fmt.Sprintf(`        <section className="section">
          <h2>%s</h2>
          %s
          <div className="grid">
            %s
          </div>
        </section>
`, jsText(section.Title), paragraph(section.Subtitle), featureCards(section.Items))
	case "pricing":
		return fmt.Sprintf(`        <section className="section">
          <h2>%s</h2>
          %s
          <div className="grid">
            %s
          </div>
        </section>
`, jsText(section.Title), paragraph(section.Subtitle), pricingCards(section.Items, routes))
	case "testimonials":
		return fmt.Sprintf(`        <section className="section">
          <h2>%s</h2>
          <div className="grid">
            %s
          </div>
        </section>
`, jsText(section.Title), testimonialCards(section.Items))
	case "stats":
		return fmt.Sprintf(`        <section className="stats">
          %s
        </section>
`, statItems(section.Items))
	case "faq":
		return fmt.Sprintf(`        <section className="section">
          <h2>%s</h2>
          <div className="faq">
            %s
          </div>
        </section>
`, jsText(section.Title), faqItems(section.Items))
	case "signup":
		label := section.PrimaryLabel
		if label == "" {
			label = "Create account"
		}
		title := section.Title
		if title == "" {
			title = "Create an account"
		}
		return fmt.Sprintf(`        <section className="section narrow">
          <h2>%s</h2>
          %s
          <form className="form" onSubmit={(event) => { event.preventDefault(); setNotice("Account created. You can keep exploring."); }}>
            <label>Name<input name="name" autoComplete="name" required /></label>
            <label>Email<input type="email" name="email" autoComplete="email" required /></label>
            <label>Password<input type="password" name="password" autoComplete="new-password" minLength={8} required /></label>
            <button type="submit" className="btn">%s</button>
          </form>
          {notice ? <p className="notice">{notice}</p> : null}
        </section>
`, jsText(title), paragraph(section.Subtitle), jsText(label))
	case "cta":
		return fmt.Sprintf(`        <section className="cta">
          <h2>%s</h2>
          %s
          %s
        </section>
`, jsText(section.Title), paragraph(section.Subtitle), actionButton("btn", section.PrimaryLabel, section.PrimaryRoute, routes))
	case "footer":
		return fmt.Sprintf(`        <footer className="footer">
          <strong>%s</strong>
          %s
        </footer>
`, jsText(section.Title), paragraph(section.Subtitle))
	default:
		return ""
	}
}

func featureCards(items []Item) string {
	var b strings.Builder
	for _, item := range items {
		b.WriteString(fmt.Sprintf(`            <article className="card">
              <h3>%s</h3>
              %s
            </article>
`, jsText(item.Title), paragraph(item.Description)))
	}
	return b.String()
}

func pricingCards(items []Item, routes map[string]struct{}) string {
	var b strings.Builder
	for _, item := range items {
		b.WriteString(fmt.Sprintf(`            <article className="card price">
              <h3>%s</h3>
              <p className="amount">%s</p>
              <ul>
                %s
              </ul>
              %s
            </article>
`, jsText(item.Title), jsText(item.Price), featureList(item.Description), actionButton("btn", "Choose "+item.Title, item.Route, routes)))
	}
	return b.String()
}

func testimonialCards(items []Item) string {
	var b strings.Builder
	for _, item := range items {
		b.WriteString(fmt.Sprintf(`            <blockquote className="card">
              %s
              <footer>%s</footer>
            </blockquote>
`, paragraph(item.Description), jsText(item.Title)))
	}
	return b.String()
}

func statItems(items []Item) string {
	var b strings.Builder
	for _, item := range items {
		b.WriteString(fmt.Sprintf(`          <div>
            <strong>%s</strong>
            <span>%s</span>
          </div>
`, jsText(item.Title), jsText(item.Description)))
	}
	return b.String()
}

func faqItems(items []Item) string {
	var b strings.Builder
	for _, item := range items {
		b.WriteString(fmt.Sprintf(`            <details>
              <summary>%s</summary>
              %s
            </details>
`, jsText(item.Title), paragraph(item.Description)))
	}
	return b.String()
}

func featureList(description string) string {
	parts := strings.Split(description, "|")
	var b strings.Builder
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("                <li>%s</li>\n", jsText(part)))
	}
	return b.String()
}

func actionButton(className, label, route string, routes map[string]struct{}) string {
	label = strings.TrimSpace(label)
	if label == "" {
		return ""
	}
	if _, ok := routes[route]; !ok {
		route = ""
	}
	if route == "" {
		return fmt.Sprintf(`<button type="button" className=%s onClick={() => setNotice(%s)}>%s</button>`, jsString(className), jsString("Thanks"), jsText(label))
	}
	return fmt.Sprintf(`<button type="button" className=%s onClick={() => { setNotice(""); setRoute(%s); }}>%s</button>`, jsString(className), jsString(route), jsText(label))
}

func heroImage(src string) string {
	if !strings.HasPrefix(src, "data:image/") {
		return ""
	}
	return fmt.Sprintf(`<img className="hero-image" alt="" src={%s} />`, jsString(src))
}

func paragraph(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return "<p>" + jsText(value) + "</p>"
}

func jsText(value string) string {
	return "{" + jsString(value) + "}"
}

func jsString(value string) string {
	return strconv.Quote(value)
}

func uniqueComponent(name string, used map[string]struct{}) string {
	var b strings.Builder
	upper := true
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if upper {
				b.WriteRune(unicode.ToUpper(r))
				upper = false
			} else {
				b.WriteRune(r)
			}
			continue
		}
		upper = true
	}
	result := b.String()
	if result == "" {
		result = "Page"
	} else if unicode.IsDigit([]rune(result)[0]) {
		result = "Page" + result
	}
	base := result
	for i := 2; ; i++ {
		if _, exists := used[result]; !exists {
			used[result] = struct{}{}
			return result
		}
		result = fmt.Sprintf("%s%d", base, i)
	}
}

const appCSS = `:root {
  color-scheme: light;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  color: #172033;
  background: #f6f7fb;
}

* { box-sizing: border-box; }

body { margin: 0; }

button, input { font: inherit; }

.app { min-height: 100vh; }

.nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 24px;
  background: #ffffff;
  border-bottom: 1px solid #e6e8f0;
  position: sticky;
  top: 0;
}

.brand, .link, .btn {
  border: 0;
  cursor: pointer;
  background: transparent;
}

.brand { font-weight: 700; font-size: 16px; color: #172033; }

.links { display: flex; gap: 8px; flex-wrap: wrap; }

.link { color: #526075; padding: 8px 10px; border-radius: 999px; }

.link.active, .link:hover { background: #eef2ff; color: #243056; }

main { max-width: 1040px; margin: 0 auto; padding: 32px 20px 64px; }

.hero { padding: 48px 0 24px; }

.hero-image { display: block; width: 100%; max-height: 280px; object-fit: cover; border-radius: 16px; margin-bottom: 24px; }

.eyebrow {
  margin: 0 0 12px;
  color: #4452d6;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

h1, h2, h3 { margin: 0; letter-spacing: -0.03em; }

h1 { font-size: 48px; line-height: 1.05; max-width: 14ch; }

h2 { font-size: 32px; }

.hero p, .section p, .cta p { color: #526075; line-height: 1.6; }

.actions, .grid, .stats { display: flex; gap: 16px; flex-wrap: wrap; }

.actions { margin-top: 24px; }

.btn {
  background: #243056;
  color: white;
  border-radius: 999px;
  padding: 12px 18px;
}

.btn.secondary { background: #e8ebf5; color: #243056; }

.section, .cta { margin-top: 40px; }

.grid { margin-top: 20px; }

.card, .cta, .faq details {
  background: white;
  border: 1px solid #e6e8f0;
  border-radius: 18px;
  padding: 20px;
}

.card { flex: 1 1 220px; }

.price .amount { font-size: 28px; color: #172033; font-weight: 700; }

.card ul { padding-left: 18px; color: #526075; }

.stats { margin-top: 28px; }

.stats div {
  background: white;
  border-radius: 18px;
  padding: 18px 20px;
  min-width: 140px;
  border: 1px solid #e6e8f0;
}

.stats strong { display: block; font-size: 28px; }

.stats span, .card footer { color: #526075; }

.faq { display: grid; gap: 12px; margin-top: 16px; }

.faq summary { cursor: pointer; font-weight: 600; }

.narrow { max-width: 460px; }

.form { display: grid; gap: 12px; margin-top: 16px; }

.form label { display: grid; gap: 6px; font-size: 14px; font-weight: 600; }

.form input {
  border: 1px solid #d5d9e6;
  border-radius: 12px;
  padding: 10px 12px;
}

.notice {
  margin-top: 16px;
  background: #e9f8ef;
  color: #146c43;
  border-radius: 12px;
  padding: 12px 14px;
}

.footer { margin-top: 48px; color: #526075; }

.cta { padding: 28px; }

@media (max-width: 720px) {
  h1 { font-size: 36px; }
  .nav { align-items: flex-start; flex-direction: column; }
}
`
