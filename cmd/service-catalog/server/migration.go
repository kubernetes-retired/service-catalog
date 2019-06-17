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
	"github.com/kubernetes-sigs/service-catalog/cmd/migration"
	"github.com/kubernetes-sigs/service-catalog/pkg/hyperkube"
)

// NewMigration creates a new hyperkube Server object that includes the
// description and flags of the Migration tool functionality.
func NewMigration() *hyperkube.Server {
	opts := migration.NewMigrationOptions()

	hks := hyperkube.Server{
		PrimaryName:     "migration",
		AlternativeName: "service-catalog-apiserver-to-crd-migration",
		SimpleUsage:     "migration",
		Long:            "The migration tool migrates Service Catalog resources from API Server (0.2.x) to CRDs (0.3.0) version",
		Run: func(_ *hyperkube.Server, args []string, stopCh <-chan struct{}) error {
			return migration.RunCommand(opts)
		},
		RespectsStopCh: false,
	}
	opts.AddFlags(hks.Flags())

	return &hks
}
