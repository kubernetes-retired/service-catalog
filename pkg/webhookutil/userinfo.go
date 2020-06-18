/*
Copyright 2020 The Kubernetes Authors.

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

package webhookutil

import (
	sc "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scfeatures "github.com/kubernetes-sigs/service-catalog/pkg/features"

	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// UserInfoFromRequest returns user.Info from the request context if feature gate is enabled.
func UserInfoFromRequest(req admission.Request) *sc.UserInfo {
	// ensure that UserInfo will not be set if the feature gate is disabled
	if !utilfeature.DefaultFeatureGate.Enabled(scfeatures.OriginatingIdentity) {
		return nil
	}

	user := req.UserInfo
	scUserInfo := &sc.UserInfo{
		Username: user.Username,
		UID:      user.UID,
		Groups:   user.Groups,
	}
	if extra := user.Extra; len(extra) > 0 {
		scUserInfo.Extra = map[string]sc.ExtraValue{}
		for k, v := range extra {
			scUserInfo.Extra[k] = sc.ExtraValue(v)
		}
	}

	return scUserInfo
}
