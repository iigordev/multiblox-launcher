package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"multi-blox/internal/manager"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"howett.net/plist"
)

// This directive bakes the icons into your MultiBlox binary so they are always available to "send"
//
//go:embed all:frontend/public/assets/images/instanceIcons
var iconFiles embed.FS

// InstanceConfig handles the "Source of Truth"
type InstanceConfig struct {
	Name       string `json:"name"`
	IconFolder string `json:"iconFolder"`
}

type Instance struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Status     string `json:"status"`
	IconFolder string `json:"iconFolder"`
	Version    string `json:"version"`
}

type App struct {
	ctx          context.Context
	baseDir      string
	instancesDir string
	configsDir   string
}

func NewApp() *App {
	base := "/Applications/MultiBlox"
	return &App{
		baseDir:      base,
		instancesDir: filepath.Join(base, "Instances"),
		configsDir:   filepath.Join(base, "Configs"),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	err := os.MkdirAll(a.instancesDir, 0755)
	if err != nil {
		fmt.Printf("Startup Error: %v\n", err)
	}
	os.MkdirAll(a.configsDir, 0755)
	go a.versionMonitor()
}

func (a *App) RequestPermissions() {
	exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_AllFiles").Run()
}

func (a *App) loadConfig(name string) (InstanceConfig, error) {
	path := filepath.Join(a.configsDir, name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return InstanceConfig{}, err
	}
	var config InstanceConfig
	err = json.Unmarshal(data, &config)
	return config, err
}

func (a *App) getPlistValue(path string, key string) string {
	plistPath := filepath.Join(path, "Contents", "Info.plist")
	file, err := os.Open(plistPath)
	if err != nil {
		return "Unknown"
	}
	defer file.Close()

	var data map[string]interface{}
	decoder := plist.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return "Unknown"
	}

	if val, ok := data[key].(string); ok {
		return val
	}
	return "Unknown"
}

func (a *App) versionMonitor() {
	masterPath := "/Applications/Roblox.app"
	lastVersion := a.GetAppVersion(masterPath)
	ticker := time.NewTicker(60 * time.Second)

	for {
		select {
		case <-ticker.C:
			currentVersion := a.GetAppVersion(masterPath)
			if currentVersion != lastVersion && currentVersion != "Unknown" {
				lastVersion = currentVersion
				runtime.EventsEmit(a.ctx, "roblox_updated", currentVersion)
			}
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *App) GetAppVersion(path string) string {
	return a.getPlistValue(path, "CFBundleShortVersionString")
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s failed: %v\n%s", name, err, string(out))
	}
	return nil
}

func (a *App) applySecurityFixes(appPath string) error {
	runCmd("xattr", "-cr", appPath)
	runCmd("codesign", "--remove-signature", appPath)
	return runCmd("codesign", "--force", "--deep", "--sign", "-", appPath)
}

func (a *App) saveConfig(name string, iconFolder string) {
	config := InstanceConfig{Name: name, IconFolder: iconFolder}
	path := filepath.Join(a.configsDir, name+".json")
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(path, data, 0644)
}

func (a *App) CreateNewInstance(name string, iconName string) string {
	masterApp := "/Applications/Roblox.app"
	instanceAppPath := filepath.Join(a.instancesDir, name+".app")

	// FORCE CAPITAL CASING: Ensures "icon5" -> "Icon5"
	digit := strings.TrimLeft(strings.ToLower(iconName), "icon")
	formattedIconName := "Icon" + digit

	if _, err := os.Stat(masterApp); os.IsNotExist(err) {
		return "Error: Roblox.app not found"
	}

	// 1. Clone the app via your internal manager
	err := manager.CreateInstance(masterApp, a.instancesDir, name, formattedIconName)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// 2. Setup Paths
	plistPath := filepath.Join(instanceAppPath, "Contents", "Info.plist")
	resPath := filepath.Join(instanceAppPath, "Contents", "Resources")
	os.MkdirAll(resPath, 0755)

	// 3. Update Plist - UNIQUE ID prevents macOS from redirecting to the main Roblox.app
	newID := "com.multiblox." + strings.ReplaceAll(name, " ", "")
	runCmd("/usr/libexec/PlistBuddy", "-c", "Set :CFBundleIdentifier "+newID, plistPath)
	runCmd("/usr/libexec/PlistBuddy", "-c", "Set :CFBundleName "+name, plistPath)
	runCmd("/usr/libexec/PlistBuddy", "-c", "Set :CFBundleIconFile "+formattedIconName, plistPath)

	// Remove Roblox binary cache to force refresh
	os.Remove(filepath.Join(resPath, "Assets.car"))

	// 4. EXTRACT EMBEDDED ICON
	targetFile := formattedIconName + ".icns"
	embedPath := "frontend/public/assets/images/instanceIcons/" + targetFile
	iconData, err := iconFiles.ReadFile(embedPath)

	if err == nil {
		// Clean out old Roblox icons
		os.Remove(filepath.Join(resPath, "RobloxApp.icns"))
		os.Remove(filepath.Join(resPath, "AppIcon.icns"))

		// Write the new icon file to the instance bundle
		dest := filepath.Join(resPath, targetFile)
		os.WriteFile(dest, iconData, 0644)
	} else {
		fmt.Printf("EMBED FAILURE: %v\n", err)
	}

	// 5. Security & LaunchServices Registration
	a.applySecurityFixes(instanceAppPath)
	a.saveConfig(name, iconName)

	// Force macOS to recognize this as a unique app
	runCmd("touch", instanceAppPath)
	runCmd("/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister", "-f", "-v", instanceAppPath)

	return "Success"
}

func (a *App) RepairInstance(name string) string {
	config, err := a.loadConfig(name)
	if err != nil {
		return "Error: Could not find config for " + name
	}
	targetPath := filepath.Join(a.instancesDir, name+".app")
	a.StopInstance(name)
	time.Sleep(500 * time.Millisecond)
	os.RemoveAll(targetPath)
	return a.CreateNewInstance(config.Name, config.IconFolder)
}

func (a *App) Launch(name string) {
	var path string
	if name == "Roblox" {
		path = "/Applications/Roblox.app"
	} else {
		path = filepath.Join(a.instancesDir, name+".app")
	}
	// Using direct path without "-a" ensures we launch THIS specific instance
	exec.Command("open", "-n", path).Run()
}

func (a *App) StopInstance(name string) string {
	runCmd("pkill", "-9", "-f", name+".app")
	return "Stopped"
}

func (a *App) DeleteInstance(name string) string {
	a.StopInstance(name)
	time.Sleep(500 * time.Millisecond)
	appPath := filepath.Join(a.instancesDir, name+".app")
	runCmd("chflags", "-R", "nouchg", appPath)
	os.RemoveAll(appPath)
	os.Remove(filepath.Join(a.configsDir, name+".json"))
	return "Deleted"
}

func (a *App) GetInstances() ([]Instance, error) {
	var instances []Instance
	masterVersion := a.GetAppVersion("/Applications/Roblox.app")
	files, _ := os.ReadDir(a.configsDir)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			name := strings.TrimSuffix(file.Name(), ".json")
			config, err := a.loadConfig(name)
			if err != nil {
				continue
			}
			appPath := filepath.Join(a.instancesDir, name+".app")
			version := a.GetAppVersion(appPath)
			status := "Operational"
			if version != masterVersion && version != "Unknown" {
				status = "Outdated"
			}
			instances = append(instances, Instance{
				Name:       config.Name,
				Path:       appPath,
				Status:     status,
				IconFolder: config.IconFolder,
				Version:    version,
			})
		}
	}
	return instances, nil
}

func (a *App) UpdateInstance(oldName string, newName string, iconName string) string {
	oldAppPath := filepath.Join(a.instancesDir, oldName+".app")
	newAppPath := filepath.Join(a.instancesDir, newName+".app")
	oldConfigPath := filepath.Join(a.configsDir, oldName+".json")
	newConfigPath := filepath.Join(a.configsDir, newName+".json")

	if oldName != newName {
		os.Rename(oldAppPath, newAppPath)
		os.Rename(oldConfigPath, newConfigPath)
	}
	a.saveConfig(newName, iconName)
	a.applySecurityFixes(newAppPath)
	return "Success"
}

func (a *App) OpenBrowser(url string) {
	runtime.BrowserOpenURL(a.ctx, url)
}
