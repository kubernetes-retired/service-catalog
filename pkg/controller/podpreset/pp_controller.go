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

package podpreset

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/apimachinery/pkg/util/wait"

	settingsinformers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions/settings/v1alpha1"
	settingslisters "github.com/kubernetes-incubator/service-catalog/pkg/client/listers_generated/settings/v1alpha1"
)

const (
	podPresetInitializerName = "podpreset.initializer.k8s.io"
)

// Controller implements PodPreset initializer.
type Controller struct {
	kubeClient kubernetes.Interface
	recorder   record.EventRecorder

	// to allow injection for testing.
	initFn func(key string) error

	podInformer     cache.SharedIndexInformer
	podLister       corelisters.PodLister
	podListerSynced cache.InformerSynced

	podpresetInformer     settingsinformers.PodPresetInformer
	podpresetLister       settingslisters.PodPresetLister
	podpresetListerSynced cache.InformerSynced

	podQueue workqueue.RateLimitingInterface
}

// NewController returns a PodPreset Controller instance.
func NewController(
	kubeClient kubernetes.Interface,
	recorder record.EventRecorder,
	podInformer cache.SharedIndexInformer,
	podpresetInformer settingsinformers.PodPresetInformer,
) (*Controller, error) {

	c := &Controller{
		kubeClient:        kubeClient,
		podpresetInformer: podpresetInformer,
		recorder:          recorder,
		podQueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "podpreset"),
	}

	c.podpresetLister = podpresetInformer.Lister()
	c.podpresetListerSynced = podpresetInformer.Informer().HasSynced

	c.podInformer = podInformer
	c.podInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.addPod,
			UpdateFunc: c.updatePod,
		},
	)
	c.podLister = corelisters.NewPodLister(c.podInformer.GetIndexer())
	c.podListerSynced = c.podInformer.HasSynced
	c.initFn = c.initPod

	return c, nil
}

func (c *Controller) addPod(obj interface{}) {
	pod := obj.(*v1.Pod)
	glog.V(5).Infof("new Pod: %s with meta: %+v received", pod.GetName(), pod.ObjectMeta)
	if needsInitialization(pod) {
		glog.Infof("found an uninitialized pod: %+v", pod.Name)
		if key, err := cache.MetaNamespaceKeyFunc(obj); err == nil {
			c.podQueue.Add(key)
		}
	}
}

func (c *Controller) updatePod(old, new interface{}) {
	pod := new.(*v1.Pod)
	glog.V(5).Infof("Pod: %s with meta: %+v update received", pod.GetName(), pod.ObjectMeta)
	if needsInitialization(pod) {
		glog.Infof("found an existing uninitialized pod: %s", pod.GetName())
		if key, err := cache.MetaNamespaceKeyFunc(new); err == nil {
			c.podQueue.Add(key)
		}
	}
}

// Run runs the controller until the given stop channel can be read from.
func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	defer func() {
		c.podQueue.ShutDown()
	}()
	glog.Info("Starting podpreset initializer")
	defer glog.Infof("Shutting down podpreset initializer")

	go c.podInformer.Run(stopCh)

	// Wait for all caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.podListerSynced) {
		glog.Error(fmt.Errorf("Timed out waiting for pod cache to sync"))
		return
	}

	// Wait for all caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.podpresetListerSynced) {
		glog.Error(fmt.Errorf("Timed out waiting for pod cache to sync"))
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	// wait unitl we are told to stop
	<-stopCh
}

func (c *Controller) runWorker() {
	for c.processNext() {
	}
}

func (c *Controller) processNext() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.podQueue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.podQueue.Done(key)
	// Initialize the incoming Pod
	err := c.initFn(key.(string))
	c.handleErr(err, key)
	return true
}

func (c *Controller) initPod(key string) error {
	glog.V(5).Infof("got key: %v", key)

	ns, podName, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("failed to parse key: %v :%v", key, err)
	}

	pod, err := c.podLister.Pods(ns).Get(podName)
	if err != nil {
		return fmt.Errorf("failed to retrieve pod %s in ns: %v : %v", podName, ns, err)
	}

	err = admit(pod, c.podpresetLister, c.recorder)
	if err != nil {
		return fmt.Errorf("failure in applying podpreset on pod %s error: %v", podName, err)
	}

	markInitializationDone(pod)

	if _, err = c.kubeClient.CoreV1().Pods(ns).Update(pod); err != nil {
		return fmt.Errorf("failed to update pod : %v", err)
	}

	return nil
}

// markInitializationDone removes the PodPreset initializer from the Pod's
// pending initializer list. And if it is the only initializer in the pending
// list, then resets the Initializers field to nil mark the initialization done.
func markInitializationDone(pod *v1.Pod) {
	pendingInitializers := pod.GetInitializers().Pending
	if len(pendingInitializers) == 1 {
		pod.ObjectMeta.Initializers = nil
	} else {
		pod.ObjectMeta.Initializers.Pending = pod.ObjectMeta.Initializers.Pending[1:]
	}
}

// isPodUninitialized determines if Pod is waiting for PodPreset initialization.
func needsInitialization(pod *v1.Pod) bool {
	initializers := pod.ObjectMeta.GetInitializers()
	if initializers != nil && len(initializers.Pending) > 0 &&
		initializers.Pending[0].Name == podPresetInitializerName {
		return true
	}
	glog.V(4).Infof("pod %s with initalizers %+v does not need initialization", pod.GetName(), initializers)
	return false
}

func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.podQueue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.podQueue.NumRequeues(key) < 5 {
		glog.Infof("Error processing pod %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.podQueue.AddRateLimited(key)
		return
	}

	c.podQueue.Forget(key)
	glog.Errorf("Dropping pod %q out of the queue: %v", key, err)
}
