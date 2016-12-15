package v1alpha1

import (
	"k8s.io/kubernetes/pkg/runtime/schema"
)

const (
	// GroupNameString is the name of the group
	GroupNameString = "catalog.k8s.io"
	// VersionString is the version of the group
	VersionString = "v1alpha1"
)

// GroupVersion is the official schema GroupVersion for this API server
var GroupVersion = schema.GroupVersion{Group: GroupNameString, Version: VersionString}
