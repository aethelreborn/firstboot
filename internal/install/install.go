package install

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/aethelreborn/firstboot/internal/catalog"
)

type Progress struct {
	Package catalog.Package
	Index   int
	Total   int
	Error   error
}

type Failure struct {
	Package catalog.Package
	Error   error
}

// NeedsPassword reports whether Linux will need a sudo password before installs.
func NeedsPassword(platform catalog.Platform) bool {
	if platform != catalog.Linux || runtime.GOOS == "windows" {
		return false
	}
	if _, err := exec.LookPath("sudo"); err != nil {
		return false
	}
	return exec.Command("sudo", "-n", "true").Run() != nil
}

// RunWithProgress installs every selected package while keeping package-manager
// stdout/stderr out of the TUI. Failures are collected so the UI can present them.
func RunWithProgress(platform catalog.Platform, packages []catalog.Package, password string, report func(Progress)) []Failure {
	failures := make([]Failure, 0)
	for i, p := range packages {
		report(Progress{Package: p, Index: i, Total: len(packages)})
		args, executable := command(platform, p.Command(platform))
		if executable == "" {
			err := fmt.Errorf("no supported package manager found")
			failures = append(failures, Failure{Package: p, Error: err})
			report(Progress{Package: p, Index: i, Total: len(packages), Error: err})
			continue
		}
		if executable == "sudo" {
			args = append([]string{"-S", "-p", ""}, args...)
		}
		cmd := exec.Command(executable, args...)
		if executable == "sudo" {
			cmd.Stdin = strings.NewReader(password + "\n")
		}
		var output bytes.Buffer
		cmd.Stdout, cmd.Stderr = &output, &output
		if err := cmd.Run(); err != nil {
			message := strings.TrimSpace(output.String())
			if message == "" {
				message = err.Error()
			}
			failure := Failure{Package: p, Error: fmt.Errorf("%s: %s", err, message)}
			failures = append(failures, failure)
			report(Progress{Package: p, Index: i, Total: len(packages), Error: failure.Error})
		}
	}
	return failures
}

// Run remains available for scripts and callers that do not need progress UI.
func Run(platform catalog.Platform, packages []catalog.Package) error {
	failures := RunWithProgress(platform, packages, "", func(Progress) {})
	if len(failures) > 0 {
		return failures[0].Error
	}
	return nil
}

func command(platform catalog.Platform, pkg string) ([]string, string) {
	switch platform {
	case catalog.Linux:
		if _, err := exec.LookPath("apt-get"); err == nil {
			return []string{"apt-get", "install", "-y", pkg}, "sudo"
		}
		if _, err := exec.LookPath("dnf"); err == nil {
			return []string{"dnf", "install", "-y", pkg}, "sudo"
		}
	case catalog.MacOS:
		return []string{"install", pkg}, "brew"
	case catalog.Windows:
		return []string{"install", "--id", pkg, "--exact", "--source", "winget", "--accept-source-agreements", "--accept-package-agreements"}, "winget"
	}
	return nil, ""
}
