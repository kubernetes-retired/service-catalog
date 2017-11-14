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
	"github.com/coreos/etcd/embed"
	"io/ioutil"
	"net/url"
	"os"
	"testing"
)

const (
	EtcdPeerURLs   = "http://localhost:2381"
	EtcdClientURLs = "http://localhost:2378"
)

var etcdTempDir string
var etcd *embed.Etcd

func startEtcd(t *testing.T) {
	etcdTempDir, err := ioutil.TempDir(os.TempDir(), "service_catalog_integration_test")
	if err != nil {
		t.Fatal(err)
	}
	cfg := embed.NewConfig()
	cfg.Dir = etcdTempDir

	if etcd, err = embed.StartEtcd(cfg); err != nil {
		t.Fatalf("Failed starting etcd: %+v", err)
	}
}

func stopEtcd(t *testing.T) {
	etcd.Server.Stop()
	os.RemoveAll(etcdTempDir)
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
