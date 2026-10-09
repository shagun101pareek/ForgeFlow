package generation

import (
	"context"
	"strings"
	"testing"
)

func TestMockProviderReturnsExistingSpec(t *testing.T) {
	spec, err := Mock{}.GenerateSpec(context.Background(), "Northstar", "Build a SaaS landing page")
	if err != nil {
		t.Fatalf("generate spec: %v", err)
	}
	if spec.Project.Name != "Northstar" {
		t.Fatalf("project = %s", spec.Project.Name)
	}
	if spec.Pages[0].Route != "/" {
		t.Fatalf("first route = %s", spec.Pages[0].Route)
	}

	types := sectionTypes(spec)
	for _, want := range []string{"navbar", "hero", "features", "cta", "pricing", "signup", "footer"} {
		if !strings.Contains(types, want) {
			t.Fatalf("spec missing %s in %s", want, types)
		}
	}
	for _, absent := range []string{"testimonials", "faq", "stats"} {
		if strings.Contains(types, absent) {
			t.Fatalf("default landing included %s", absent)
		}
	}

	normalized, err := Normalize(spec, "Northstar")
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	files := GenerateFiles(normalized)
	if len(files) != 2 || files[0].Path != "/App.js" {
		t.Fatalf("files = %#v", files)
	}
	if !strings.Contains(files[0].Code, "setRoute") || !strings.Contains(files[0].Code, `type="email"`) {
		t.Fatal("mock spec did not reach the code generator")
	}
}

func TestPromptAddsRequestedSections(t *testing.T) {
	plain, err := Mock{}.GenerateSpec(context.Background(), "Northstar", "A SaaS landing page for a product team")
	if err != nil {
		t.Fatal(err)
	}
	plainTypes := sectionTypes(plain)
	for _, absent := range []string{"testimonials", "faq", "stats"} {
		if strings.Contains(plainTypes, absent) {
			t.Fatalf("plain landing included %s", absent)
		}
	}

	asked, err := Mock{}.GenerateSpec(context.Background(), "Northstar", "A SaaS landing page with testimonials, an FAQ, and stats")
	if err != nil {
		t.Fatal(err)
	}
	askedTypes := sectionTypes(asked)
	for _, want := range []string{"testimonials", "faq", "stats"} {
		if !strings.Contains(askedTypes, want) {
			t.Fatalf("requested landing missing %s in %s", want, askedTypes)
		}
	}
	if themeClass(asked) != "theme-landing" {
		t.Fatalf("theme = %s", themeClass(asked))
	}

	dashboard, err := Mock{}.GenerateSpec(context.Background(), "Northstar", "A team dashboard for active projects")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sectionTypes(dashboard), "stats") || themeClass(dashboard) != "theme-dashboard" {
		t.Fatalf("dashboard = %s %s", themeClass(dashboard), sectionTypes(dashboard))
	}
}

func TestMockProviderPatternsAreDeterministic(t *testing.T) {
	prompts := []struct {
		prompt string
		want   string
	}{
		{prompt: "landing page", want: "keeps the work moving"},
		{prompt: "SaaS landing page with pricing", want: "keeps the work moving"},
		{prompt: "team dashboard", want: "overview"},
		{prompt: "design portfolio", want: "Selected product work"},
		{prompt: "pricing page", want: "Pricing for"},
		{prompt: "something else", want: "keeps the work moving"},
	}

	for _, tc := range prompts {
		first, err := Mock{}.GenerateSpec(context.Background(), "ForgeFlow", tc.prompt)
		if err != nil {
			t.Fatalf("%s: %v", tc.prompt, err)
		}
		second, err := Mock{}.GenerateSpec(context.Background(), "ForgeFlow", tc.prompt)
		if err != nil {
			t.Fatalf("%s again: %v", tc.prompt, err)
		}
		if specText(first) != specText(second) {
			t.Fatalf("%s was not deterministic", tc.prompt)
		}
		if !strings.Contains(specText(first), tc.want) {
			t.Fatalf("%s produced %s", tc.prompt, specText(first))
		}
	}
}

func TestProviderSelection(t *testing.T) {
	mockProvider, err := NewProvider("mock", "")
	if err != nil {
		t.Fatalf("mock provider: %v", err)
	}
	if _, ok := mockProvider.(Mock); !ok {
		t.Fatalf("mock provider type = %T", mockProvider)
	}
	if _, err := mockProvider.GenerateSpec(context.Background(), "ForgeFlow", "landing page"); err != nil {
		t.Fatalf("mock generation without an API key: %v", err)
	}

	openAIProvider, err := NewProvider("openai", "")
	if err != nil {
		t.Fatalf("openai provider: %v", err)
	}
	if _, ok := openAIProvider.(*OpenAI); !ok {
		t.Fatalf("openai provider type = %T", openAIProvider)
	}
	_, err = openAIProvider.GenerateSpec(context.Background(), "ForgeFlow", "landing page")
	if err == nil || !strings.Contains(err.Error(), "api key is not configured") {
		t.Fatalf("openai error = %v", err)
	}

	if _, err := NewProvider("local", ""); err == nil {
		t.Fatal("unknown provider was accepted")
	}
}

func sectionTypes(spec Spec) string {
	var b strings.Builder
	for _, page := range spec.Pages {
		for _, section := range page.Sections {
			b.WriteString(section.Type)
			b.WriteByte(' ')
		}
	}
	return b.String()
}

func specText(spec Spec) string {
	var b strings.Builder
	b.WriteString(spec.Project.Name)
	for _, page := range spec.Pages {
		b.WriteString(page.Name)
		b.WriteString(page.Route)
		for _, section := range page.Sections {
			b.WriteString(section.Type)
			b.WriteString(section.Title)
			b.WriteString(section.Subtitle)
		}
	}
	return b.String()
}
