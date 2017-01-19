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

package controller

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"github.com/kubernetes-incubator/service-catalog/pkg/util"

	"github.com/ghodss/yaml"
)

const (
	// 'helm [--host <server>]' install <chart> --name <name> [--values <values file>]
	// Where parameters are key0-=value0,key1=value1
	createServiceFmt = "%s install %s --name %s %s"

	// 'helm [--host <server>]' upgrade <release> <chart> [--values <values file>]
	upgradeServiceFmt = "%s upgrade %s %s %s"

	// 'helm [--host <server>]' delete --purge %s'
	deleteServiceFmt = "%s delete --purge %s"

	// 'helm [--host <server>]' status name
	serviceStatusFmt = "%s status %s"
)

type helmReifier struct {
	client string
	server string
	cmd    string
}

// NewHelmReifier creates an instance of a Reifier interface which uses Helm
// as the underlying deployment implementation.
func NewHelmReifier(client string, server string) Reifier {
	cmd := client
	if len(server) > 0 {
		cmd = cmd + " --host " + server
	}
	return &helmReifier{client: client, server: server, cmd: cmd}
}

// Catalog returns all the available Services that can be instantiated
func (h *helmReifier) Catalog() ([]*brokerapi.Service, error) {
	return nil, errors.New("Implement me")
}

func shortenInstanceID(instanceID string) string {
	if len(instanceID) > 14 {
		instanceID = instanceID[0:13]
	}
	return instanceID
}

// CreateServiceInstance creates a new Service Instance
func (h *helmReifier) CreateServiceInstance(instanceID string, tmpl string, sir *brokerapi.CreateServiceInstanceRequest) (*brokerapi.CreateServiceInstanceResponse, error) {
	values := ""
	if len(sir.Parameters) > 0 {
		y, err := yaml.Marshal(sir.Parameters)
		if err != nil {
			glog.Errorf("Failed to marshal %#v : %v", sir.Parameters, err)
			return nil, err
		}
		glog.Infof("Marshalled into:\n%s\n", y)
		f, err := ioutil.TempFile("", "values-")
		if err != nil {
			glog.Errorf("Failed to create TempFile for values file: %v", err)
			return nil, err
		}
		defer os.Remove(f.Name())
		_, err = f.Write([]byte(y))
		if err != nil {
			return nil, err
		}
		values = "--values " + f.Name()
	}

	// If this instance already exists, just perform an upgrade.
	cmd := fmt.Sprintf(createServiceFmt, h.cmd, tmpl, shortenInstanceID(instanceID), values)
	if _, err := h.getStatus(instanceID); err == nil {
		cmd = fmt.Sprintf(upgradeServiceFmt, h.cmd, shortenInstanceID(instanceID), tmpl, values)
	}

	out, err := util.ExecCmd(cmd)
	if err != nil {
		return nil, err
	}
	notes := ParseNotes(out)
	glog.Infof("NOTES SECTION: '%s'\n", notes)

	return &brokerapi.CreateServiceInstanceResponse{}, nil
}

// RemoveServiceInstance removes an existing Service Instance
func (h *helmReifier) RemoveServiceInstance(instanceID string) error {
	cmd := fmt.Sprintf(deleteServiceFmt, h.cmd, shortenInstanceID(instanceID))
	out, err := util.ExecCmd(cmd)
	glog.Infof("Helm Delete Result:\n%s\n", out)
	return err
}

func (h *helmReifier) CreateServiceBinding(instanceID string, sir *brokerapi.BindingRequest) (*brokerapi.CreateServiceBindingResponse, error) {
	out, err := h.getStatus(instanceID)
	if err != nil {
		return nil, err
	}
	glog.Infof("GOT BACK: %s", out)
	notes := ParseNotes(out)

	var c brokerapi.Credential
	err = yaml.Unmarshal([]byte(notes), &c)

	return &brokerapi.CreateServiceBindingResponse{Credentials: c}, nil
}

func (h *helmReifier) RemoveServiceBinding(instanceID string) error {
	// TODO: Implement
	glog.Infof("Removing Service Binding: %s\n", instanceID)
	return nil
}

func (h *helmReifier) getStatus(instanceID string) (string, error) {
	cmd := fmt.Sprintf(serviceStatusFmt, h.cmd, shortenInstanceID(instanceID))
	return util.ExecCmd(cmd)
}

// ParseNotes will take the output of a Helm Install and return the NOTES.txt section out of it.
func ParseNotes(status string) string {
	scanner := bufio.NewScanner(strings.NewReader(status))
	ret := ""
	notesSection := false
	for scanner.Scan() {
		if notesSection == true {
			if ret != "" {
				ret = ret + "\n"
			}
			ret = ret + scanner.Text()
		}
		if scanner.Text() == "NOTES:" {
			notesSection = true
		}
	}
	return ret
}
