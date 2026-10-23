package catalog

import "testing"

func TestParseLinuxResults(t *testing.T) {
	results := parseLinuxResults(Linux, "code - Visual Studio Code\nvlc - multimedia player\n")
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Name != "code" || results[0].Command(Linux) != "code" {
		t.Fatalf("unexpected first result: %+v", results[0])
	}
}

func TestParseWingetResults(t *testing.T) {
	output := `
Name                 Id                         Version Source
---------------------------------------------------------------
Visual Studio Code   Microsoft.VisualStudioCode 1.99.0  winget
Discord              Discord.Discord            1.0.0    winget
Android Studio       Google.AndroidStudio       2025.1   winget
`
	results := parseWingetResults(output)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	if results[0].Name != "Visual Studio Code" {
		t.Fatalf("unexpected name: %q", results[0].Name)
	}
	if results[0].Command(Windows) != "Microsoft.VisualStudioCode" {
		t.Fatalf("unexpected id: %q", results[0].Command(Windows))
	}
	if results[2].Command(Windows) != "Google.AndroidStudio" {
		t.Fatalf("unexpected Android Studio id: %q", results[2].Command(Windows))
	}
}
