package install

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/aethelreborn/firstboot/internal/catalog"
)

func Run(platform catalog.Platform, packages []catalog.Package) error {
	for _, p := range packages {
		fmt.Printf("Installing %s...\n", p.Name)
		args, name := command(platform, p.Command(platform))
		if name == "" {
			return fmt.Errorf("no supported package manager found for %s", p.Name)
		}
		cmd := exec.Command(name, args...)
		cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("install %s: %w", p.Name, err)
		}
	}
	fmt.Println("Done. Review any package-manager messages above.")
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
