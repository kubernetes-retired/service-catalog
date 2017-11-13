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

// KubeGet makes a --raw GET call to the specified core api server endpoint.
// The "endpoint" parameter must begin without a slash and contain both a
// <resource_kind_plural>/<resource_name>: namespaces/foo
func KubeGet(endpoint string) ([]byte, error) {
	caller := os.Getenv("KUBECTL_PLUGINS_CALLER")
	kubeCmd := exec.Command(caller, "get", "--raw", "/api/v1/"+endpoint)
	return kubeCmd.Output()
}

// CheckNamespaceExists will query our kube apiserver to see if the
// specified namespace exists - if not it returns an error
func CheckNamespaceExists(name string) error {
	result, err := KubeGet("namespaces/" + name)
	if err != nil {
		return fmt.Errorf("error: unable to perform namespace lookup: %v", err)
	}

	ns := namespace{}
	err = json.Unmarshal(result, &ns)
	if err != nil {
		return fmt.Errorf("error: unable to parse core api server response during namespace lookup: %v", err)
	}

	if ns.Code == 404 || len(ns.Metadata.Name) == 0 {
		return fmt.Errorf("error: namespace %q does not exist", name)
	}

	return nil
}

// SCUrlEnv will return the value of the SERVICE_CATALOG_URL env var
func SCUrlEnv() string {
	return os.Getenv("SERVICE_CATALOG_URL")
}

// Exit1 will print the specified error string to the screen and
// then stop the program, with an exit code of 1
func Exit1(errStr string) {
	Error(errStr)
	os.Exit(1)
}
