package generation

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maxPages           = 4
	maxSectionsPerPage = 8
	maxItemsPerSection = 6
	maxTitleLen        = 120
	maxSubtitleLen     = 400
	maxItemLen         = 240
)

var (
	routePattern = regexp.MustCompile(`^/[a-z0-9/-]*$`)
	allowedTypes = map[string]struct{}{
		"navbar":       {},
		"hero":         {},
		"features":     {},
		"pricing":      {},
		"testimonials": {},
		"signup":       {},
		"cta":          {},
		"faq":          {},
		"stats":        {},
		"footer":       {},
	}
)

type Spec struct {
	Project Project `json:"project"`
	Pages   []Page  `json:"pages"`
}

type Project struct {
	Name string `json:"name"`
}

type Page struct {
	Name     string    `json:"name"`
	Route    string    `json:"route"`
	Sections []Section `json:"sections"`
}

type Section struct {
	Type           string `json:"type"`
	Title          string `json:"title"`
	Subtitle       string `json:"subtitle"`
	PrimaryLabel   string `json:"primaryLabel"`
	PrimaryRoute   string `json:"primaryRoute"`
	SecondaryLabel string `json:"secondaryLabel"`
	SecondaryRoute string `json:"secondaryRoute"`
	Image          string `json:"image,omitempty"`
	Items          []Item `json:"items"`
}

type Item struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Route       string `json:"route"`
}

func Normalize(spec Spec, fallbackName string) (Spec, error) {
	spec.Project.Name = clip(spec.Project.Name, maxTitleLen)
	if spec.Project.Name == "" {
		spec.Project.Name = clip(fallbackName, maxTitleLen)
	}
	if spec.Project.Name == "" {
		spec.Project.Name = "ForgeFlow"
	}

	if len(spec.Pages) > maxPages {
		spec.Pages = spec.Pages[:maxPages]
	}

	pages := make([]Page, 0, len(spec.Pages))
	seenRoutes := map[string]struct{}{}
	for i, page := range spec.Pages {
		page.Name = clip(page.Name, maxTitleLen)
		if page.Name == "" {
			page.Name = "Page"
		}
		page.Route = normalizeRoute(page.Route, page.Name, i)
		if _, exists := seenRoutes[page.Route]; exists {
			continue
		}
		seenRoutes[page.Route] = struct{}{}

		if len(page.Sections) > maxSectionsPerPage {
			page.Sections = page.Sections[:maxSectionsPerPage]
		}
		sections := make([]Section, 0, len(page.Sections))
		for _, section := range page.Sections {
			section.Type = strings.ToLower(strings.TrimSpace(section.Type))
			if _, ok := allowedTypes[section.Type]; !ok {
				continue
			}
			section.Title = clip(section.Title, maxTitleLen)
			section.Subtitle = clip(section.Subtitle, maxSubtitleLen)
			section.PrimaryLabel = clip(section.PrimaryLabel, 40)
			section.SecondaryLabel = clip(section.SecondaryLabel, 40)
			section.PrimaryRoute = cleanRoute(section.PrimaryRoute)
			section.SecondaryRoute = cleanRoute(section.SecondaryRoute)
			if len(section.Items) > maxItemsPerSection {
				section.Items = section.Items[:maxItemsPerSection]
			}
			items := make([]Item, 0, len(section.Items))
			for _, item := range section.Items {
				item.Title = clip(item.Title, maxItemLen)
				item.Description = clip(item.Description, maxItemLen)
				item.Price = clip(item.Price, 40)
				item.Route = cleanRoute(item.Route)
				if item.Title == "" && item.Description == "" && item.Price == "" {
					continue
				}
				items = append(items, item)
			}
			section.Items = items
			sections = append(sections, section)
		}
		if len(sections) == 0 {
			sections = append(sections, Section{
				Type:  "hero",
				Title: page.Name,
			})
		}
		page.Sections = sections
		pages = append(pages, page)
	}

	if len(pages) == 0 {
		pages = append(pages, Page{
			Name:  "Home",
			Route: "/",
			Sections: []Section{{
				Type:     "hero",
				Title:    spec.Project.Name,
				Subtitle: "Describe the product to shape this page.",
			}},
		})
	}
	if pages[0].Route != "/" {
		pages[0].Route = "/"
	}

	spec.Pages = pages
	return spec, nil
}

func normalizeRoute(route, name string, index int) string {
	route = cleanRoute(route)
	if route == "/" || routePattern.MatchString(route) && route != "" {
		if index == 0 {
			return "/"
		}
		if route == "/" {
			return "/" + slug(name)
		}
		return route
	}
	if index == 0 {
		return "/"
	}
	return "/" + slug(name)
}

func cleanRoute(route string) string {
	route = strings.ToLower(strings.TrimSpace(route))
	if route == "" {
		return ""
	}
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	route = strings.ReplaceAll(route, " ", "")
	if route != "/" && !routePattern.MatchString(route) {
		return ""
	}
	return route
}

func slug(name string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "page"
	}
	return out
}

func clip(value string, max int) string {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) <= max {
		return value
	}
	runes := []rune(value)
	return string(runes[:max])
}
