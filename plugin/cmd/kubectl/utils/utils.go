package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

type namespace struct {
	Metadata metadata `json:"metadata"`
	Code     int      `json:"code"`
}

type metadata struct {
	Name string `json:"name"`
}

func CheckNamespaceExists(name string) error {
	proxyURL := "http://127.0.0.1"
	proxyPort := "8881"

	kubeProxy := exec.Command("kubectl", "proxy", "-p", proxyPort)
	defer func() {
		if err := kubeProxy.Process.Kill(); err != nil {
			Exit1(fmt.Sprintf("failed to kill kubectl proxy (%s)", err))
		}
	}()

	err := kubeProxy.Start()
	if err != nil {
		return fmt.Errorf("Cannot start kubectl proxy (%s)", err)
	}

	resp, err := http.Get(fmt.Sprintf("%s:%s/api/v1/namespaces/%s", proxyURL, proxyPort, name))
	if err != nil {
		fmt.Errorf("Error looking up namespace from core api server (%s)", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error retrieving core api server response body during namespace lookup (%s)", err)
	}

	ns := namespace{}
	err = json.Unmarshal(body, &ns)
	if err != nil {
		return fmt.Errorf("Error parsing core api server response body during namespace lookup (%s)", err)
	}

	if ns.Code == 404 || ns.Metadata.Name == "" {
		return fmt.Errorf("Namespace not found")
	}

	return nil
}

func SCUrlEnv() string {
	url := os.Getenv("SERVICE_CATALOG_URL")
	if url == "" {
		return ""
	}
	return url
}

func Exit1(errStr string) {
	Error(errStr)
	os.Exit(1)
}
