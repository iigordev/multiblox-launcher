package manager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CreateInstance handles the full lifecycle of creating a new Roblox clone.
func CreateInstance(originalPath, targetDir, name, iconName string) error {
	instancePath := filepath.Join(targetDir, name+".app")

	// 1. Check if an instance with this name already exists
	if _, err := os.Stat(instancePath); !os.IsNotExist(err) {
		return fmt.Errorf("instance '%s' already exists", name)
	}

	// 2. Use native macOS 'cp' to clone.
	// -R: Recursive
	// -p: Preserve permissions and attributes (Crucial for the binary)
	cmd := exec.Command("cp", "-Rp", originalPath, instancePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone bundle using cp: %w", err)
	}

	return nil
}

// DeleteInstance removes the .app bundle from the disk.
func DeleteInstance(instancePath string) error {
	if instancePath == "/" || instancePath == "/Applications" {
		return fmt.Errorf("safety trigger: cannot delete root directories")
	}
	return os.RemoveAll(instancePath)
}

// PatchInstance replaces an old clone with a fresh one while keeping its identity
func PatchInstance(originalPath, instancePath, name, iconName string) error {
	if err := os.RemoveAll(instancePath); err != nil {
		return fmt.Errorf("failed to remove old instance: %w", err)
	}

	targetDir := filepath.Dir(instancePath)
	return CreateInstance(originalPath, targetDir, name, iconName)
}
