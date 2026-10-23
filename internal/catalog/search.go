package catalog

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

// Search queries the native package index and returns installable catalog entries.
// The result's Command(platform) is the exact package identifier selected by the user.
func Search(platform Platform, query string) ([]Package, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	var output []byte
	var err error
	switch platform {
	case Linux:
		switch {
		case commandExists("apt-cache"):
			output, err = exec.Command("apt-cache", "search", query).Output()
		case commandExists("dnf"):
			output, err = exec.Command("dnf", "search", query).Output()
		case commandExists("snap"):
			output, err = exec.Command("snap", "find", query).Output()
		default:
			return nil, fmt.Errorf("no supported Linux package search tool found (apt-cache, dnf, or snap)")
		}
	case Windows:
		if !commandExists("winget") {
			return nil, fmt.Errorf("winget is not available")
		}
		output, err = exec.Command("winget", "search", query, "--accept-source-agreements", "--disable-interactivity").Output()
	case MacOS:
		if !commandExists("brew") {
			return nil, fmt.Errorf("Homebrew is not available")
		}
		output, err = exec.Command("brew", "search", query).Output()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
	if err != nil {
		return nil, fmt.Errorf("package search failed: %w", err)
	}
	return parseResults(platform, string(output)), nil
}

func commandExists(name string) bool { _, err := exec.LookPath(name); return err == nil }

func parseResults(platform Platform, output string) []Package {
	switch platform {
	case Windows:
		return parseWingetResults(output)
	case MacOS:
		return parseNameResults(MacOS, output)
	default:
		return parseLinuxResults(platform, output)
	}
}

func parseLinuxResults(platform Platform, output string) []Package {
	results := make([]Package, 0)
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Last metadata expiration") {
			continue
		}
		name, description := line, "Package from native index"
		if parts := strings.SplitN(line, " - ", 2); len(parts) == 2 {
			name, description = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
		if name == "" || strings.ContainsAny(name, " \t") {
			continue
		}
		results = append(results, Package{Name: name, Description: description, Supported: true, Commands: map[Platform]string{platform: name}})
	}
	return results
}

func parseNameResults(platform Platform, output string) []Package {
	results := make([]Package, 0)
	for _, raw := range strings.Split(output, "\n") {
		name := strings.TrimSpace(raw)
		if name == "" || strings.HasPrefix(name, "==>") {
			continue
		}
		results = append(results, Package{Name: name, Description: "Native package index", Supported: true, Commands: map[Platform]string{platform: name}})
	}
	return results
}

func parseWingetResults(output string) []Package {
	results := make([]Package, 0)
	data := false
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "---") {
			data = true
			continue
		}
		if !data {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		idIndex := -1
		for i, field := range fields {
			if strings.Contains(field, ".") && !strings.HasPrefix(field, "...") {
				idIndex = i
				break
			}
		}
		if idIndex <= 0 {
			continue
		}
		name := strings.Join(fields[:idIndex], " ")
		id := fields[idIndex]
		results = append(results, Package{Name: name, Description: "WinGet · " + id, Supported: true, Commands: map[Platform]string{Windows: id}})
	}
	return results
}
