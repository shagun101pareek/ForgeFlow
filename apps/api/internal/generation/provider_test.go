package generation

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"
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
		{prompt: "landing page", want: "Built for the first version"},
		{prompt: "SaaS landing page with pricing", want: "Built for the first version"},
		{prompt: "team dashboard", want: "Recent work"},
		{prompt: "design portfolio", want: "Selected projects"},
		{prompt: "pricing page", want: "Start with a plan"},
		{prompt: "something else", want: "Built for the first version"},
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

func TestHeroSubtitleUsesThePrompt(t *testing.T) {
	notes, err := Mock{}.GenerateSpec(context.Background(), "Northstar", "A quiet notes app for people who write every morning.")
	if err != nil {
		t.Fatal(err)
	}
	bakery, err := Mock{}.GenerateSpec(context.Background(), "Northstar", "A neighborhood bakery with seasonal bread and a weekend market.")
	if err != nil {
		t.Fatal(err)
	}
	if heroSubtitle(notes) == heroSubtitle(bakery) {
		t.Fatal("two landing prompts shared a subtitle")
	}
	if heroSubtitle(notes) != "A quiet notes app for people who write every morning." {
		t.Fatalf("subtitle = %q", heroSubtitle(notes))
	}
	if heroTitle(notes) != "A quiet notes app for people" {
		t.Fatalf("headline = %q", heroTitle(notes))
	}
	if heroTitle(bakery) != "A neighborhood bakery with seasonal bread" {
		t.Fatalf("headline = %q", heroTitle(bakery))
	}
	if sectionTypes(notes) != sectionTypes(bakery) {
		t.Fatalf("sections changed: %s vs %s", sectionTypes(notes), sectionTypes(bakery))
	}

	multi, err := Mock{}.GenerateSpec(context.Background(), "Northstar", "Ship a notes app. Ignore this second sentence.")
	if err != nil {
		t.Fatal(err)
	}
	if heroSubtitle(multi) != "Ship a notes app." {
		t.Fatalf("subtitle = %q", heroSubtitle(multi))
	}
	if heroTitle(multi) != "Ship a notes app" {
		t.Fatalf("headline = %q", heroTitle(multi))
	}

	long := strings.Repeat("seasonal ", 30)
	clipped, err := Mock{}.GenerateSpec(context.Background(), "Northstar", long)
	if err != nil {
		t.Fatal(err)
	}
	if n := utf8.RuneCountInString(heroSubtitle(clipped)); n == 0 || n > maxHeroSentence {
		t.Fatalf("subtitle length = %d (%q)", n, heroSubtitle(clipped))
	}

	blank, err := Mock{}.GenerateSpec(context.Background(), "Northstar", "   ")
	if err != nil {
		t.Fatal(err)
	}
	if heroSubtitle(blank) != "A focused workspace for teams that want a clear product story, a simple signup, and a plan that fits." {
		t.Fatalf("blank prompt subtitle = %q", heroSubtitle(blank))
	}
	if heroTitle(blank) != "Northstar keeps the work moving" {
		t.Fatalf("blank prompt headline = %q", heroTitle(blank))
	}
}

func TestFeatureCardsUsePromptList(t *testing.T) {
	plain, err := Mock{}.GenerateSpec(context.Background(), "Northstar", "A quiet notes app for people who write every morning.")
	if err != nil {
		t.Fatal(err)
	}
	if titles := featureTitles(plain); titles[0] != "Shared projects" || titles[1] != "Live preview" || titles[2] != "Clear plans" {
		t.Fatalf("titles = %#v", titles)
	}

	listed, err := Mock{}.GenerateSpec(context.Background(), "Northstar", "A notes app with capture, search, and sharing.")
	if err != nil {
		t.Fatal(err)
	}
	titles := featureTitles(listed)
	if len(titles) != 3 || titles[0] != "A notes app with capture" || titles[1] != "search" || titles[2] != "sharing" {
		t.Fatalf("titles = %#v", titles)
	}
	if descs := featureDescriptions(listed); descs[0] != "Keep every prototype in one place instead of a folder of screenshots." || descs[1] != "Click through the generated interface instead of imagining the interactions." || descs[2] != "Show pricing without sending people to a separate document." {
		t.Fatalf("descriptions = %#v", descs)
	}
	if sectionTypes(plain) != sectionTypes(listed) {
		t.Fatalf("sections changed: %s vs %s", sectionTypes(plain), sectionTypes(listed))
	}

	partial, err := Mock{}.GenerateSpec(context.Background(), "Northstar", "A notes app with capture, search.")
	if err != nil {
		t.Fatal(err)
	}
	partialTitles := featureTitles(partial)
	if len(partialTitles) != 3 || partialTitles[0] != "A notes app with capture" || partialTitles[1] != "search" || partialTitles[2] != "Clear plans" {
		t.Fatalf("titles = %#v", partialTitles)
	}

	extra, err := Mock{}.GenerateSpec(context.Background(), "Northstar", "A notes app with capture, search, sharing, and export.")
	if err != nil {
		t.Fatal(err)
	}
	extraTitles := featureTitles(extra)
	if len(extraTitles) != 3 || extraTitles[2] != "sharing" {
		t.Fatalf("titles = %#v", extraTitles)
	}

	later, err := Mock{}.GenerateSpec(context.Background(), "Northstar", "Ship a notes app. Then add capture, search, and sharing.")
	if err != nil {
		t.Fatal(err)
	}
	if laterTitles := featureTitles(later); laterTitles[0] != "Shared projects" {
		t.Fatalf("titles = %#v", laterTitles)
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

func featuresSection(spec Spec) Section {
	for _, page := range spec.Pages {
		if page.Route != "/" {
			continue
		}
		for _, section := range page.Sections {
			if section.Type == "features" {
				return section
			}
		}
	}
	return Section{}
}

func featureTitles(spec Spec) []string {
	items := featuresSection(spec).Items
	titles := make([]string, len(items))
	for i, item := range items {
		titles[i] = item.Title
	}
	return titles
}

func featureDescriptions(spec Spec) []string {
	items := featuresSection(spec).Items
	descriptions := make([]string, len(items))
	for i, item := range items {
		descriptions[i] = item.Description
	}
	return descriptions
}

func heroSection(spec Spec) Section {
	for _, page := range spec.Pages {
		if page.Route != "/" {
			continue
		}
		for _, section := range page.Sections {
			if section.Type == "hero" {
				return section
			}
		}
	}
	return Section{}
}

func heroTitle(spec Spec) string {
	return heroSection(spec).Title
}

func heroSubtitle(spec Spec) string {
	return heroSection(spec).Subtitle
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
