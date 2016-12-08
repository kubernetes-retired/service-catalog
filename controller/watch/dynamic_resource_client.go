package watch

import (
	"k8s.io/client-go/1.5/dynamic"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/watch"
)

type dynamicResourceClient interface {
	Watch(*v1.ListOptions) (watch.Interface, error)
}

type realDynamicResourceClient dynamic.ResourceClient

func newRealDynamicResourceClient(rc *dynamic.ResourceClient) *realDynamicResourceClient {
	ret := realDynamicResourceClient(*rc)
	return &ret
}

func (r *realDynamicResourceClient) Watch(opts *v1.ListOptions) (watch.Interface, error) {
	return r.Watch(opts)
}
