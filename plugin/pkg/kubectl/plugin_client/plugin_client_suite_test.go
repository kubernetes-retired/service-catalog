package plugin_client_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPluginClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PluginClient Suite")
}

var _ = BeforeEach(func() {
	os.Setenv("KUBECTL_PLUGINS_GLOBAL_FLAG_KUBECONFIG", "assets/config")
})
