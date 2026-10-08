package generation

import (
	"strings"
	"testing"
)

func TestNormalizeKeepsInteractivePages(t *testing.T) {
	spec, err := Normalize(Spec{
		Project: Project{Name: "AI Notes"},
		Pages: []Page{
			{
				Name:  "Home",
				Route: "/",
				Sections: []Section{
					{Type: "navbar", Title: "AI Notes", Items: []Item{{Title: "Pricing", Route: "/pricing"}}},
					{Type: "hero", Title: `Say "hello" <script>`, PrimaryLabel: "Start", PrimaryRoute: "/signup"},
					{Type: "unknown", Title: "skip me"},
				},
			},
			{
				Name:  "Pricing",
				Route: "pricing",
				Sections: []Section{
					{Type: "pricing", Title: "Plans", Items: []Item{{Title: "Pro", Price: "$12", Description: "Notes | Search"}}},
				},
			},
		},
	}, "Fallback")
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if spec.Pages[0].Route != "/" {
		t.Fatalf("home route = %s", spec.Pages[0].Route)
	}
	if spec.Pages[1].Route != "/pricing" {
		t.Fatalf("pricing route = %s", spec.Pages[1].Route)
	}
	if len(spec.Pages[0].Sections) != 2 {
		t.Fatalf("sections = %d, want navbar and hero only", len(spec.Pages[0].Sections))
	}
}

func TestThemeClassMatchesTheFourLayouts(t *testing.T) {
	cases := []struct {
		spec Spec
		want string
	}{
		{dashboardSpec("Northwind"), "theme-dashboard"},
		{portfolioSpec("Northwind"), "theme-portfolio"},
		{pricingSpec("Northwind"), "theme-pricing"},
		{landingSpec("Northwind"), "theme-landing"},
	}
	for _, tc := range cases {
		code := GenerateFiles(tc.spec)[0].Code
		if !strings.Contains(code, tc.want) {
			t.Fatalf("missing %s", tc.want)
		}
	}
	pricing := GenerateFiles(pricingSpec("Northwind"))[0].Code
	if !strings.Contains(pricing, "price featured") {
		t.Fatal("middle plan was not featured")
	}
}

func TestCodegenEmitsInteractiveApp(t *testing.T) {
	spec, err := Normalize(Spec{
		Project: Project{Name: "AI Notes"},
		Pages: []Page{
			{
				Name:  "Home",
				Route: "/",
				Sections: []Section{
					{Type: "hero", Title: `Say "hello" <script>`, Subtitle: "Capture ideas", PrimaryLabel: "Get started", PrimaryRoute: "/signup"},
					{Type: "features", Title: "Features", Items: []Item{{Title: "Search", Description: "Find any note"}}},
				},
			},
			{
				Name:  "Signup",
				Route: "/signup",
				Sections: []Section{
					{Type: "signup", Title: "Create your account", PrimaryLabel: "Sign up"},
				},
			},
		},
	}, "AI Notes")
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}

	files := GenerateFiles(spec)
	if len(files) != 2 {
		t.Fatalf("files = %d", len(files))
	}
	if files[0].Path != "/App.js" || files[1].Path != "/styles.css" {
		t.Fatalf("paths = %s %s", files[0].Path, files[1].Path)
	}

	code := files[0].Code
	for _, want := range []string{
		`const INITIAL_ROUTE = "/"`,
		`setRoute`,
		`type="email"`,
		`type="password"`,
		`Say \"hello\" <script>`,
		`/signup`,
	} {
		if !strings.Contains(code, want) {
			t.Fatalf("generated app missing %q", want)
		}
	}
}
