package terraform

import (
	"os/exec"
)

type TerraformInfo struct {
	Binary   string
	Registry string
}

func FindTerraformBinary() TerraformInfo {
	if _, err := exec.LookPath("tofu"); err == nil {
		return TerraformInfo{
			Binary:   "tofu",
			Registry: "registry.opentofu.org",
		}
	}
	
	return TerraformInfo{
		Binary:   "terraform",
		Registry: "registry.terraform.io",
	}
}