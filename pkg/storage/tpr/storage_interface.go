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
	"errors"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"k8s.io/kubernetes/pkg/api/meta"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
	"k8s.io/kubernetes/pkg/conversion"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/storage"
	"k8s.io/kubernetes/pkg/storage/storagebackend/factory"
	"k8s.io/kubernetes/pkg/watch"
)

var (
	errNotImplemented = errors.New("not implemented for third party resources")
)

type storageInterface struct {
	hasNamespace     bool
	codec            runtime.Codec
	defaultNamespace string
	cl               clientset.Interface
	singularKind     Kind
	singularShell    func(string, string) runtime.Object
	listKind         Kind
	listShell        func() runtime.Object
	checkObject      func(runtime.Object) error
	decodeKey        func(string) (string, string, error)
}

// NewStorageInterface creates a new TPR-based storage.Interface implementation
func NewStorageInterface(opts Options) (storage.Interface, factory.DestroyFunc) {
	return &storageInterface{
		hasNamespace:     opts.HasNamespace,
		codec:            opts.RESTOptions.StorageConfig.Codec,
		defaultNamespace: opts.DefaultNamespace,
		cl:               opts.Client,
		singularKind:     opts.SingularKind,
		singularShell:    opts.NewSingularFunc,
		listKind:         opts.ListKind,
		listShell:        opts.NewListFunc,
		checkObject:      opts.CheckObjectFunc,
		decodeKey:        opts.Keyer.NamespaceAndNameFromKey,
	}, opts.DestroyFunc
}

// Versioned returns the versioned associated with this interface
func (t *storageInterface) Versioner() storage.Versioner {
	return &storageVersioner{
		codec:        t.codec,
		singularKind: t.singularKind,
		listKind:     t.listKind,
		checkObject:  t.checkObject,
		defaultNS:    t.defaultNamespace,
		cl:           t.cl,
	}
}

// Create adds a new object at a key unless it already exists. 'ttl' is time-to-live
// in seconds (0 means forever). If no error is returned and out is not nil, out will be
// set to the read value from database.
func (t *storageInterface) Create(
	ctx context.Context,
	key string,
	obj,
	out runtime.Object,
	ttl uint64,
) error {

	ns, _, err := t.decodeKey(key)
	if err != nil {
		glog.Errorf("decoding key %s (%s)", key, err)
		return err
	}

	data, err := runtime.Encode(t.codec, obj)
	if err != nil {
		return err
	}

	req := t.cl.Core().RESTClient().Post().AbsPath(
		"apis",
		groupName,
		tprVersion,
		"namespaces",
		ns,
		t.singularKind.URLName(),
	).Body(data)

	var unknown runtime.Unknown
	if err := req.Do().Into(&unknown); err != nil {
		glog.Errorf("decoding response (%s)", err)
		return err
	}

	if err := decode(t.codec, nil, unknown.Raw, out); err != nil {
		return err
	}
	return nil
}

// Delete removes the specified key and returns the value that existed at that spot.
// If key didn't exist, it will return NotFound storage error.
//
// In this implementation, Delete will not write the deleted object back to out
func (t *storageInterface) Delete(
	ctx context.Context,
	key string,
	out runtime.Object,
	preconditions *storage.Preconditions,
) error {
	ns, name, err := t.decodeKey(key)
	if err != nil {
		glog.Errorf("decoding key %s (%s)", key, err)
		return err
	}

	req := t.cl.Core().RESTClient().Delete().AbsPath(
		"apis",
		groupName,
		tprVersion,
		"namespaces",
		ns,
		t.singularKind.URLName(),
		name,
	)
	if err := req.Do().Error(); err != nil {
		glog.Errorf("error deleting (%s)", err)
		return err
	}

	return nil
}

// Watch begins watching the specified key. Events are decoded into API objects,
// and any items selected by 'p' are sent down to returned watch.Interface.
// resourceVersion may be used to specify what version to begin watching,
// which should be the current resourceVersion, and no longer rv+1
// (e.g. reconnecting without missing any updates).
// If resource version is "0", this interface will get current object at given key
// and send it in an "ADDED" event, before watch starts.
func (t *storageInterface) Watch(
	ctx context.Context,
	key string,
	resourceVersion string,
	p storage.SelectionPredicate,
) (watch.Interface, error) {
	ns, name, err := t.decodeKey(key)
	if err != nil {
		return nil, err
	}

	req := t.cl.Core().RESTClient().Get().AbsPath(
		"apis",
		groupName,
		tprVersion,
		"watch",
		"namespaces",
		ns,
		t.singularKind.URLName(),
		name,
	).Param("resourceVersion", resourceVersion)
	watchIface, err := req.Watch()
	if err != nil {
		glog.Errorf("initiating the raw watch (%s)", err)
		return nil, err
	}
	filteredIFace := watch.Filter(watchIface, watchFilterer(t, ns))
	return filteredIFace, nil
}

func watchFilterer(t *storageInterface, ns string) func(watch.Event) (watch.Event, bool) {
	return func(in watch.Event) (watch.Event, bool) {
		encodedBytes, err := runtime.Encode(t.codec, in.Object)
		if err != nil {
			glog.Errorf("couldn't encode watch event object (%s)", err)
			return watch.Event{}, false
		}
		finalObj := t.singularShell("", "")
		if err := decode(t.codec, nil, encodedBytes, finalObj); err != nil {
			glog.Errorf("couldn't decode watch event bytes (%s)", err)
			return watch.Event{}, false
		}
		if !t.hasNamespace {
			if err := removeNamespace(finalObj); err != nil {
				glog.Errorf("couldn't remove namespace from %#v (%s)", finalObj, err)
				return watch.Event{}, false
			}
		}
		return watch.Event{
			Type:   in.Type,
			Object: finalObj,
		}, true
	}
}

// WatchList begins watching the specified key's items. Items are decoded into API
// objects and any item selected by 'p' are sent down to returned watch.Interface.
// resourceVersion may be used to specify what version to begin watching,
// which should be the current resourceVersion, and no longer rv+1
// (e.g. reconnecting without missing any updates).
// If resource version is "0", this interface will list current objects directory defined by key
// and send them in "ADDED" events, before watch starts.
func (t *storageInterface) WatchList(
	ctx context.Context,
	key string,
	resourceVersion string,
	p storage.SelectionPredicate,
) (watch.Interface, error) {
	ns, _, err := t.decodeKey(key)
	if err != nil {
		return nil, err
	}

	req := t.cl.Core().RESTClient().Get().AbsPath(
		"apis",
		groupName,
		tprVersion,
		"watch",
		"namespaces",
		ns,
		t.singularKind.URLName(),
	).Param("resourceVersion", resourceVersion)

	watchIface, err := req.Watch()
	if err != nil {
		glog.Errorf("initiating the raw watch (%s)", err)
		return nil, err
	}
	return watch.Filter(watchIface, watchFilterer(t, ns)), nil
}

// Get unmarshals json found at key into objPtr. On a not found error, will either
// return a zero object of the requested type, or an error, depending on ignoreNotFound.
// Treats empty responses and nil response nodes exactly like a not found error.
// The returned contents may be delayed, but it is guaranteed that they will
// be have at least 'resourceVersion'.
func (t *storageInterface) Get(
	ctx context.Context,
	key string,
	resourceVersion string,
	objPtr runtime.Object,
	ignoreNotFound bool,
) error {
	ns, name, err := t.decodeKey(key)
	if err != nil {
		glog.Errorf("decoding key %s (%s)", key, err)
		return err
	}
	req := t.cl.Core().RESTClient().Get().AbsPath(
		"apis",
		groupName,
		tprVersion,
		"namespaces",
		ns,
		t.singularKind.URLName(),
		name,
	)

	var unknown runtime.Unknown
	if err := req.Do().Into(&unknown); err != nil {
		glog.Errorf("decoding response (%s)", err)
		return err
	}

	if err := decode(t.codec, nil, unknown.Raw, objPtr); err != nil {
		return nil
	}
	if !t.hasNamespace {
		if err := removeNamespace(objPtr); err != nil {
			glog.Errorf("removing namespace from %#v (%s)", objPtr, err)
			return err
		}
	}
	return nil
}

// GetToList unmarshals json found at key and opaque it into *List api object
// (an object that satisfies the runtime.IsList definition).
// The returned contents may be delayed, but it is guaranteed that they will
// be have at least 'resourceVersion'.
func (t *storageInterface) GetToList(
	ctx context.Context,
	key string,
	resourceVersion string,
	p storage.SelectionPredicate,
	listObj runtime.Object,
) error {
	return t.List(ctx, key, resourceVersion, p, listObj)
}

// List unmarshalls jsons found at directory defined by key and opaque them
// into *List api object (an object that satisfies runtime.IsList definition).
// The returned contents may be delayed, but it is guaranteed that they will
// be have at least 'resourceVersion'.
func (t *storageInterface) List(
	ctx context.Context,
	key string,
	resourceVersion string,
	p storage.SelectionPredicate,
	listObj runtime.Object,
) error {
	ns, _, err := t.decodeKey(key)
	if err != nil {
		glog.Errorf("decoding %s (%s)", key, err)
		return err
	}

	req := t.cl.Core().RESTClient().Get().AbsPath(
		"apis",
		groupName,
		tprVersion,
		"namespaces",
		ns,
		t.singularKind.URLName(),
	)

	var unknown runtime.Unknown
	if err := req.Do().Into(&unknown); err != nil {
		glog.Errorf("doing request (%s)", err)
		return err
	}

	if err := decode(t.codec, nil, unknown.Raw, listObj); err != nil {
		return err
	}

	if !t.hasNamespace {
		if err := meta.EachListItem(listObj, removeNamespace); err != nil {
			glog.Errorf("removing namespace from all items in list (%s)", err)
			return err
		}
	}
	return nil
}

// GuaranteedUpdate keeps calling 'tryUpdate()' to update key 'key' (of type 'ptrToType')
// retrying the update until success if there is index conflict.
// Note that object passed to tryUpdate may change across invocations of tryUpdate() if
// other writers are simultaneously updating it, so tryUpdate() needs to take into account
// the current contents of the object when deciding how the update object should look.
// If the key doesn't exist, it will return NotFound storage error if ignoreNotFound=false
// or zero value in 'ptrToType' parameter otherwise.
// If the object to update has the same value as previous, it won't do any update
// but will return the object in 'ptrToType' parameter.
// If 'suggestion' can contain zero or one element - in such case this can be used as
// a suggestion about the current version of the object to avoid read operation from
// storage to get it.
//
// Example:
//
// s := /* implementation of Interface */
// err := s.GuaranteedUpdate(
//     "myKey", &MyType{}, true,
//     func(input runtime.Object, res ResponseMeta) (runtime.Object, *uint64, error) {
//       // Before each incovation of the user defined function, "input" is reset to
//       // current contents for "myKey" in database.
//       curr := input.(*MyType)  // Guaranteed to succeed.
//
//       // Make the modification
//       curr.Counter++
//
//       // Return the modified object - return an error to stop iterating. Return
//       // a uint64 to alter the TTL on the object, or nil to keep it the same value.
//       return cur, nil, nil
//    }
// })
func (t *storageInterface) GuaranteedUpdate(
	ctx context.Context,
	key string,
	out runtime.Object,
	ignoreNotFound bool,
	precondtions *storage.Preconditions,
	tryUpdate storage.UpdateFunc,
	suggestion ...runtime.Object,
) error {
	var origState *objState
	if len(suggestion) == 1 && suggestion[0] != nil {
		s, err := getStateFromObject(t, suggestion[0])
		if err != nil {
			glog.Errorf("getting state from suggested object (%s)", err)
			return err
		}
		origState = s
	} else {
		if err := t.Get(ctx, key, "", out, false); err != nil {
			glog.Errorf("getting initial object (%s)", err)
			return err
		}
		s, err := getStateFromObject(t, out)
		if err != nil {
			glog.Errorf("getting state from fetched object (%s)", err)
			return err
		}
		origState = s
	}
	for {
		ret, _, err := updateState(t.Versioner(), origState, tryUpdate)
		if err != nil {
			glog.Errorf("updating the state (%s)", err)
			return err
		}
		data, err := runtime.Encode(t.codec, ret)
		if err != nil {
			glog.Errorf("encoding return object (%s)", err)
			return err
		}
		return decode(t.codec, t.Versioner(), data, out)
	}
}

func decode(
	codec runtime.Codec,
	versioner storage.Versioner,
	value []byte,
	objPtr runtime.Object,
) error {
	if _, err := conversion.EnforcePtr(objPtr); err != nil {
		panic("unable to convert output object to pointer")
	}
	_, _, err := codec.Decode(value, nil, objPtr)
	if err != nil {
		return err
	}
	return nil
}

func removeNamespace(obj runtime.Object) error {
	if err := accessor.SetNamespace(obj, ""); err != nil {
		glog.Errorf("removing namespace from %#v (%s)", obj, err)
		return err
	}
	return nil
}
