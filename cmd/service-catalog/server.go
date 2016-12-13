package main

import (
	"github.com/golang/glog"
	"os"

	// set up logging the k8s way
	"k8s.io/kubernetes/pkg/util/logs"

	"github.com/kubernetes-incubator/service-catalog/pkg/cmd/server"
	// commented out until I know what this does
	// install all APIs
	// _ "github.com/openshift/kube-aggregator/pkg/apis/apifederation/install"
	// _ "k8s.io/kubernetes/pkg/api/install"
)

func main() {
	logs.InitLogs()
	// make sure we print all the logs while shutting down.
	defer logs.FlushLogs()

	cmd := server.NewCommandServer(os.Stdout)
	if err := cmd.Execute(); err != nil {
		glog.Errorln(err)
		os.Exit(1)
	}
}
