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

package main

import (
	"flag"
	"fmt"
	"os"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/pluginutils"

	_ "github.com/golang/glog" // Initialize glog flags
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/binding"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/broker"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/class"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/completion"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/instance"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/plan"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/plugin"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/versions"
	svcatclient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/kube"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	k8sclient "k8s.io/client-go/kubernetes"
)

// These are build-time values, set during an official release
var (
	commit  string
	version string
)

func main() {
	// root command context
	cxt := &command.Context{
		Viper: viper.New(),
	}
	cmd := buildRootCommand(cxt)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func buildRootCommand(cxt *command.Context) *cobra.Command {
	// Make cobra aware of select glog flags
	// Enabling all flags causes unwanted deprecation warnings from glog to always print in plugin mode
	pflag.CommandLine.AddGoFlag(flag.CommandLine.Lookup("v"))
	pflag.CommandLine.AddGoFlag(flag.CommandLine.Lookup("logtostderr"))
	pflag.CommandLine.Set("logtostderr", "true")

	// root command flags
	var opts struct {
		KubeConfig  string
		KubeContext string
	}

	cmd := &cobra.Command{
		Use:          "svcat",
		Short:        "The Kubernetes Service Catalog Command-Line Interface (CLI)",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Enable tests to swap the output
			if cxt.Output == nil {
				cxt.Output = cmd.OutOrStdout()
			}

			// Initialize flags from kubectl plugin environment variables
			if plugin.IsPlugin() {
				plugin.BindEnvironmentVariables(cxt.Viper, cmd)
			}

			// Initialize the context if not already configured (by tests)
			if cxt.App == nil {
				k8sClient, svcatClient, namespace, err := getClients(opts.KubeConfig, opts.KubeContext)
				if err != nil {
					return err
				}

				app, err := svcat.NewApp(k8sClient, svcatClient, namespace)
				if err != nil {
					return err
				}

				cxt.App = app
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprint(cxt.Output, cmd.UsageString())
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&opts.KubeContext, "kube-context", "", "name of the kube context to use")
	cmd.PersistentFlags().StringVar(&opts.KubeConfig, "kubeconfig", "", "path to kubeconfig file. Overrides $KUBECONFIG")

	cmd.AddCommand(newGetCmd(cxt))
	cmd.AddCommand(newDescribeCmd(cxt))
	cmd.AddCommand(instance.NewProvisionCmd(cxt))
	cmd.AddCommand(instance.NewDeprovisionCmd(cxt))
	cmd.AddCommand(binding.NewBindCmd(cxt))
	cmd.AddCommand(binding.NewUnbindCmd(cxt))
	cmd.AddCommand(newSyncCmd(cxt))
	cmd.AddCommand(newInstallCmd(cxt))
	cmd.AddCommand(newTouchCmd(cxt))
	cmd.AddCommand(versions.NewVersionCmd(cxt))
	cmd.AddCommand(newCompletionCmd(cxt))

	return cmd
}

func newSyncCmd(cxt *command.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sync",
		Short:   "Syncs service catalog for a service broker",
		Aliases: []string{"relist"},
	}
	cmd.AddCommand(broker.NewSyncCmd(cxt))

	return cmd
}

func newGetCmd(cxt *command.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "List a resource, optionally filtered by name",
	}
	cmd.AddCommand(binding.NewGetCmd(cxt))
	cmd.AddCommand(broker.NewGetCmd(cxt))
	cmd.AddCommand(class.NewGetCmd(cxt))
	cmd.AddCommand(instance.NewGetCmd(cxt))
	cmd.AddCommand(plan.NewGetCmd(cxt))

	return cmd
}

func newDescribeCmd(cxt *command.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Show details of a specific resource",
	}
	cmd.AddCommand(binding.NewDescribeCmd(cxt))
	cmd.AddCommand(broker.NewDescribeCmd(cxt))
	cmd.AddCommand(class.NewDescribeCmd(cxt))
	cmd.AddCommand(instance.NewDescribeCmd(cxt))
	cmd.AddCommand(plan.NewDescribeCmd(cxt))

	return cmd
}

func newInstallCmd(cxt *command.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use: "install",
	}
	cmd.AddCommand(plugin.NewInstallCmd(cxt))

	return cmd
}

func newTouchCmd(cxt *command.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "touch",
		Short: "Force Service Catalog to reprocess a resource",
	}
	cmd.AddCommand(instance.NewTouchCommand(cxt))
	return cmd
}

func newCompletionCmd(ctx *command.Context) *cobra.Command {
	return completion.NewCompletionCmd(ctx)
}

// getClients loads api clients based on the plugin context if present, otherwise the specified kube config.
func getClients(kubeConfig, kubeContext string) (k8sClient k8sclient.Interface, svcatClient svcatclient.Interface, namespaces string, err error) {
	var restConfig *rest.Config
	var config clientcmd.ClientConfig

	if plugin.IsPlugin() {
		restConfig, config, err = pluginutils.InitClientAndConfig()
		if err != nil {
			return nil, nil, "", fmt.Errorf("could not get Kubernetes config from kubectl plugin context: %s", err)
		}
	} else {
		config = kube.GetConfig(kubeContext, kubeConfig)
		restConfig, err = config.ClientConfig()
		if err != nil {
			return nil, nil, "", fmt.Errorf("could not get Kubernetes config for context %q: %s", kubeContext, err)
		}
	}

	namespace, _, err := config.Namespace()
	k8sClient, err = k8sclient.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, "", err
	}
	svcatClient, err = svcatclient.NewForConfig(restConfig)
	return k8sClient, svcatClient, namespace, nil
}
