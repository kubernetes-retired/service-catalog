/*
Copyright 2019 The Kubernetes Authors.

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

package server

import (
	"github.com/kubernetes-incubator/service-catalog/cmd/cleaner/server"
	"github.com/kubernetes-incubator/service-catalog/pkg/hyperkube"
)

// NewCleaner creates a new hyperkube Server object that includes the
// description and flags of the Cleaner functionality.
func NewCleaner() *hyperkube.Server {
	opts := server.NewCleanerOptions()

	hks := hyperkube.Server{
		PrimaryName:     "cleaner",
		AlternativeName: "service-catalog-cleaner",
		SimpleUsage:     "cleaner",
		Long:            "The cleaner asserts all CRD will be removed before removing helm release, it also removes all finalizers from CRs",
		Run: func(_ *hyperkube.Server, args []string, stopCh <-chan struct{}) error {
			return server.RunCommand(opts)
		},
		RespectsStopCh: false,
	}
	opts.AddFlags(hks.Flags())

	return &hks
}
