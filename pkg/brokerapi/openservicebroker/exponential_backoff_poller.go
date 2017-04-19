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

package openservicebroker

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
)

const (
	// Max specifies the maximum duration we're going to back off up to
	max = 1 * time.Hour
)

// ExponentialBackoffPoller implements a (./pkg/brokerapi).Poller with exponential backoff.
// TODO(vaikas): Add cancel, or use queues.
type ExponentialBackoffPoller struct {
	d  time.Duration
	cb *brokerapi.PollerCallback
}

// NewPoller creates a Poller with specified backoff parameters for polling
func NewPoller() brokerapi.Poller {
	return &ExponentialBackoffPoller{d: time.Second}
}

// CreatePoller creates a new poller for a given broker and instance
func (p *ExponentialBackoffPoller) CreatePoller(client brokerapi.BrokerClient, instance *v1alpha1.Instance, req *brokerapi.LastOperationRequest, cb brokerapi.PollerCallback) error {
	go p.poll(client, instance, req, cb)
	return nil
}

func (p *ExponentialBackoffPoller) poll(client brokerapi.BrokerClient, instance *v1alpha1.Instance, req *brokerapi.LastOperationRequest, cb brokerapi.PollerCallback) {
	for {
		instanceID := instance.Spec.OSBGUID
		glog.V(3).Infof("Polling instance %q", instanceID)
		resp, err := client.PollServiceInstance(instanceID, req)
		if err != nil {
			s := fmt.Sprintf("PollServiceInstance failed for %q : %s", instanceID, err)
			glog.Warning(s)
		} else {
			if !cb(instance, resp) {
				glog.V(3).Infof("Stopping polling for instance %q", instanceID)
				return
			}
		}
		time.Sleep(p.d)
		p.d = 2 * p.d
		if p.d > max {
			p.d = max
		}
	}

}
