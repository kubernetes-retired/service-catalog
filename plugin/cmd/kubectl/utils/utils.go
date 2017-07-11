/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

// CheckNamespaceExists will query our kube apiserver to see if the
// specified namespace exists - if not it returns an error
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
		return fmt.Errorf("Error looking up namespace from core api server (%s)", err)
	}

	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()
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

// SCUrlEnv will return the value of the SERVICE_CATALOG_URL env var
func SCUrlEnv() string {
	url := os.Getenv("SERVICE_CATALOG_URL")
	if url == "" {
		return ""
	}
	return url
}

// Exit1 will print the specified error string to the screen and
// then stop the program, with an exit code of 1
func Exit1(errStr string) {
	Error(errStr)
	os.Exit(1)
}
