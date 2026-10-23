package catalog

import "runtime"

type Platform string

const (
	Linux   Platform = "linux"
	MacOS   Platform = "darwin"
	Windows Platform = "windows"
)

type Package struct {
	Name        string
	Description string
	Detect      string
	Installed   bool
	Supported   bool
	Selected    bool
	Commands    map[Platform]string
}

func (p Package) Command(platform Platform) string { return p.Commands[platform] }

func For(platform Platform) []Package {
	common := []Package{
		{Name: "Git", Description: "Version control", Detect: "git", Commands: map[Platform]string{Linux: "git", MacOS: "git", Windows: "Git.Git"}},
		{Name: "Docker", Description: "Containers and images", Detect: "docker", Commands: map[Platform]string{Linux: "docker.io", MacOS: "docker", Windows: "Docker.DockerDesktop"}},
		{Name: "Node.js", Description: "JavaScript runtime", Detect: "node", Commands: map[Platform]string{Linux: "nodejs", MacOS: "node", Windows: "OpenJS.NodeJS.LTS"}},
		{Name: "Python", Description: "Python 3 runtime", Detect: "python3", Commands: map[Platform]string{Linux: "python3", MacOS: "python", Windows: "Python.Python.3.12"}},
		{Name: "JDK", Description: "Java development kit", Detect: "java", Commands: map[Platform]string{Linux: "default-jdk", MacOS: "openjdk", Windows: "Microsoft.OpenJDK.21"}},
		{Name: "Rust", Description: "Rust toolchain", Detect: "rustc", Commands: map[Platform]string{Linux: "rustc", MacOS: "rust", Windows: "Rustlang.Rustup"}},
		{Name: "Go", Description: "Go toolchain", Detect: "go", Commands: map[Platform]string{Linux: "golang", MacOS: "go", Windows: "GoLang.Go"}},
	}
	ai := []Package{
		{Name: "Claude Code", Description: "Anthropic coding agent", Detect: "claude", Commands: map[Platform]string{Linux: "anthropic-ai/claude-code", MacOS: "anthropic-ai/claude-code", Windows: "anthropic-ai/claude-code"}},
		{Name: "OpenCode", Description: "Open-source coding agent", Detect: "opencode", Commands: map[Platform]string{Linux: "opencode", MacOS: "opencode", Windows: "opencode"}},
		{Name: "Hermes", Description: "AI agent CLI", Detect: "hermes", Commands: map[Platform]string{Linux: "hermes", MacOS: "hermes", Windows: "hermes"}},
		{Name: "OpenClaw", Description: "AI automation CLI", Detect: "openclaw", Commands: map[Platform]string{Linux: "openclaw", MacOS: "openclaw", Windows: "openclaw"}},
	}
	return append(common, ai...)
}

func Current() Platform {
	switch runtime.GOOS {
	case "darwin":
		return MacOS
	case "windows":
		return Windows
	default:
		return Linux
	}
}
