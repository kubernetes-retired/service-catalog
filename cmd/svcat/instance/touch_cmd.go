/*
Copyright 2018 The Kubernetes Authors.

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

package instance

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type touchInstanceCmd struct {
	*command.Namespaced
	name string
}

// NewTouchCommand builds a "svcat touch instance" command.
func NewTouchCommand(cxt *command.Context) *cobra.Command {
	touchInstanceCmd := &touchInstanceCmd{Namespaced: command.NewNamespaced(cxt)}
	cmd := &cobra.Command{
		Use:   "instance",
		Short: "Touch an instance to make service-catalog try to process the spec again",
		Long: `Touch instance will increment the updateRequests field on the instance. 
Then, service catalog will process the instance's spec again. It might do an update, a delete, or 
nothing.`,
		Example: command.NormalizeExamples(`svcat touch instance wordpress-mysql-instance --namespace mynamespace`),
		PreRunE: command.PreRunE(touchInstanceCmd),
		RunE:    command.RunE(touchInstanceCmd),
	}
	touchInstanceCmd.AddNamespaceFlags(cmd.Flags(), false)

	return cmd
}

func (c *touchInstanceCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("an instance name is required")
	}
	c.name = args[0]

	return nil
}

func (c *touchInstanceCmd) Run() error {
	instance, err := c.App.RetrieveInstance(c.Namespace, c.name)
	if err != nil {
		return nil
	}

	planName := instance.Spec.PlanReference.ClusterServicePlanExternalName
	params := servicecatalog.BuildParametersFromInstance(instance.Spec.Parameters)
	secrets := servicecatalog.BuildMapFromInstanceSecretRefs(instance.Spec.ParametersFrom)

	const retries = 3
	for j := 0; j < retries; j++ {
		c.App.UpdateInstance(c.Namespace, c.name, planName, params, secrets)
		if err == nil {
			return nil
		}
		// if we didn't get a conflict, no idea what happened
		if !apierrors.IsConflict(err) {
			return fmt.Errorf("could not touch instance (%s)", err)
		}
	}

	// conflict after `retries` tries
	return fmt.Errorf("could not sync service broker after %d tries", retries)
}
