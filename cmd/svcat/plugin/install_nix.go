// +build !windows

package plugin

import "os"

func getUserHomeDir() string {
	return os.Getenv("HOME")
}

func getFileExt() string {
	return ""
}
