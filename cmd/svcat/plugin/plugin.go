// plugin package helps apply kubectl plugin-specific cli configuration.
// See https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/#accessing-runtime-attributes.
package plugin

import (
	"os"

	"github.com/spf13/viper"
)

// IsPlugin determines if the cli is running as a kubectl plugin
func IsPlugin() bool {
	_, ok := os.LookupEnv("KUBECTL_PLUGINS_CALLER")
	return ok
}

// Bind plugin-specific environment variables to command flags.
func BindEnvironmentVariables(vip *viper.Viper) {
	// KUBECTL_PLUGINS_CURRENT_NAMESPACE provides the final namespace
	// computed by kubectl.
	vip.BindEnv("namespace", "KUBECTL_PLUGINS_CURRENT_NAMESPACE")

	// kubectl intercepts all flags passed to a plugin, and replaces them
	// with prefixed environment variables
	// --foo becomes KUBECTL_PLUGINS_LOCAL_FLAG_FOO
	vip.SetEnvPrefix("KUBECTL_PLUGINS_LOCAL_FLAG")
}
