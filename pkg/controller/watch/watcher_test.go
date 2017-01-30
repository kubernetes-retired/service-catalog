/*
Copyright 2016 The Kubernetes Authors.

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

package watch

import (
	"reflect"
	"testing"
	"time"

	"k8s.io/client-go/1.5/pkg/api/v1"
	extv1beta1 "k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.5/pkg/watch"
)

func TestThirdPartyWatcher(t *testing.T) {
	const evtChTimeout = 100 * time.Millisecond
	evtCh := make(chan watch.Event)
	cb := watchCallback(func(evt watch.Event) error {
		evtCh <- evt
		return nil
	})
	fakeWatcher := watch.NewFake()
	fakeRC := &fakeDynamicResourceClient{watchRet: fakeWatcher, watchRetErr: nil}
	go thirdPartyWatcher(fakeRC, cb)

	select {
	case evt := <-evtCh:
		t.Fatalf("received event (%s) before watcher sent any", evt)
	case <-time.After(evtChTimeout):
	}

	addedDepl := &extv1beta1.Deployment{ObjectMeta: v1.ObjectMeta{Name: "depl1"}}
	fakeWatcher.Add(addedDepl)

	select {
	case evt := <-evtCh:
		recvDepl, ok := evt.Object.(*extv1beta1.Deployment)
		if !ok {
			t.Fatalf("received event was not a deployment")
		}
		if !reflect.DeepEqual(addedDepl, recvDepl) {
			t.Fatal("received deployment was not equal to sent deployment")
		}
	case <-time.After(evtChTimeout):
		t.Fatalf("no event received after watcher sent (after %s)", evtChTimeout)
	}

	fakeWatcher.Stop()
	select {
	case evt := <-evtCh:
		t.Fatalf("received event (%s) after watcher stopped", evt)
	case <-time.After(evtChTimeout):
	}
}

func TestWatch(t *testing.T) {
	// TODO: implement
}
