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
	"io"
	"io/ioutil"
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
func KubeGet(endpoint string) (stdout, stderr io.ReadCloser, err error) {
	v, logLevel := Loglevel()
	caller := os.Getenv("KUBECTL_PLUGINS_CALLER")
	kubeCmd := exec.Command(caller, "--namespace", Namespace(), v, logLevel, "get", "--raw", "/api/v1/"+endpoint)

	stdout, err = kubeCmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	stderr, err = kubeCmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	err = kubeCmd.Start()
	if err != nil {
		return nil, nil, err
	}

	return stdout, stderr, nil
}

// CheckNamespaceExists will query our kube apiserver to see if the
// specified namespace exists - if not it returns an error
func CheckNamespaceExists(name string) error {
	stdout, stderr, err := KubeGet("namespaces/" + name)
	if err != nil {
		return fmt.Errorf("error: unable to perform namespace lookup: %v", err)
	}
	defer stdout.Close()
	defer stderr.Close()

	result, err := ioutil.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("error: unable to read standard output while performing a namespace lookup: %v", err)
	}

	errOutput, err := ioutil.ReadAll(stderr)
	if err == nil && len(errOutput) > 0 {
		fmt.Fprintf(os.Stderr, "%v\n", string(errOutput))
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

// Namespace will return the value of KUBECTL_PLUGINS_CURRENT_NAMESPACE env var
func Namespace() string {
	return os.Getenv("KUBECTL_PLUGINS_CURRENT_NAMESPACE")
}

func Loglevel() (flagName, flagValue string) {
	kubeLoglevel := os.Getenv("KUBECTL_PLUGINS_GLOBAL_FLAG_V")
	otherLoglevel := os.Getenv("KUBECTL_PLUGINS_GLOBAL_FLAG_LOGLEVEL")
	if len(otherLoglevel) > 0 {
		return "--loglevel", otherLoglevel
	}
	if len(kubeLoglevel) == 0 {
		kubeLoglevel = "0"
	}
	return "--v", kubeLoglevel
}

// Exit1 will print the specified error string to the screen and
// then stop the program, with an exit code of 1
func Exit1(errStr string) {
	Error(errStr)
	os.Exit(1)
}
