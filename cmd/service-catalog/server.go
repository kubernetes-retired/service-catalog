package main

import (
	"os"

	"github.com/golang/glog"
	// set up logging the k8s way
	"k8s.io/kubernetes/pkg/util/logs"

	"github.com/kubernetes-incubator/service-catalog/pkg/cmd/server"
	// TODO: may be necessary
	_ "k8s.io/kubernetes/pkg/api/install"
	// install our API groups
	_ "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/install"
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
