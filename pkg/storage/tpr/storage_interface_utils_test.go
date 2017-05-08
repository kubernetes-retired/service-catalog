package tpr

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/rest/core/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/storage"
)

func runWatchListTest(keyer Keyer, fakeCl *fake.RESTClient, iface storage.Interface, obj runtime.Object) error {
	const timeout = 1 * time.Second
	key, err := keyer.Key(request.NewContext(), name)
	resourceVsn := "1234"
	predicate := storage.SelectionPredicate{}
	sendObjErrCh := make(chan error)
	go func() {
		defer fakeCl.Watcher.Close()
		if err := fakeCl.Watcher.SendObject(watch.Added, obj, 1*time.Second); err != nil {
			sendObjErrCh <- err
			return
		}
	}()
	watchIface, err := iface.WatchList(context.Background(), key, resourceVsn, predicate)
	if err != nil {
		return err
	}
	if watchIface == nil {
		return errors.New("expected non-nil watch interface")
	}
	defer watchIface.Stop()
	ch := watchIface.ResultChan()
	select {
	case err := <-sendObjErrCh:
		return fmt.Errorf("error sending object (%s)", err)
	case evt, ok := <-ch:
		if !ok {
			return errors.New("watch channel was closed")
		}
		if evt.Type != watch.Added {
			return errors.New("event type was not ADDED")
		}
		if err := deepCompare("expected", obj, "actual", evt.Object); err != nil {
			return fmt.Errorf("received objects aren't the same (%s)", err)
		}
	case <-time.After(timeout):
		return fmt.Errorf("didn't receive after %s", timeout)
	}
	select {
	case _, ok := <-ch:
		if ok {
			return errors.New("watch channel was not closed")
		}
	case <-time.After(timeout):
		return fmt.Errorf("watch channel didn't receive after %s", timeout)
	}
	return nil
}
