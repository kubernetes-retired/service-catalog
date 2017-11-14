/*
Copyright 2017 The Kubernetes Authors.

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

package integration

import (
	"fmt"
	"github.com/coreos/etcd/embed"
	"github.com/golang/glog"
	"io/ioutil"
	"net/url"
	"os"
	"testing"
	"time"
)

const (
	EtcdPeerURLs   = "http://localhost:2381"
	EtcdClientURLs = "http://localhost:2378"
)

type EtcdContext struct {
	etcd     *embed.Etcd
	dir      string
	Endpoint string
}

var etcdContext = EtcdContext{}

func startEtcd() error {
	var err error
	if etcdContext.dir, err = ioutil.TempDir(os.TempDir(), "service_catalog_integration_test"); err != nil {
		return fmt.Errorf("Could not create TempDir: %v", err)
	}
	cfg := embed.NewConfig()
	cfg.Dir = etcdContext.dir

	if etcdContext.etcd, err = embed.StartEtcd(cfg); err != nil {
		return fmt.Errorf("Failed starting etcd: %+v", err)
	}

	select {
	case <-etcdContext.etcd.Server.ReadyNotify():
		glog.Info("Server is ready!")
	case <-time.After(60 * time.Second):
		etcdContext.etcd.Server.Stop() // trigger a shutdown
		glog.Error("Server took too long to start!")
	}
	return nil
}

func stopEtcd() {
	etcdContext.etcd.Server.Stop()
	os.RemoveAll(etcdContext.dir)
}

func TestStartEtcd(t *testing.T) {
	tdir, err := ioutil.TempDir(os.TempDir(), "start-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tdir)
	cfg := embed.NewConfig()
	cfg.Dir = tdir

	lpurl, _ := url.Parse(EtcdPeerURLs)
	lcurl, _ := url.Parse(EtcdClientURLs)

	cfg.LPUrls = []url.URL{*lpurl}
	cfg.LCUrls = []url.URL{*lcurl}

	if _, err = embed.StartEtcd(cfg); err != nil {
		t.Fatalf("got %+v", err)
	}
}

func TestMain(m *testing.M) {
	// Setup
	if err := startEtcd(); err != nil {
		fmt.Println("Failed to start etcd, %v", err)
		os.Exit(1)
	}

	// Tests
	result := m.Run()

	// Teardown
	stopEtcd()
	os.Exit(result)
}
