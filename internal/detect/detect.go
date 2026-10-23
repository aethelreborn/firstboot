package detect

import (
	"github.com/aethelreborn/firstboot/internal/catalog"
	"os/exec"
)

func Platform() catalog.Platform    { return catalog.Current() }
func Available(command string) bool { _, err := exec.LookPath(command); return err == nil }
