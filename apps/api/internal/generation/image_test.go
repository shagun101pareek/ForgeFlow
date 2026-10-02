package generation

import (
	"strings"
	"testing"
)

func TestApplyProjectImageUsesTheFirstHero(t *testing.T) {
	spec := Spec{Pages: []Page{{
		Route: "/",
		Sections: []Section{
			{Type: "navbar", Title: "Notes"},
			{Type: "hero", Title: "Home"},
			{Type: "hero", Title: "Second"},
		},
	}}}
	next := applyProjectImage(spec, "data:image/png;base64,aaaa")
	if next.Pages[0].Sections[1].Image != "data:image/png;base64,aaaa" {
		t.Fatalf("first hero = %q", next.Pages[0].Sections[1].Image)
	}
	if next.Pages[0].Sections[2].Image != "" {
		t.Fatal("second hero should stay empty")
	}
	files := GenerateFiles(next)
	if !strings.Contains(files[0].Code, `src={"data:image/png;base64,aaaa"}`) && !strings.Contains(files[0].Code, "data:image/png;base64,aaaa") {
		t.Fatal("hero image was not rendered")
	}
	if strings.Contains(files[0].Code, "second hero") {
		t.Fatal("unexpected marker")
	}
	cleared := stripImages(next)
	if cleared.Pages[0].Sections[1].Image != "" {
		t.Fatal("image was not stripped")
	}
}

func TestDetectImageType(t *testing.T) {
	if DetectImageType([]byte{0xff, 0xd8, 0xff, 0x00}) != "image/jpeg" {
		t.Fatal("jpeg")
	}
	if DetectImageType([]byte("GIF89a")) != "image/gif" {
		t.Fatal("gif")
	}
	if DetectImageType([]byte("not an image")) != "" {
		t.Fatal("unknown should be rejected")
	}
}
