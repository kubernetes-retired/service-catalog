package e2e_test

import (
	"flag"
	"testing"

	"k8s.io/client-go/1.5/kubernetes"

	kv1 "k8s.io/client-go/1.5/pkg/api/v1"

	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	clientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"

	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/restclient"

	// avoid error `servicecatalog/v1alpha1 is not enabled`

	_ "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/install"
	// avoid error `no kind is registered for the type v1.ListOptions`

	_ "k8s.io/kubernetes/pkg/api/install"

	// client-go has got to come from somewhere, and this is part of it, cause we need the tools

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	servicecatalogclient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
)

var kubeconfigPath string
var po *clientcmd.PathOptions

func init() {
	po = clientcmd.NewDefaultPathOptions()
	flag.StringVar(&po.LoadingRules.ExplicitPath, po.ExplicitFileFlag, po.LoadingRules.ExplicitPath, "what do you think this is?")
}

// TestBrokerInstall relies on:
// - a running k8s
// - running apiserver
// - running broker (maybe? if we don't actually contact it, this shouldn't matter)
// - controller (to do the watch on the brokers and contact it to change it's status)
//
// It will call the apiserver to
func TestBrokerInstall(t *testing.T) {
	_ = clientcmd.NewDefaultClientConfig(*clientcmdapi.NewConfig(), &clientcmd.ConfigOverrides{})

	config2, err := po.GetStartingConfig()
	//kubeconfigPath, err := filepath.Abs(po.kubeconfigPath)
	//kclient := util.NewFactory(kconfig)
	//kgv := kclient.Core().RESTClient().APIVersion()
	//t.Log(kgv)
	if nil != err {
		t.Fatalf("Failed to create a kube config, could not figure out the path\n:%v\n", err)
	}
	// k8s client
	kconfig, err := clientcmd.BuildConfigFromFlags("", po.LoadingRules.ExplicitPath)
	if err != nil {
		t.Fatalf("Failed to create a kube config\n:%v\n", err)
	}
	//kconfig := &rest.Config{}
	//kconfig.Host = "https://localhost:6443"
	//kconfig.Insecure = true
	kclient, err := kubernetes.NewForConfig(kconfig)
	if nil != err {
		t.Fatal("can't make the client from the config", err)
	}
	kgv = kclient.Core().GetRESTClient().APIVersion()
	t.Log(kgv)

	pod := &kv1.Pod{}
	pod, err = kclient.Core().Pods("default").Create(pod)
	if nil != err {
		t.Fatal("error creating pod\n", err)
	}

	// sc client
	config := &restclient.Config{}
	config.Host = "https://localhost:30000"
	config.Insecure = true
	client, err := servicecatalogclient.NewForConfig(config)
	if nil != err {
		t.Fatal("can't make the client from the config", err)
	}
	gv := client.Servicecatalog().RESTClient().APIVersion()
	t.Log(gv)

	brokerClient := client.Servicecatalog().Brokers()
	_ = brokerClient.Delete("test-broker", &v1.DeleteOptions{})

	brokers, err := brokerClient.List(v1.ListOptions{})
	if nil != err {
		t.Fatal("error listing the broker\n", err)
	}
	if len(brokers.Items) > 0 {
		t.Fatalf("brokers should not exist on start, had %v brokers", len(brokers.Items))
	}
	t.Log(brokers)

	broker := &v1alpha1.Broker{
		ObjectMeta: v1.ObjectMeta{Name: "test-broker"},
		Spec: v1alpha1.BrokerSpec{
			URL:          "https://example.com",
			AuthUsername: "auth username field value",
			AuthPassword: "auth password field value",
			OSBGUID:      "OSBGUID field",
		},
	}

	brokerServer, err := brokerClient.Create(broker)
	if nil != err {
		t.Fatal("error creating the broker\n", err, "\nbroker ", broker)
	}
	if broker.Name != brokerServer.Name {
		t.Fatalf("didn't get the same broker back from the server \n%+v\n%+v", broker, brokerServer)
	}

	brokers, err = brokerClient.List(v1.ListOptions{})
	if 1 != len(brokers.Items) {
		t.Fatalf("should have exactly one broker, had %v brokers", len(brokers.Items))
	}

	brokerServer, err = brokerClient.Get(broker.Name)
	if broker.Name != brokerServer.Name &&
		broker.ResourceVersion == brokerServer.ResourceVersion {
		t.Fatalf("didn't get the same broker back from the server \n%+v\n%+v", broker, brokerServer)
	}
	err = brokerClient.Delete("test-broker", &v1.DeleteOptions{})
	if nil != err {
		t.Fatal("broker should be deleted", err)
	}

	brokerDeleted, err := brokerClient.Get("test-broker")
	if nil == err {
		t.Fatal("broker should be deleted", brokerDeleted)
	}

}
