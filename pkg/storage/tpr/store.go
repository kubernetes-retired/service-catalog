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

package tpr

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/etcd"
	"k8s.io/apiserver/pkg/storage/storagebackend/factory"
	restclient "k8s.io/client-go/rest"
)

type store struct {
	hasNamespace     bool
	codec            runtime.Codec
	defaultNamespace string
	cl               restclient.Interface
	singularKind     Kind
	singularShell    func(string, string) runtime.Object
	listKind         Kind
	listShell        func() runtime.Object
	checkObject      func(runtime.Object) error
	decodeKey        func(string) (string, string, error)
	versioner        storage.Versioner
}

type objState struct {
	obj  runtime.Object
	meta *storage.ResponseMeta
	rev  uint64
	data []byte
}

// New creates a new TPR-based storage.Interface implementation
func New(opts Options) (storage.Interface, factory.DestroyFunc) {
	return &store{
		hasNamespace:     opts.HasNamespace,
		codec:            opts.RESTOptions.StorageConfig.Codec,
		defaultNamespace: opts.DefaultNamespace,
		cl:               opts.RESTClient,
		singularKind:     opts.SingularKind,
		singularShell:    opts.NewSingularFunc,
		listKind:         opts.ListKind,
		listShell:        opts.NewListFunc,
		checkObject:      opts.CheckObjectFunc,
		decodeKey:        opts.Keyer.NamespaceAndNameFromKey,
		versioner:        etcd.APIObjectVersioner{},
	}, opts.DestroyFunc
}

// Versioner implements storage.Interface.Versioner.
func (s *store) Versioner() storage.Versioner {
	return s.versioner
}

// Get implements storage.Interface.Get.
func (s *store) Get(
	ctx context.Context,
	key string,
	resourceVersion string,
	out runtime.Object,
	ignoreNotFound bool,
) error {
	ns, name, err := s.decodeKey(key)
	if err != nil {
		glog.Errorf("decoding key %s: %s", key, err)
		return err
	}
	req := s.buildRequest("GET", ns, name, false)
	_, statusCode, err := s.exec(req, out)
	if err != nil {
		glog.Errorf("executing GET for %s/%s: %s", ns, name, err)
		return err
	}
	if statusCode == http.StatusNotFound {
		if ignoreNotFound {
			return runtime.SetZeroValue(out)
		}
		glog.Errorf("executing GET for %s/%s: not found", ns, name)
		return storage.NewKeyNotFoundError(key, 0)
	}
	if !s.hasNamespace {
		if err := removeNamespace(out); err != nil {
			glog.Errorf("removing namespace from %#v: %s", out, err)
			return err
		}
	}
	return nil
}

// Create implements storage.Interface.Create.
func (s *store) Create(
	ctx context.Context,
	key string,
	obj,
	out runtime.Object,
	ttl uint64,
) error {
	version, err := s.versioner.ObjectResourceVersion(obj)
	if err != nil {
		glog.Errorf("getting resource version for object %#v: %s", obj, err)
		return err
	}
	if version != 0 {
		return errors.New(
			"resourceVersion should not be set on objects to be created",
		)
	}
	ns, name, err := s.decodeKey(key)
	if err != nil {
		glog.Errorf("decoding key %s: %s", key, err)
		return err
	}
	// Check if this resource already exists
	req := s.buildRequest("GET", ns, name, false)
	_, statusCode, err := s.exec(req, nil)
	if err != nil {
		glog.Errorf("executing GET for %s/%s: %s", ns, name, err)
		return err
	}
	if statusCode != http.StatusNotFound {
		return storage.NewKeyExistsError(key, 0)
	}
	data, err := runtime.Encode(s.codec, obj)
	if err != nil {
		glog.Errorf("encoding %#v: %s", obj, err)
		return err
	}
	req = s.buildRequest("POST", ns, "", false).Body(data)
	_, statusCode, err = s.exec(req, out)
	if err != nil {
		glog.Errorf("executing POST of %s to %s: %s", string(data), ns, err)
		return err
	}
	if statusCode != http.StatusCreated {
		return fmt.Errorf("executing POST of %s to %s: status code %d", string(data), ns, statusCode)
	}
	return nil
}

// Delete implements storage.Interface.Delete.
func (s *store) Delete(
	ctx context.Context,
	key string,
	out runtime.Object,
	preconditions *storage.Preconditions,
) error {
	if preconditions == nil {
		return s.unconditionalDelete(ctx, key, out)
	}
	return s.conditionalDelete(ctx, key, out, preconditions)
}

func (s *store) unconditionalDelete(
	ctx context.Context,
	key string,
	out runtime.Object,
) error {
	ns, name, err := s.decodeKey(key)
	if err != nil {
		glog.Errorf("decoding key %s: %s", key, err)
		return err
	}
	// Check if this resource exists
	req := s.buildRequest("GET", ns, name, false)
	_, statusCode, err := s.exec(req, nil)
	if err != nil {
		glog.Errorf("executing GET for %s/%s: %s", ns, name, err)
		return err
	}
	if statusCode == http.StatusNotFound {
		return storage.NewKeyNotFoundError(key, 0)
	}
	req = s.buildRequest("DELETE", ns, name, false)
	_, statusCode, err = s.exec(req, out)
	if err != nil {
		glog.Errorf("executing DELETE of %s/%s: %s", ns, name, err)
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("executing DELETE of %s/%s: status code %d", ns, name, statusCode)
	}
	if !s.hasNamespace {
		if err := removeNamespace(out); err != nil {
			glog.Errorf("removing namespace from %#v: %s", out, err)
			return err
		}
	}
	return nil
}

func (s *store) conditionalDelete(
	ctx context.Context,
	key string,
	out runtime.Object,
	preconditions *storage.Preconditions,
) error {
	ns, name, err := s.decodeKey(key)
	if err != nil {
		glog.Errorf("decoding key %s: %s", key, err)
		return err
	}
	// Check if this resource exists
	req := s.buildRequest("GET", ns, name, false)
	res, statusCode, err := s.exec(req, nil)
	if err != nil {
		glog.Errorf("executing GET for %s/%s: %s", ns, name, err)
		return err
	}
	if statusCode == http.StatusNotFound {
		return storage.NewKeyNotFoundError(key, 0)
	}
	for {
		origState, err := s.getState(res, key, false)
		if err != nil {
			glog.Errorf("getting state for %s: %s", key, err)
			return err
		}
		err = checkPreconditions(key, preconditions, origState.obj)
		if err != nil {
			glog.Errorf("checking preconditions for %s: %s", key, err)
			return err
		}
		curObj := s.singularShell("", "")
		curReq := s.buildRequest("GET", ns, name, false)
		curRes, _, err := s.exec(curReq, curObj)
		if err != nil {
			glog.Errorf("executing GET for %s/%s: %s", ns, name, err)
			return err
		}
		curResourceVersion, err := s.versioner.ObjectResourceVersion(curObj)
		if err != nil {
			glog.Errorf("getting state from object %#v: %s", curObj, err)
			return err
		}
		if curResourceVersion == origState.rev {
			delReq := s.buildRequest("DELETE", ns, name, false)
			_, statusCode, err := s.exec(delReq, nil)
			if err != nil {
				glog.Errorf("executing DELETE of %s/%s: %s", ns, name, err)
				return err
			}
			if statusCode != http.StatusOK {
				return fmt.Errorf("executing DELETE of %s/%s: status code %d", ns, name, statusCode)
			}
		} else {
			res = curRes
			glog.V(4).Infof(
				"deletion of %s failed because of a conflict; going to retry",
				key,
			)
			continue
		}
		err = s.decode(origState.data, out)
		if err != nil {
			glog.Errorf("decoding %s: %s", string(origState.data), err)
			return err
		}
		err = s.versioner.UpdateObject(out, origState.rev)
		if err != nil {
			glog.Errorf(
				"updating object %#v with resource version %d: %s",
				out, origState.rev, err,
			)
		}
		return err
	}
}

// GuaranteedUpdate implements storage.Interface.GuaranteedUpdate.
func (s *store) GuaranteedUpdate(
	ctx context.Context,
	key string,
	out runtime.Object,
	ignoreNotFound bool,
	precondtions *storage.Preconditions,
	tryUpdate storage.UpdateFunc,
	suggestion ...runtime.Object,
) error {
	ns, name, err := s.decodeKey(key)
	if err != nil {
		glog.Errorf("decoding key %s: %s", key, err)
		return err
	}
	var origState *objState
	if len(suggestion) == 1 && suggestion[0] != nil {
		origState, err = s.getStateFromObject(suggestion[0])
		if err != nil {
			glog.Errorf("getting state from object %#v: %s", suggestion[0], err)
			return err
		}
	} else {
		req := s.buildRequest("GET", ns, name, false)
		res, _, err := s.exec(req, nil)
		if err != nil {
			glog.Errorf("executing GET for %s/%s: %s", ns, name, err)
			return err
		}
		origState, err = s.getState(res, key, ignoreNotFound)
		if err != nil {
			glog.Errorf("getting state for %s: %s", key, err)
			return err
		}
	}
	for {
		if err := checkPreconditions(key, precondtions, origState.obj); err != nil {
			glog.Errorf("checking preconditions for %s: %s", key, err)
			return err
		}
		ret, err := s.updateState(origState, tryUpdate)
		if err != nil {
			glog.Errorf("updating state: %s", err)
			return err
		}
		data, err := runtime.Encode(s.codec, ret)
		if err != nil {
			glog.Errorf("encoding %#v: %s", ret, err)
			return err
		}
		if bytes.Equal(data, origState.data) {
			err := s.decode(origState.data, out)
			if err != nil {
				glog.Errorf("decoding %s: %s", string(origState.data), err)
				return err
			}
			err = s.versioner.UpdateObject(out, origState.rev)
			if err != nil {
				glog.Errorf(
					"updating object %#v with resource version %d: %s",
					out, origState.rev, err,
				)
			}
			return err
		}
		curObj := s.singularShell("", "")
		curReq := s.buildRequest("GET", ns, name, false)
		curRes, _, err := s.exec(curReq, curObj)
		if err != nil {
			glog.Errorf("executing GET for %s/%s: %s", ns, name, err)
			return err
		}
		curResourceVersion, err := s.versioner.ObjectResourceVersion(curObj)
		if err != nil {
			glog.Errorf("getting resource version for object %#v: %s", curObj, err)
			return err
		}
		var putRes restclient.Result
		if curResourceVersion == origState.rev {
			putReq := s.buildRequest("PUT", ns, name, false).Body(data)
			putRes, _, err = s.exec(putReq, nil)
			if err != nil {
				glog.Errorf(
					"executing PUT of %s to %s/%s: %s",
					string(data), ns, name, err,
				)
				return err
			}
		} else {
			glog.V(4).Infof(
				"GuaranteedUpdate of %s failed because of a conflict, going to retry",
				key,
			)
			origState, err = s.getState(curRes, key, ignoreNotFound)
			if err != nil {
				glog.Errorf("getting state for %s: %s", key, err)
				return err
			}
			continue
		}
		var unknown runtime.Unknown
		if err := putRes.Into(&unknown); err != nil {
			glog.Errorf("reading response: %s", err)
			return err
		}
		obj := s.singularShell("", "")
		if err := s.decode(unknown.Raw, obj); err != nil {
			glog.Errorf("decoding response body %s: %s", string(unknown.Raw), err)
			return err
		}
		err = s.decode(data, out)
		if err != nil {
			glog.Errorf("decoding %s: %s", string(data), err)
			return err
		}
		newResourceVersion, err := s.versioner.ObjectResourceVersion(obj)
		if err != nil {
			glog.Errorf("getting resource version for object %#v: %s", obj, err)
			return err
		}
		err = s.versioner.UpdateObject(out, newResourceVersion)
		if err != nil {
			glog.Errorf(
				"updating object %#v with resource version %d: %s",
				out, newResourceVersion, err,
			)
		}
		return err
	}
}

// GetToList implements storage.Interface.GetToList.
func (s *store) GetToList(
	ctx context.Context,
	key string,
	resourceVersion string,
	pred storage.SelectionPredicate,
	listObj runtime.Object,
) error {
	return s.List(ctx, key, resourceVersion, pred, listObj)
}

// List implements storage.Interface.List.
func (s *store) List(
	ctx context.Context,
	key string,
	resourceVersion string,
	pred storage.SelectionPredicate,
	listObj runtime.Object,
) error {
	ns, _, err := s.decodeKey(key)
	if err != nil {
		glog.Errorf("decoding %s: %s", key, err)
		return err
	}
	req := s.buildRequest("GET", ns, "", false)
	_, statusCode, err := s.exec(req, listObj)
	if err != nil {
		glog.Errorf("executing GET for %s: %s", ns, err)
		return err
	}
	if statusCode != http.StatusOK {
		return storage.NewKeyNotFoundError(key, 0)
	}
	filter := storage.SimpleFilter(pred)
	filteredList := make([]runtime.Object, 0)
	if err := meta.EachListItem(listObj, func(obj runtime.Object) error {
		if filter(obj) {
			filteredList = append(filteredList, obj)
		}
		return nil
	}); err != nil {
		glog.Errorf("filtering list items: %s", err)
		return err
	}
	if err = meta.SetList(listObj, filteredList); err != nil {
		glog.Errorf("setting filtered list: %s", err)
		return err
	}
	if !s.hasNamespace {
		if err := meta.EachListItem(listObj, removeNamespace); err != nil {
			glog.Errorf("removing namespace from all items in list: %s", err)
			return err
		}
	}
	return nil
}

// Watch implements storage.Interface.Watch.
func (s *store) Watch(
	ctx context.Context,
	key string,
	resourceVersion string,
	p storage.SelectionPredicate,
) (watch.Interface, error) {
	ns, name, err := s.decodeKey(key)
	if err != nil {
		glog.Errorf("decoding key %s: %s", key, err)
		return nil, err
	}
	req := s.buildRequest("GET", ns, name, true)
	req.Param("resourceVersion", resourceVersion)
	watchIface, err := req.Watch()
	if err != nil {
		glog.Errorf("initiating the raw watch for %s/%s: %s", ns, name, err)
		return nil, err
	}
	return watch.Filter(watchIface, watchFilterer(s, ns)), nil
}

func watchFilterer(s *store, ns string) func(watch.Event) (watch.Event, bool) {
	return func(in watch.Event) (watch.Event, bool) {
		encodedBytes, err := runtime.Encode(s.codec, in.Object)
		if err != nil {
			glog.Errorf("encoding watch event object %#v: %s", in.Object, err)
			return watch.Event{}, false
		}
		finalObj := s.singularShell("", "")
		if err := s.decode(encodedBytes, finalObj); err != nil {
			glog.Errorf(
				"decoding watch event bytes %s: %s", string(encodedBytes), err,
			)
			return watch.Event{}, false
		}
		if !s.hasNamespace {
			if err := removeNamespace(finalObj); err != nil {
				glog.Errorf("removing namespace from %#v: %s", finalObj, err)
				return watch.Event{}, false
			}
		}
		return watch.Event{
			Type:   in.Type,
			Object: finalObj,
		}, true
	}
}

// WatchList implements storage.Interface.WatchList.
func (s *store) WatchList(
	ctx context.Context,
	key string,
	resourceVersion string,
	p storage.SelectionPredicate,
) (watch.Interface, error) {
	ns, _, err := s.decodeKey(key)
	if err != nil {
		glog.Errorf("decoding key %s: %s", key, err)
		return nil, err
	}
	req := s.buildRequest("GET", ns, "", true).Param("resourceVersion", resourceVersion)
	watchIface, err := req.Watch()
	if err != nil {
		glog.Errorf("initiating the raw watch for %s: %s", ns, err)
		return nil, err
	}
	return watch.Filter(watchIface, watchFilterer(s, ns)), nil
}

func (s *store) getState(
	res restclient.Result,
	key string,
	ignoreNotFound bool,
) (*objState, error) {
	state := &objState{
		obj:  s.singularShell("", ""),
		meta: &storage.ResponseMeta{},
	}
	var statusCode int
	res.StatusCode(&statusCode)
	if statusCode == http.StatusNotFound {
		if !ignoreNotFound {
			return nil, storage.NewKeyNotFoundError(key, 0)
		}
		if err := runtime.SetZeroValue(state.obj); err != nil {
			glog.Errorf(
				"setting object of type %s to zero value: %s",
				reflect.TypeOf(state.obj), err,
			)
			return nil, err
		}
	} else {
		var unknown runtime.Unknown
		if err := res.Into(&unknown); err != nil {
			glog.Errorf("reading response: %s", err)
			return nil, err
		}
		if err := s.decode(unknown.Raw, state.obj); err != nil {
			glog.Errorf("decoding response body %s: %s", string(unknown.Raw), err)
			return nil, err
		}
		if !s.hasNamespace {
			err := accessor.SetNamespace(state.obj, "")
			if err != nil {
				glog.Errorf("setting namespace to \"\": %s", err)
				return nil, err
			}
			// Replace the raw bytes with bytes that have the namespace set to ""
			unknown.Raw, err = runtime.Encode(s.codec, state.obj)
			if err != nil {
				glog.Errorf("encoding %#v: %s", state.obj, err)
			}
		}
		var err error
		state.rev, err = s.versioner.ObjectResourceVersion(state.obj)
		if err != nil {
			glog.Errorf("getting resource version for object %#v: %s", state.obj, err)
			return nil, err
		}
		state.meta.ResourceVersion = state.rev
		state.data = unknown.Raw
	}
	return state, nil
}

func (s *store) getStateFromObject(obj runtime.Object) (*objState, error) {
	versioner := s.versioner
	state := &objState{
		obj:  obj,
		meta: &storage.ResponseMeta{},
	}
	var err error
	state.rev, err = versioner.ObjectResourceVersion(obj)
	if err != nil {
		return nil,
			fmt.Errorf("couldn't get resource version: %s", err)
	}
	state.meta.ResourceVersion = state.rev
	state.data, err = runtime.Encode(s.codec, obj)
	if err != nil {
		glog.Errorf("encoding %#v: %s", obj, err)
		return nil, err
	}
	return state, nil
}

func (s *store) updateState(
	st *objState,
	userUpdate storage.UpdateFunc,
) (runtime.Object, error) {
	ret, _, err := userUpdate(st.obj, *st.meta)
	if err != nil {
		glog.Errorf("applying user update: %s", err)
		return nil, err
	}
	return ret, nil
}

func (s *store) decode(
	value []byte,
	objPtr runtime.Object,
) error {
	if _, err := conversion.EnforcePtr(objPtr); err != nil {
		glog.Errorf("converting output object to pointer: %s", err)
		return err
	}
	_, _, err := s.codec.Decode(value, nil, objPtr)
	if err != nil {
		glog.Errorf("decoding %s: %s", string(value), err)
	}
	return err
}

func checkPreconditions(
	key string,
	preconditions *storage.Preconditions,
	out runtime.Object,
) error {
	if preconditions == nil {
		return nil
	}
	objMeta, err := v1.ObjectMetaFor(out)
	if err != nil {
		return storage.NewInternalErrorf(
			"can't enforce preconditions %v on un-introspectable object %v, got error: %v",
			*preconditions, out, err,
		)
	}
	if preconditions.UID != nil && *preconditions.UID != objMeta.UID {
		errMsg := fmt.Sprintf(
			"Precondition failed: UID in precondition: %v, UID in object meta: %v",
			*preconditions.UID, objMeta.UID,
		)
		return storage.NewInvalidObjError(key, errMsg)
	}
	return nil
}

func removeNamespace(obj runtime.Object) error {
	if err := accessor.SetNamespace(obj, ""); err != nil {
		glog.Errorf("removing namespace from %#v: %s", obj, err)
		return err
	}
	return nil
}

func (s *store) buildRequest(method, ns, name string, watch bool) *restclient.Request {
	args := []string{
		"apis",
		groupName,
		tprVersion,
	}
	if watch {
		args = append(args, "watch")
	}
	args = append(args, []string{
		"namespaces",
		ns,
		s.singularKind.URLName(),
	}...)
	if name != "" {
		args = append(args, name)
	}
	method = strings.ToUpper(method)
	var req *restclient.Request
	switch method {
	case "GET":
		req = s.cl.Get()
	case "POST":
		req = s.cl.Post()
	case "PUT":
		req = s.cl.Put()
	case "DELETE":
		req = s.cl.Delete()
	}
	return req.AbsPath(args...)
}

func (s *store) exec(
	req *restclient.Request,
	out runtime.Object,
) (restclient.Result, int, error) {
	res := req.Do()
	var statusCode int
	res.StatusCode(&statusCode)
	if out != nil {
		var unknown runtime.Unknown
		if err := res.Into(&unknown); err != nil {
			glog.Errorf("reading response: %s", err)
			return res, statusCode, err
		}
		if err := s.decode(unknown.Raw, out); err != nil {
			glog.Errorf("decoding response body %s: %s", string(unknown.Raw), err)
			return res, statusCode, err
		}
	}
	return res, statusCode, nil
}
