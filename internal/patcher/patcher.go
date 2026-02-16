package patcher

import (
	"os"

	"howett.net/plist"
)

// GetVersion extracts the CFBundleShortVersionString from a bundle
func GetVersion(appPath string) (string, error) {
	plistPath := appPath + "/Contents/Info.plist"

	file, err := os.Open(plistPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var data map[string]interface{}
	decoder := plist.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return "", err
	}

	version, ok := data["CFBundleShortVersionString"].(string)
	if !ok {
		return "0.0.0.0", nil
	}
	return version, nil
}

// NeedsPatch returns true if the clone version doesn't match the master
func NeedsPatch(masterApp, cloneApp string) bool {
	vMaster, errM := GetVersion(masterApp)
	vClone, errC := GetVersion(cloneApp)

	if errM != nil || errC != nil {
		return true // If we can't tell, assume it needs a patch for safety
	}

	return vMaster != vClone
}
