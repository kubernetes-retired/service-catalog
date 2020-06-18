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

package webhookutil_test

import (
	"fmt"
	"testing"

	scfeatures "github.com/kubernetes-sigs/service-catalog/pkg/features"
	"github.com/kubernetes-sigs/service-catalog/pkg/webhookutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	authenticationv1 "k8s.io/api/authentication/v1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestUserInfoFromRequestSetUserInfoIfOriginatingIdentityIsEnabled(t *testing.T) {
	// given
	req := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			UserInfo: authenticationv1.UserInfo{
				Username: "username@domain.com",
				UID:      "123-123-123-123-123",
				Groups:   []string{"foo", "bar"},
				Extra:    nil,
			},
		},
	}

	// when
	gotUserInfo := webhookutil.UserInfoFromRequest(req)

	// then
	assert.EqualValues(t, req.UserInfo.Username, gotUserInfo.Username)
	assert.EqualValues(t, req.UserInfo.UID, gotUserInfo.UID)
	assert.EqualValues(t, req.UserInfo.Groups, gotUserInfo.Groups)
	assert.EqualValues(t, req.UserInfo.Extra, gotUserInfo.Extra)
}

func TestUserInfoFromRequestReturnNiltUserInfoIfOriginatingIdentityIsDisabled(t *testing.T) {
	err := utilfeature.DefaultMutableFeatureGate.Set(fmt.Sprintf("%v=false", scfeatures.OriginatingIdentity))
	require.NoError(t, err, "cannot disable OriginatingIdentity feature")
	// restore default state
	defer utilfeature.DefaultMutableFeatureGate.Set(fmt.Sprintf("%v=true", scfeatures.OriginatingIdentity))

	// given
	req := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			UserInfo: authenticationv1.UserInfo{
				Username: "username@domain.com",
				UID:      "123-123-123-123-123",
				Groups:   []string{"foo", "bar"},
				Extra:    nil,
			},
		},
	}

	// when
	gotUserInfo := webhookutil.UserInfoFromRequest(req)

	// then
	assert.Nil(t, gotUserInfo)
}
