// +build windows
package plugin

import "os"

func getUserHomeDir() string {
	return os.Getenv("USERPROFILE")
}

func getFileExt() string {
	return ".exe"
}
