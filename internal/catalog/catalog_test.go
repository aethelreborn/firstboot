package catalog

import "testing"

func TestForIncludesCoreAndAIChoices(t *testing.T) {
	packages := For(Linux)
	want := map[string]bool{"Git": false, "Python": false, "Rust": false, "Claude Code": false, "OpenCode": false, "Hermes": false, "OpenClaw": false}
	for _, p := range packages {
		if _, ok := want[p.Name]; ok {
			want[p.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("catalog missing %q", name)
		}
	}
}

func TestEveryPlatformHasCoreCommands(t *testing.T) {
	for _, platform := range []Platform{Linux, MacOS, Windows} {
		for _, p := range For(platform)[:7] {
			if p.Command(platform) == "" {
				t.Errorf("%s has no command for %s", p.Name, platform)
			}
		}
	}
}
