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
	"strconv"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/storage"
	restclient "k8s.io/client-go/rest"
)

type versioner struct {
	codec        runtime.Codec
	singularKind Kind
	listKind     Kind
	checkObject  func(runtime.Object) error
	restClient   restclient.Interface
	defaultNS    string
}

// UpdateObject sets storage metadata into an API object. Returns an error if the object
// cannot be updated correctly. May return nil if the requested object does not need metadata
// from database.
func (v *versioner) UpdateObject(obj runtime.Object, resourceVersion uint64) error {
	if err := accessor.SetResourceVersion(obj, strconv.Itoa(int(resourceVersion))); err != nil {
		glog.Errorf("setting resource version (%s)", err)
		return err
	}
	name, err := accessor.Name(obj)
	if err != nil {
		glog.Errorf("getting name of the object (%s)", err)
		return err
	}
	// the Namespace function may return a nil error and an empty namespace. if it returns
	// a non-nil error and/or an empty namespace, use the default namespace
	ns, err := accessor.Namespace(obj)
	if err != nil || ns == "" {
		ns = v.defaultNS
	}

	data, err := runtime.Encode(v.codec, obj)
	if err != nil {
		glog.Errorf("encoding obj (%s)", err)
		return err
	}
	req := v.restClient.Put().AbsPath(
		"apis",
		servicecatalog.GroupName,
		tprVersion,
		"namespaces",
		ns,
		v.singularKind.URLName(),
		name,
	).Body(data)

	if err := req.Do().Error(); err != nil {
		glog.Errorf("error updating object %s (%s)", name, err)
		return err
	}
	return nil
}

func updateState(v storage.Versioner, st *objState, userUpdate storage.UpdateFunc) (runtime.Object, uint64, error) {
	ret, ttlPtr, err := userUpdate(st.obj, *st.meta)
	if err != nil {
		glog.Errorf("user update (%s)", err)
		return nil, 0, err
	}

	version, err := v.ObjectResourceVersion(ret)
	if err != nil {
		glog.Errorf("getting resource version (%s)", err)
		return nil, 0, err
	}
	if version != 0 {
		// We cannot store object with resourceVersion. We need to reset it.
		if err := v.UpdateObject(ret, 0); err != nil {
			glog.Errorf("updating object (%s)", err)
			return nil, 0, err
		}
	}
	var ttl uint64
	if ttlPtr != nil {
		ttl = *ttlPtr
	}
	return ret, ttl, nil
}

// UpdateList receives a list object and ranges over constituent objects,
// calling UpdateObject for each. Contrasted with the ETCD-based implementation
// of UpdateList, here in this TPR-based implementation, we cannot update all
// the objects within the list atomically. If any one object update fails, the
// failure is logged and all remaining updates are aborted, but completed
// updates are not rolled back.
func (v *versioner) UpdateList(listObj runtime.Object, listResourceVersion uint64) error {
	return meta.EachListItem(listObj, func(obj runtime.Object) error {
		objResourceVersion, err := v.ObjectResourceVersion(obj)
		if err != nil {
			glog.Errorf("error getting resource version from %#v; aborting further updates (%s)", listObj, err)
			return err
		}
		if err := v.UpdateObject(obj, objResourceVersion); err != nil {
			glog.Errorf("error updating list object; aborting further updates (%s)", err)
			return err
		}
		return nil
	})
}

// ObjectResourceVersion returns the resource version (for persistence) of the specified object.
// Should return an error if the specified object does not have a persistable version.
func (v *versioner) ObjectResourceVersion(obj runtime.Object) (uint64, error) {
	vsnStr, err := GetAccessor().ResourceVersion(obj)
	if err != nil {
		return 0, err
	}
	ret, err := strconv.ParseUint(vsnStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return ret, nil
}
