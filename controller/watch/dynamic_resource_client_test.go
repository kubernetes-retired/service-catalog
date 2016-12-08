package watch

import (
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/watch"
)

type fakeDynamicResourceClient struct {
	watchRet    watch.Interface
	watchRetErr error
}

func (f *fakeDynamicResourceClient) Watch(*v1.ListOptions) (watch.Interface, error) {
	return f.watchRet, f.watchRetErr
}
