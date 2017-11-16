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

// NOTE: This file is based on a similar test at
// https://github.com/kubernetes/kubernetes/blob/master/test/integration/etcd/etcd_storage_path_test.go
// Changes made to that file should be monitored to see if they are applicable
// here as well.

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	"k8s.io/apimachinery/pkg/util/sets"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"

	"github.com/coreos/etcd/clientv3"

	appserver "github.com/kubernetes-incubator/service-catalog/cmd/apiserver/app/server"
	"github.com/kubernetes-incubator/service-catalog/pkg/api"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	registryserver "github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/server"
)

// Etcd data for all persisted objects.
var etcdStorageData = map[schema.GroupVersionResource]struct {
	stub             string                   // Valid JSON stub to use during create
	prerequisites    []prerequisite           // Optional, ordered list of JSON objects to create before stub
	expectedEtcdPath string                   // Expected location of object in etcd, do not use any variables, constants, etc to derive this value - always supply the full raw string
	expectedGVK      *schema.GroupVersionKind // The GVK that we expect this object to be stored as - leave this nil to use the default
}{
	gvr("servicecatalog.k8s.io", "v1beta1", "clusterservicebrokers"): {
		stub:             `{"metadata": {"name": "broker1"}, "spec": {"url": "https://broker1.com"}}`,
		expectedEtcdPath: "/servicecatalog.k8s.io/clusterservicebrokers/broker1",
	},
	gvr("servicecatalog.k8s.io", "v1beta1", "clusterserviceclasses"): {
		stub:             `{"metadata": {"name": "class1"}, "spec": {"clusterServiceBrokerName": "broker1", "externalName": "class-name1", "externalID": "class1", "description": "desc"}}`,
		expectedEtcdPath: "/servicecatalog.k8s.io/clusterserviceclasses/class1",
	},
	gvr("servicecatalog.k8s.io", "v1beta1", "clusterserviceplans"): {
		stub:             `{"metadata": {"name": "plan1"}, "spec": {"clusterServiceBrokerName": "broker1", "externalName": "plan-name1", "externalID": "plan1", "description": "desc", "clusterServiceClassRef": {"Name": "class1"}}}`,
		expectedEtcdPath: "/servicecatalog.k8s.io/clusterserviceplans/plan1",
	},
	gvr("servicecatalog.k8s.io", "v1beta1", "serviceinstances"): {
		stub:             `{"metadata": {"namespace": "etcdstoragepathtestnamespace", "name": "instance1"}, "spec": {"clusterServiceClassExternalName": "class1", "clusterServicePlanExternalName": "plan1", "externalID": "instance1"}}`,
		expectedEtcdPath: "/servicecatalog.k8s.io/serviceinstances/etcdstoragepathtestnamespace/instance1",
	},
	gvr("servicecatalog.k8s.io", "v1beta1", "servicebindings"): {
		stub:             `{"metadata": {"namespace": "etcdstoragepathtestnamespace", "name": "binding1"}, "spec": {"instanceRef": {"name": "instance1"}, "externalID": "binding1"}}`,
		expectedEtcdPath: "/servicecatalog.k8s.io/servicebindings/etcdstoragepathtestnamespace/binding1",
	},
}

// Be very careful when whitelisting an object as ephemeral.
// Doing so removes the safety we gain from this test by skipping that object.
var ephemeralWhiteList = createEphemeralWhiteList()

// Only add kinds to this list when there is no way to create the object
var kindWhiteList = sets.NewString(
	// k8s.io/kubernetes/pkg/api/v1
	"DeleteOptions",
	"ExportOptions",
	"ListOptions",
	"GetOptions",
	"APIGroup",
	"APIVersions",
	// --

	// k8s.io/kubernetes/pkg/watch/versioned
	"WatchEvent",
	// --

	// k8s.io/kubernetes/pkg/api/unversioned
	"Status",
	// --
)

// namespace used for all tests, do not change this
const testEtcdStorageNamespace = "etcdstoragepathtestnamespace"

// TestEtcdStoragePath tests to make sure that all objects are stored in an expected location in etcd.
// It will start failing when a new type is added to ensure that all future types are added to this test.
// It will also fail when a type gets moved to a different location. Be very careful in this situation because
// it essentially means that you will be break old clusters unless you create some migration path for the old data.
func TestEtcdStoragePath(t *testing.T) {
	certDir, _ := ioutil.TempDir("", "test-integration-etcd")
	defer os.RemoveAll(certDir)

	client, serverOptions, kvClient, mapper, shutdownServer := startRealMasterOrDie(t, certDir)
	defer func() {
		dumpEtcdKVOnFailure(t, kvClient)
	}()
	defer shutdownServer()

	kindSeen := sets.NewString()
	pathSeen := map[string][]schema.GroupVersionResource{}
	etcdSeen := map[schema.GroupVersionResource]empty{}
	ephemeralSeen := map[schema.GroupVersionResource]empty{}
	cohabitatingResources := map[string]map[schema.GroupVersionKind]empty{}

	for gvk, apiType := range api.Scheme.AllKnownTypes() {
		// we do not care about internal objects or lists // TODO make sure this is always true
		if gvk.Version == runtime.APIVersionInternal || strings.HasSuffix(apiType.Name(), "List") {
			continue
		}

		kind := gvk.Kind
		pkgPath := apiType.PkgPath()

		if kindWhiteList.Has(kind) {
			kindSeen.Insert(kind)
			continue
		}

		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			t.Errorf("unexpected error getting mapping for %s from %s with GVK %s: %v", kind, pkgPath, gvk, err)
			continue
		}

		gvResource := gvk.GroupVersion().WithResource(mapping.Resource)
		etcdSeen[gvResource] = empty{}

		testData, hasTest := etcdStorageData[gvResource]
		_, isEphemeral := ephemeralWhiteList[gvResource]

		if !hasTest && !isEphemeral {
			t.Errorf("no test data for %s from %s.  Please add a test for your new type to etcdStorageData.", kind, pkgPath)
			continue
		}

		if hasTest && isEphemeral {
			t.Errorf("duplicate test data for %s from %s.  Object has both test data and is ephemeral.", kind, pkgPath)
			continue
		}

		if isEphemeral { // TODO it would be nice if we could remove this and infer if an object is not stored in etcd
			// t.Logf("Skipping test for %s from %s", kind, pkgPath)
			ephemeralSeen[gvResource] = empty{}
			delete(etcdSeen, gvResource)
			continue
		}

		if len(testData.expectedEtcdPath) == 0 {
			t.Errorf("empty test data for %s from %s", kind, pkgPath)
			continue
		}
		expectedEtcdPath := serverOptions.EtcdOptions.StorageConfig.Prefix + testData.expectedEtcdPath

		shouldCreate := len(testData.stub) != 0 // try to create only if we have a stub

		var input *metaObject
		if shouldCreate {
			if input, err = jsonToMetaObject([]byte(testData.stub)); err != nil || input.isEmpty() {
				t.Errorf("invalid test data for %s from %s: %v", kind, pkgPath, err)
				continue
			}
		}

		func() { // forces defer to run per iteration of the for loop
			all := &[]cleanupData{}
			defer func() {
				if !t.Failed() { // do not cleanup if test has already failed since we may need things in the etcd dump
					if err := client.cleanup(all); err != nil {
						t.Fatalf("failed to clean up etcd: %#v", err)
					}
				}
			}()

			if err := client.createPrerequisites(mapper, testEtcdStorageNamespace, testData.prerequisites, all); err != nil {
				t.Errorf("failed to create prerequisites for %s from %s: %#v", kind, pkgPath, err)
				return
			}

			if shouldCreate { // do not try to create items with no stub
				if err := client.create(testData.stub, testEtcdStorageNamespace, mapping, all); err != nil {
					t.Errorf("failed to create stub for %s from %s: %#v", kind, pkgPath, err)
					return
				}
			}

			output, err := getFromEtcd(kvClient, expectedEtcdPath)
			if err != nil {
				t.Errorf("failed to get from etcd for %s from %s: %#v", kind, pkgPath, err)
				return
			}

			expectedGVK := gvk
			if testData.expectedGVK != nil {
				if gvk == *testData.expectedGVK {
					t.Errorf("GVK override %s for %s from %s is unnecessary or something was changed incorrectly", testData.expectedGVK, kind, pkgPath)
				}
				expectedGVK = *testData.expectedGVK
			}

			actualGVK := output.getGVK()
			if actualGVK != expectedGVK {
				t.Errorf("GVK for %s from %s does not match, expected %s got %s", kind, pkgPath, expectedGVK, actualGVK)
			}

			if !apiequality.Semantic.DeepDerivative(input, output) {
				t.Errorf("Test stub for %s from %s does not match: %s", kind, pkgPath, diff.ObjectGoPrintDiff(input, output))
			}

			addGVKToEtcdBucket(cohabitatingResources, actualGVK, getEtcdBucket(expectedEtcdPath))
			pathSeen[expectedEtcdPath] = append(pathSeen[expectedEtcdPath], gvResource)
		}()
	}

	if inEtcdData, inEtcdSeen := diffMaps(etcdStorageData, etcdSeen); len(inEtcdData) != 0 || len(inEtcdSeen) != 0 {
		t.Errorf("etcd data does not match the types we saw:\nin etcd data but not seen:\n%s\nseen but not in etcd data:\n%s", inEtcdData, inEtcdSeen)
	}

	if inEphemeralWhiteList, inEphemeralSeen := diffMaps(ephemeralWhiteList, ephemeralSeen); len(inEphemeralWhiteList) != 0 || len(inEphemeralSeen) != 0 {
		t.Errorf("ephemeral whitelist does not match the types we saw:\nin ephemeral whitelist but not seen:\n%s\nseen but not in ephemeral whitelist:\n%s", inEphemeralWhiteList, inEphemeralSeen)
	}

	if inKindData, inKindSeen := diffMaps(kindWhiteList, kindSeen); len(inKindData) != 0 || len(inKindSeen) != 0 {
		t.Errorf("kind whitelist data does not match the types we saw:\nin kind whitelist but not seen:\n%s\nseen but not in kind whitelist:\n%s", inKindData, inKindSeen)
	}

	for bucket, gvks := range cohabitatingResources {
		if len(gvks) != 1 {
			gvkStrings := []string{}
			for key := range gvks {
				gvkStrings = append(gvkStrings, keyStringer(key))
			}
			t.Errorf("cohabitating resources in etcd bucket %s have inconsistent GVKs\nyou may need to use DefaultStorageFactory.AddCohabitatingResources to sync the GVK of these resources:\n%s", bucket, gvkStrings)
		}
	}

	for path, gvrs := range pathSeen {
		if len(gvrs) != 1 {
			gvrStrings := []string{}
			for _, key := range gvrs {
				gvrStrings = append(gvrStrings, keyStringer(key))
			}
			t.Errorf("invalid test data, please ensure all expectedEtcdPath are unique, path %s has duplicate GVRs:\n%s", path, gvrStrings)
		}
	}
}

func startRealMasterOrDie(t *testing.T, certDir string) (*allClient, *appserver.ServiceCatalogServerOptions, clientv3.KV, meta.RESTMapper, func()) {
	_, serverOptions, catalogClientConfig, kvClient, shutdownServer := getFreshApiserverAndClientAndEtcdClient(t, registryserver.StorageTypeEtcd.String(), func() runtime.Object {
		return &servicecatalog.ClusterServiceBroker{}
	})

	client, err := newClient(*catalogClientConfig)
	if err != nil {
		t.Fatal(err)
	}

	mapper := api.Registry.RESTMapper()

	return client, serverOptions, kvClient, mapper, shutdownServer
}

func dumpEtcdKVOnFailure(t *testing.T, kvClient clientv3.KV) {
	if t.Failed() {
		response, err := kvClient.Get(context.Background(), "/", clientv3.WithPrefix())
		if err != nil {
			t.Fatal(err)
		}

		for _, kv := range response.Kvs {
			t.Error(string(kv.Key), "->", string(kv.Value))
		}
	}
}

func addGVKToEtcdBucket(cohabitatingResources map[string]map[schema.GroupVersionKind]empty, gvk schema.GroupVersionKind, bucket string) {
	if cohabitatingResources[bucket] == nil {
		cohabitatingResources[bucket] = map[schema.GroupVersionKind]empty{}
	}
	cohabitatingResources[bucket][gvk] = empty{}
}

// getEtcdBucket assumes the last segment of the given etcd path is the name of the object.
// Thus it strips that segment to extract the object's storage "bucket" in etcd. We expect
// all objects that share the a bucket (cohabitating resources) to be stored as the same GVK.
func getEtcdBucket(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx == -1 {
		panic("path with no slashes " + path)
	}
	bucket := path[:idx]
	if len(bucket) == 0 {
		panic("invalid bucket for path " + path)
	}
	return bucket
}

// stable fields to compare as a sanity check
type metaObject struct {
	// all of type meta
	Kind       string `json:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`
	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,2,opt,name=apiVersion"`

	// parts of object meta
	Metadata struct {
		Name      string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
		Namespace string `json:"namespace,omitempty" protobuf:"bytes,2,opt,name=namespace"`
	} `json:"metadata,omitempty" protobuf:"bytes,3,opt,name=metadata"`
}

func (obj *metaObject) getGVK() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)
}

func (obj *metaObject) isEmpty() bool {
	return obj == nil || *obj == metaObject{} // compare to zero value since all fields are strings
}

type prerequisite struct {
	gvrData schema.GroupVersionResource
	stub    string
}

type empty struct{}

type cleanupData struct {
	obj     runtime.Object
	mapping *meta.RESTMapping
}

func gvr(g, v, r string) schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: g, Version: v, Resource: r}
}

func gvkP(g, v, k string) *schema.GroupVersionKind {
	return &schema.GroupVersionKind{Group: g, Version: v, Kind: k}
}

func createEphemeralWhiteList(gvrs ...schema.GroupVersionResource) map[schema.GroupVersionResource]empty {
	ephemeral := map[schema.GroupVersionResource]empty{}
	for _, gvResource := range gvrs {
		if _, ok := ephemeral[gvResource]; ok {
			panic("invalid ephemeral whitelist contains duplicate keys")
		}
		ephemeral[gvResource] = empty{}
	}
	return ephemeral
}

func jsonToMetaObject(stub []byte) (*metaObject, error) {
	obj := &metaObject{}
	if err := json.Unmarshal(stub, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func keyStringer(i interface{}) string {
	base := "\n\t"
	switch key := i.(type) {
	case string:
		return base + key
	case schema.GroupVersionResource:
		return base + key.String()
	case schema.GroupVersionKind:
		return base + key.String()
	default:
		panic("unexpected type")
	}
}

type allClient struct {
	client  *http.Client
	config  *restclient.Config
	backoff restclient.BackoffManager
}

func (c *allClient) verb(verb string, gvk schema.GroupVersionKind) (*restclient.Request, error) {
	apiPath := "/apis"
	baseURL, versionedAPIPath, err := restclient.DefaultServerURL(c.config.Host, apiPath, gvk.GroupVersion(), true)
	if err != nil {
		return nil, err
	}
	contentConfig := c.config.ContentConfig
	gv := gvk.GroupVersion()
	contentConfig.GroupVersion = &gv
	serializers, err := createSerializers(contentConfig)
	if err != nil {
		return nil, err
	}
	return restclient.NewRequest(c.client, verb, baseURL, versionedAPIPath, contentConfig, *serializers, c.backoff, c.config.RateLimiter), nil
}

func (c *allClient) create(stub, ns string, mapping *meta.RESTMapping, all *[]cleanupData) error {
	req, err := c.verb("POST", mapping.GroupVersionKind)
	if err != nil {
		return err
	}
	namespaced := mapping.Scope.Name() == meta.RESTScopeNameNamespace
	output, err := req.NamespaceIfScoped(ns, namespaced).Resource(mapping.Resource).Body(strings.NewReader(stub)).Do().Get()
	if err != nil {
		return err
	}
	*all = append(*all, cleanupData{output, mapping})
	return nil
}

func (c *allClient) destroy(obj runtime.Object, mapping *meta.RESTMapping) error {
	req, err := c.verb("DELETE", mapping.GroupVersionKind)
	if err != nil {
		return err
	}
	namespaced := mapping.Scope.Name() == meta.RESTScopeNameNamespace
	name, err := mapping.MetadataAccessor.Name(obj)
	if err != nil {
		return err
	}
	ns, err := mapping.MetadataAccessor.Namespace(obj)
	if err != nil {
		return err
	}
	return req.NamespaceIfScoped(ns, namespaced).Resource(mapping.Resource).Name(name).Do().Error()
}

func (c *allClient) cleanup(all *[]cleanupData) error {
	for i := len(*all) - 1; i >= 0; i-- { // delete in reverse order in case creation order mattered
		obj := (*all)[i].obj
		mapping := (*all)[i].mapping

		if err := c.destroy(obj, mapping); err != nil {
			return err
		}
	}
	return nil
}

func (c *allClient) createPrerequisites(mapper meta.RESTMapper, ns string, prerequisites []prerequisite, all *[]cleanupData) error {
	for _, prerequisite := range prerequisites {
		gvk, err := mapper.KindFor(prerequisite.gvrData)
		if err != nil {
			return err
		}
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}
		if err := c.create(prerequisite.stub, ns, mapping, all); err != nil {
			return err
		}
	}
	return nil
}

func newClient(config restclient.Config) (*allClient, error) {
	config.ContentConfig.NegotiatedSerializer = api.Codecs
	config.ContentConfig.ContentType = "application/json"
	config.Timeout = 30 * time.Second
	config.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(3, 10)

	transport, err := restclient.TransportFor(&config)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	backoff := &restclient.URLBackoff{
		Backoff: flowcontrol.NewBackOff(1*time.Second, 10*time.Second),
	}

	return &allClient{
		client:  client,
		config:  &config,
		backoff: backoff,
	}, nil
}

// copied from restclient
func createSerializers(config restclient.ContentConfig) (*restclient.Serializers, error) {
	mediaTypes := config.NegotiatedSerializer.SupportedMediaTypes()
	contentType := config.ContentType
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, fmt.Errorf("the content type specified in the client configuration is not recognized: %v", err)
	}
	info, ok := runtime.SerializerInfoForMediaType(mediaTypes, mediaType)
	if !ok {
		if len(contentType) != 0 || len(mediaTypes) == 0 {
			return nil, fmt.Errorf("no serializers registered for %s", contentType)
		}
		info = mediaTypes[0]
	}

	internalGV := schema.GroupVersions{
		{
			Group:   config.GroupVersion.Group,
			Version: runtime.APIVersionInternal,
		},
		// always include the legacy group as a decoding target to handle non-error `Status` return types
		{
			Group:   "",
			Version: runtime.APIVersionInternal,
		},
	}

	s := &restclient.Serializers{
		Encoder: config.NegotiatedSerializer.EncoderForVersion(info.Serializer, *config.GroupVersion),
		Decoder: config.NegotiatedSerializer.DecoderToVersion(info.Serializer, internalGV),

		RenegotiatedDecoder: func(contentType string, params map[string]string) (runtime.Decoder, error) {
			info, ok := runtime.SerializerInfoForMediaType(mediaTypes, contentType)
			if !ok {
				return nil, fmt.Errorf("serializer for %s not registered", contentType)
			}
			return config.NegotiatedSerializer.DecoderToVersion(info.Serializer, internalGV), nil
		},
	}
	if info.StreamSerializer != nil {
		s.StreamingSerializer = info.StreamSerializer.Serializer
		s.Framer = info.StreamSerializer.Framer
	}

	return s, nil
}

func getFromEtcd(keys clientv3.KV, path string) (*metaObject, error) {
	response, err := keys.Get(context.Background(), path)
	if err != nil {
		return nil, err
	}
	if response.More || response.Count != 1 || len(response.Kvs) != 1 {
		return nil, fmt.Errorf("Invalid etcd response (not found == %v): %#v", response.Count == 0, response)
	}
	return jsonToMetaObject(response.Kvs[0].Value)
}

func diffMaps(a, b interface{}) ([]string, []string) {
	inA := diffMapKeys(a, b, keyStringer)
	inB := diffMapKeys(b, a, keyStringer)
	return inA, inB
}

func diffMapKeys(a, b interface{}, stringer func(interface{}) string) []string {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	ret := []string{}

	for _, ka := range av.MapKeys() {
		kat := ka.Interface()
		found := false
		for _, kb := range bv.MapKeys() {
			kbt := kb.Interface()
			if kat == kbt {
				found = true
				break
			}
		}
		if !found {
			ret = append(ret, stringer(kat))
		}
	}

	return ret
}

type allResourceSource struct{}

func (*allResourceSource) AnyVersionOfResourceEnabled(resource schema.GroupResource) bool { return true }
func (*allResourceSource) AllResourcesForVersionEnabled(version schema.GroupVersion) bool { return true }
func (*allResourceSource) AnyResourcesForGroupEnabled(group string) bool                  { return true }
func (*allResourceSource) ResourceEnabled(resource schema.GroupVersionResource) bool      { return true }
func (*allResourceSource) AnyResourcesForVersionEnabled(version schema.GroupVersion) bool { return true }
