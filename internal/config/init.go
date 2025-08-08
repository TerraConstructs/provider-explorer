package config

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/terraconstructs/provider-explorer/internal/terraform"
)

func HasTerraformConfig(dir string) bool {
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(d.Name(), ".tf") ||
			(strings.HasSuffix(d.Name(), ".json") && strings.Contains(d.Name(), ".tf.")) {
			return fmt.Errorf("found") // Use error to break out of walk
		}

		return nil
	})

	return err != nil && err.Error() == "found"
}

func InitTerraformDirectory(dir string) error {
	tfInfo := terraform.FindTerraformBinary()

	fmt.Printf("Terraform configuration detected in %s\n", dir)
	fmt.Printf("This will run '%s init' to download providers.\n", tfInfo.Binary)
	fmt.Print("Continue? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		return fmt.Errorf("terraform init cancelled by user")
	}

	fmt.Printf("\nRunning %s init...\n", tfInfo.Binary)

	cmd := exec.Command(tfInfo.Binary, "init")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	fmt.Printf("\n%s init completed successfully!\n", tfInfo.Binary)
	return nil
}
