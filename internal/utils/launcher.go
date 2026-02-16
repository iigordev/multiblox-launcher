package utils

import (
	"fmt"
	"os"
	"os/exec"
)

// SetInstanceIcon replaces the .icns file inside the app bundle
func SetInstanceIcon(appPath string, sourceIconPath string) error {
	destIconPath := appPath + "/Contents/Resources/AppIcon.icns"

	// Read the new icon
	iconData, err := os.ReadFile(sourceIconPath)
	if err != nil {
		return fmt.Errorf("could not read source icon: %w", err)
	}

	// Overwrite the original icon inside the clone
	err = os.WriteFile(destIconPath, iconData, 0644)
	if err != nil {
		return fmt.Errorf("could not write icon to bundle: %w", err)
	}

	return nil
}

// LaunchInstance uses the macOS 'open' command to start the clone
func LaunchInstance(appPath string) error {
	// -n opens a new instance even if one is already running
	cmd := exec.Command("open", "-n", appPath)
	return cmd.Run()
}
