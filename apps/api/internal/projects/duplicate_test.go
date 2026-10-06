package projects

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestDuplicateName(t *testing.T) {
	if duplicateName("Northwind") != "Northwind copy" {
		t.Fatalf("name = %s", duplicateName("Northwind"))
	}
	copied := duplicateName(strings.Repeat("a", 120))
	if utf8.RuneCountInString(copied) != 120 || !strings.HasSuffix(copied, " copy") {
		t.Fatalf("long name length %d", utf8.RuneCountInString(copied))
	}
}
