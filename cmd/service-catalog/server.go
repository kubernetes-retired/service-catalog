package main

import (
	"os"

	"github.com/golang/glog"
	// set up logging the k8s way
	"k8s.io/kubernetes/pkg/util/logs"

	"github.com/kubernetes-incubator/service-catalog/pkg/cmd/server"
	// this is necessary at startup to register the apis.
	// mhb doesn't like how disconnected this is from the rest of the setup of the apis.
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
		logs.FlushLogs()
		os.Exit(1)
	}
}
