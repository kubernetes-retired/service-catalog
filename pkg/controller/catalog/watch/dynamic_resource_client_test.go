package watch

import (
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/watch"
)

// fakeDynamicResourceClient is a dynamicResourceClient implementation intended for unit tests.
// it contains a watch.Interface and error that it will always return in its Watch func.
//
// The watch.Interface will usually be a (k8s.io/client-go/1.5/pkg/watch).FakeWatcher, which will
// be driven (i.e. 'Add' func called) from unit tests
type fakeDynamicResourceClient struct {
	watchRet    watch.Interface
	watchRetErr error
}

func (f *fakeDynamicResourceClient) Watch(*v1.ListOptions) (watch.Interface, error) {
	return f.watchRet, f.watchRetErr
}
