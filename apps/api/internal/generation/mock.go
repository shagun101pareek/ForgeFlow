package generation

import (
	"context"
	"strings"
	"unicode/utf8"
)

type Mock struct{}

func (Mock) GenerateSpec(_ context.Context, projectName, prompt string) (Spec, error) {
	name := strings.TrimSpace(projectName)
	if name == "" {
		name = "ForgeFlow"
	}
	var spec Spec
	switch classifyPrompt(prompt) {
	case "dashboard":
		spec = dashboardSpec(name)
	case "portfolio":
		spec = portfolioSpec(name)
	case "pricing":
		spec = pricingSpec(name)
	default:
		spec = landingSpec(name)
	}
	return shapeSpec(spec, name, prompt), nil
}

func classifyPrompt(prompt string) string {
	text := strings.ToLower(prompt)
	switch {
	case strings.Contains(text, "dashboard"):
		return "dashboard"
	case strings.Contains(text, "portfolio"):
		return "portfolio"
	case strings.Contains(text, "pricing page") || (strings.Contains(text, "pricing") && !strings.Contains(text, "landing")):
		return "pricing"
	default:
		return "landing"
	}
}

func landingSpec(name string) Spec {
	return Spec{
		Project: Project{Name: name},
		Pages: []Page{
			{
				Name:  "Home",
				Route: "/",
				Sections: []Section{
					navbar(name, []Item{{Title: "Pricing", Route: "/pricing"}, {Title: "Sign up", Route: "/signup"}}),
					{
						Type:           "hero",
						Title:          name + " keeps the work moving",
						Subtitle:       "A focused workspace for teams that want a clear product story, a simple signup, and a plan that fits.",
						PrimaryLabel:   "Start free",
						PrimaryRoute:   "/signup",
						SecondaryLabel: "View pricing",
						SecondaryRoute: "/pricing",
					},
					{
						Type:     "features",
						Title:    "Built for the first version",
						Subtitle: "The pieces a SaaS landing page needs before the product gets complicated.",
						Items: []Item{
							{Title: "Shared projects", Description: "Keep every prototype in one place instead of a folder of screenshots."},
							{Title: "Live preview", Description: "Click through the generated interface instead of imagining the interactions."},
							{Title: "Clear plans", Description: "Show pricing without sending people to a separate document."},
						},
					},
					{
						Type:  "testimonials",
						Title: "Teams using " + name,
						Items: []Item{
							{Title: "Ava Chen", Description: "We replaced a week of static mockups with a page people could actually click."},
							{Title: "Noah Patel", Description: "The signup flow made the prototype feel like the product, not a slide."},
						},
					},
					{
						Type:         "cta",
						Title:        "Ready to try " + name + "?",
						Subtitle:     "Create an account and open the workspace.",
						PrimaryLabel: "Create account",
						PrimaryRoute: "/signup",
					},
					footer(name),
				},
			},
			{
				Name:  "Pricing",
				Route: "/pricing",
				Sections: []Section{
					{
						Type:     "pricing",
						Title:    "Simple pricing",
						Subtitle: "Start small. Upgrade when the team grows.",
						Items: []Item{
							{Title: "Starter", Price: "$0", Description: "1 project | Live preview | Community support", Route: "/signup"},
							{Title: "Pro", Price: "$19", Description: "Unlimited projects | Custom pages | Email support", Route: "/signup"},
							{Title: "Team", Price: "$49", Description: "Shared workspace | Priority support | Export", Route: "/signup"},
						},
					},
					{
						Type:  "faq",
						Title: "Questions",
						Items: []Item{
							{Title: "Can I change plans later?", Description: "Yes. Pick a plan now and switch when the prototype is ready to share."},
							{Title: "Is there a free start?", Description: "The Starter plan is free and includes a live preview."},
						},
					},
				},
			},
			{
				Name:  "Sign up",
				Route: "/signup",
				Sections: []Section{
					{
						Type:         "signup",
						Title:        "Create your " + name + " account",
						Subtitle:     "Use your email to open the workspace.",
						PrimaryLabel: "Create account",
					},
				},
			},
		},
	}
}

func dashboardSpec(name string) Spec {
	return Spec{
		Project: Project{Name: name},
		Pages: []Page{
			{
				Name:  "Overview",
				Route: "/",
				Sections: []Section{
					navbar(name, []Item{{Title: "Projects", Route: "/projects"}}),
					{
						Type:         "hero",
						Title:        name + " overview",
						Subtitle:     "See active work, recent updates, and the next thing to review.",
						PrimaryLabel: "New project",
						PrimaryRoute: "/projects",
					},
					{
						Type:  "stats",
						Title: "Today",
						Items: []Item{
							{Title: "12", Description: "Active projects"},
							{Title: "4", Description: "Waiting for review"},
							{Title: "28", Description: "Previews this week"},
						},
					},
					{
						Type:  "features",
						Title: "Recent work",
						Items: []Item{
							{Title: "Launch page", Description: "Hero, pricing, and signup are ready for review."},
							{Title: "Customer portal", Description: "Dashboard cards cover projects and account status."},
							{Title: "Onboarding", Description: "A short form collects the account details."},
						},
					},
				},
			},
			{
				Name:  "Projects",
				Route: "/projects",
				Sections: []Section{
					{
						Type:     "features",
						Title:    "Projects",
						Subtitle: "Open a card to keep working.",
						Items: []Item{
							{Title: "Northstar", Description: "Marketing site with pricing and a signup form."},
							{Title: "Atlas", Description: "Internal dashboard for the operations team."},
							{Title: "Harbor", Description: "Portfolio page with selected case studies."},
						},
					},
					{
						Type:         "cta",
						Title:        "Start another project",
						Subtitle:     "Keep the next prototype in the same workspace.",
						PrimaryLabel: "Back to overview",
						PrimaryRoute: "/",
					},
				},
			},
		},
	}
}

func portfolioSpec(name string) Spec {
	return Spec{
		Project: Project{Name: name},
		Pages: []Page{
			{
				Name:  "Work",
				Route: "/",
				Sections: []Section{
					navbar(name, []Item{{Title: "Contact", Route: "/contact"}}),
					{
						Type:           "hero",
						Title:          name,
						Subtitle:       "Selected product work, from first sketch to a clickable interface.",
						PrimaryLabel:   "View work",
						SecondaryLabel: "Contact",
						SecondaryRoute: "/contact",
					},
					{
						Type:  "features",
						Title: "Selected projects",
						Items: []Item{
							{Title: "Fieldnotes", Description: "A writing product with a quiet reading layout and a simple account form."},
							{Title: "Lumen", Description: "A SaaS marketing page with pricing, testimonials, and a clear call to action."},
							{Title: "Kin", Description: "A team dashboard that leads with status and recent work."},
						},
					},
					{
						Type:  "testimonials",
						Title: "Notes from collaborators",
						Items: []Item{
							{Title: "Maya Ortiz", Description: "The portfolio felt like the product: direct, clickable, and easy to follow."},
							{Title: "Jonah Ellis", Description: "We could move from the work to a conversation without leaving the page."},
						},
					},
					footer(name),
				},
			},
			{
				Name:  "Contact",
				Route: "/contact",
				Sections: []Section{
					{
						Type:         "signup",
						Title:        "Start a project with " + name,
						Subtitle:     "Send your name and email. The form confirms the message in the preview.",
						PrimaryLabel: "Send message",
					},
					{
						Type:         "cta",
						Title:        "Prefer to look around first?",
						PrimaryLabel: "Back to work",
						PrimaryRoute: "/",
					},
				},
			},
		},
	}
}

func pricingSpec(name string) Spec {
	return Spec{
		Project: Project{Name: name},
		Pages: []Page{
			{
				Name:  "Pricing",
				Route: "/",
				Sections: []Section{
					navbar(name, []Item{{Title: "Sign up", Route: "/signup"}}),
					{
						Type:         "hero",
						Title:        "Pricing for " + name,
						Subtitle:     "Pick a plan, compare what is included, and continue to signup.",
						PrimaryLabel: "Choose a plan",
						PrimaryRoute: "/signup",
					},
					{
						Type:  "pricing",
						Title: "Plans",
						Items: []Item{
							{Title: "Starter", Price: "$0", Description: "1 project | Preview | Email signup", Route: "/signup"},
							{Title: "Studio", Price: "$24", Description: "10 projects | Custom pages | Testimonials", Route: "/signup"},
							{Title: "Company", Price: "$79", Description: "Unlimited projects | Shared review | Support", Route: "/signup"},
						},
					},
					{
						Type:  "faq",
						Title: "Before you choose",
						Items: []Item{
							{Title: "What happens after signup?", Description: "You land in the workspace with the plan you selected."},
							{Title: "Can I stay on Starter?", Description: "Yes. Starter is enough to try the preview and the signup flow."},
						},
					},
					{
						Type:         "cta",
						Title:        "Start with a plan",
						PrimaryLabel: "Create account",
						PrimaryRoute: "/signup",
					},
					footer(name),
				},
			},
			{
				Name:  "Sign up",
				Route: "/signup",
				Sections: []Section{
					{
						Type:         "signup",
						Title:        "Create your account",
						Subtitle:     "The form stays in the preview so you can try the flow.",
						PrimaryLabel: "Create account",
					},
				},
			},
		},
	}
}

func shapeSpec(spec Spec, name, prompt string) Spec {
	text := strings.ToLower(prompt)
	wantTestimonials := strings.Contains(text, "testimonial") || strings.Contains(text, "quotes")
	wantFAQ := strings.Contains(text, "faq") || strings.Contains(text, "question")
	wantStats := strings.Contains(text, "stats") || strings.Contains(text, "statistics") || strings.Contains(text, "metrics")

	if wantTestimonials {
		spec = ensureSection(spec, "/", "features", testimonialSection(name))
	} else {
		spec = dropSection(spec, "testimonials")
	}
	if wantFAQ {
		spec = ensureFAQ(spec, name)
	} else {
		spec = dropSection(spec, "faq")
	}
	if wantStats {
		spec = ensureSection(spec, "/", "hero", statsSection())
	}
	return applyHeroCopy(spec, prompt)
}

const maxHeroSentence = 140

func applyHeroCopy(spec Spec, prompt string) Spec {
	subtitle := promptSentence(prompt)
	if subtitle == "" {
		return spec
	}
	for i, page := range spec.Pages {
		if page.Route != "/" {
			continue
		}
		for j, section := range page.Sections {
			if section.Type != "hero" {
				continue
			}
			spec.Pages[i].Sections[j].Subtitle = subtitle
			return spec
		}
	}
	return spec
}

func promptSentence(prompt string) string {
	text := strings.Join(strings.Fields(prompt), " ")
	if text == "" {
		return ""
	}
	end := len(text)
	for i, r := range text {
		if r == '.' || r == '!' || r == '?' {
			end = i + utf8.RuneLen(r)
			break
		}
	}
	sentence := text[:end]
	if utf8.RuneCountInString(sentence) <= maxHeroSentence {
		return sentence
	}
	runes := []rune(sentence)
	cut := maxHeroSentence
	for cut > 0 && runes[cut-1] != ' ' {
		cut--
	}
	if cut == 0 {
		cut = maxHeroSentence
	}
	return strings.TrimSpace(string(runes[:cut]))
}

func dropSection(spec Spec, sectionType string) Spec {
	for i, page := range spec.Pages {
		kept := make([]Section, 0, len(page.Sections))
		for _, section := range page.Sections {
			if section.Type != sectionType {
				kept = append(kept, section)
			}
		}
		spec.Pages[i].Sections = kept
	}
	return spec
}

func ensureSection(spec Spec, route, after string, section Section) Spec {
	if hasSection(spec, section.Type) {
		return spec
	}
	for i, page := range spec.Pages {
		if page.Route != route {
			continue
		}
		spec.Pages[i] = insertAfter(page, after, section)
		return spec
	}
	if len(spec.Pages) == 0 {
		return spec
	}
	spec.Pages[0] = insertAfter(spec.Pages[0], after, section)
	return spec
}

func ensureFAQ(spec Spec, name string) Spec {
	if hasSection(spec, "faq") {
		return spec
	}
	section := faqSection(name)
	for i, page := range spec.Pages {
		for _, existing := range page.Sections {
			if existing.Type == "pricing" {
				spec.Pages[i] = insertAfter(page, "pricing", section)
				return spec
			}
		}
	}
	return ensureSection(spec, "/", "features", section)
}

func insertAfter(page Page, after string, section Section) Page {
	out := make([]Section, 0, len(page.Sections)+1)
	inserted := false
	for _, existing := range page.Sections {
		out = append(out, existing)
		if !inserted && existing.Type == after {
			out = append(out, section)
			inserted = true
		}
	}
	if inserted {
		page.Sections = out
		return page
	}
	out = out[:0]
	for _, existing := range page.Sections {
		if existing.Type == "footer" {
			out = append(out, section)
		}
		out = append(out, existing)
	}
	if len(out) == len(page.Sections) {
		out = append(out, section)
	}
	page.Sections = out
	return page
}

func hasSection(spec Spec, sectionType string) bool {
	for _, page := range spec.Pages {
		for _, section := range page.Sections {
			if section.Type == sectionType {
				return true
			}
		}
	}
	return false
}

func testimonialSection(name string) Section {
	return Section{
		Type:  "testimonials",
		Title: "Notes on " + name,
		Items: []Item{
			{Title: "Ava Chen", Description: "The page was something people could click through, not a static mockup."},
			{Title: "Noah Patel", Description: "We could see the flow without leaving the preview."},
		},
	}
}

func faqSection(name string) Section {
	return Section{
		Type:  "faq",
		Title: "Questions",
		Items: []Item{
			{Title: "Where do I start?", Description: "Create an account and open the " + name + " workspace."},
			{Title: "Can I change this later?", Description: "Yes. The preview stays editable after you generate it."},
		},
	}
}

func statsSection() Section {
	return Section{
		Type:  "stats",
		Title: "At a glance",
		Items: []Item{
			{Title: "12", Description: "Active projects"},
			{Title: "4", Description: "Waiting for review"},
			{Title: "28", Description: "Previews this week"},
		},
	}
}

func navbar(brand string, links []Item) Section {
	return Section{Type: "navbar", Title: brand, Items: links}
}

func footer(name string) Section {
	return Section{Type: "footer", Title: name, Subtitle: "Generated preview for local development."}
}
