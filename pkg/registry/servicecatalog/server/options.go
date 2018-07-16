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

package server

import (
	"fmt"
)

type errUnsupportedStorageType struct {
	t StorageType
}

func (e errUnsupportedStorageType) Error() string {
	return fmt.Sprintf("unsupported storage type %s", e.t)
}

// StorageType represents the type of storage a storage interface should use
type StorageType string

// StorageTypeFromString converts s to a valid StorageType. Returns StorageType("") and a non-nil
// error if s names an invalid or unsupported storage type
func StorageTypeFromString(s string) (StorageType, error) {
	switch s {
	case StorageTypeEtcd.String():
		return StorageTypeEtcd, nil
	default:
		return StorageType(""), errUnsupportedStorageType{t: StorageType(s)}
	}
}

func (s StorageType) String() string {
	return string(s)
}

const (
	// StorageTypeEtcd indicates a storage interface should use etcd
	StorageTypeEtcd StorageType = "etcd"
)

// Options is the extension of a generic.RESTOptions struct, complete with service-catalog
// specific things
type Options struct {
	storageType StorageType
}

// NewOptions returns a new Options with the given parameters
func NewOptions(
	sType StorageType,
) *Options {
	return &Options{
		storageType: sType,
	}
}

// StorageType returns the storage type the rest server should use, or an error if an unsupported
// storage type is indicated
func (o Options) StorageType() (StorageType, error) {
	switch o.storageType {
	case StorageTypeEtcd:
		return o.storageType, nil
	default:
		return StorageType(""), errUnsupportedStorageType{t: o.storageType}
	}
}
