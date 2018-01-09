package environment

import (
	"os"

	"github.com/spf13/pflag"
)

// EnvSettings describes all of the environment settings.
type EnvSettings struct {
	// KubeContext is the name of the kube context.
	KubeContext string
	// KubeConfig is the name of the kubeconfig file.
	KubeConfig string
}

// AddFlags binds flags to the given flagset.
func (s *EnvSettings) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.KubeContext, "kube-context", "", "name of the kube context to use")
	fs.StringVar(&s.KubeConfig, "kubeconfig", "", "path to kubeconfig file. Overrides $KUBECONFIG")
}

// Init sets values from the environment if they aren't set explicitly.
func (s *EnvSettings) Init() {
	if s.KubeConfig == "" {
		if kubeconfig, ok := os.LookupEnv("KUBECONFIG"); ok {
			s.KubeConfig = kubeconfig
		}
	}
}
