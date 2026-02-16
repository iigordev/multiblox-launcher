package utils

import (
	"os"
	"os/exec"

	"howett.net/plist"
)

// UpdateBundleMetadata changes the identity of the app so they don't conflict
func UpdateBundleMetadata(appPath string, newName string, newBundleID string) error {
	plistPath := appPath + "/Contents/Info.plist"

	// 1. Open and Decode the Plist
	file, err := os.Open(plistPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var data map[string]interface{}
	decoder := plist.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	// 2. Modify Identity Keys
	data["CFBundleIdentifier"] = newBundleID
	data["CFBundleName"] = newName
	data["CFBundleDisplayName"] = newName

	// 3. Encode and Save back to file
	outFile, err := os.Create(plistPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	encoder := plist.NewEncoder(outFile)
	encoder.Indent("    ")
	return encoder.Encode(data)
}

// RegisterWithOS forces macOS to refresh its app database
func RegisterWithOS(appPath string) {
	exec.Command("/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister", "-f", appPath).Run()
}
