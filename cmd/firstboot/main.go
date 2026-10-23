package main

import (
	"fmt"
	"os"

	"github.com/aethelreborn/firstboot/internal/catalog"
	"github.com/aethelreborn/firstboot/internal/detect"
	"github.com/aethelreborn/firstboot/internal/tui"
)

func main() {
	platform := detect.Platform()
	packages := catalog.For(platform)
	for i := range packages {
		packages[i].Installed = detect.Available(packages[i].Detect)
		packages[i].Supported = packages[i].Command(platform) != ""
	}
	result, err := tui.Run(platform, packages)
	if err != nil {
		fmt.Fprintln(os.Stderr, "firstboot:", err)
		os.Exit(1)
	}
	if len(result) == 0 {
		return
	}
	if err := tui.RunInstall(platform, result); err != nil {
		fmt.Fprintln(os.Stderr, "firstboot:", err)
		os.Exit(1)
	}
}
