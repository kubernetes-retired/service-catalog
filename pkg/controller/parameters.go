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

package controller

import (
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func buildParameters(kubeClient kubernetes.Interface, namespace string, parametersFrom []v1alpha1.ParametersFromSource, parameters []v1alpha1.Parameter) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if parametersFrom != nil {
		for _, p := range parametersFrom {
			fps, err := fetchParametersFromSource(kubeClient, namespace, &p)
			if err != nil {
				return nil, err
			}
			for k, v := range fps {
				params[k] = v
			}
		}
	}
	if parameters != nil {
		for _, p := range parameters {
			v, err := fetchParameter(kubeClient, namespace, &p)
			if err != nil {
				return nil, err
			}
			params[p.Name] = v
		}
	}
	return params, nil
}

func fetchParametersFromSource(kubeClient kubernetes.Interface, namespace string, parametersFrom *v1alpha1.ParametersFromSource) (map[string]interface{}, error) {
	var params map[string]interface{}
	if parametersFrom.Value != nil {
		p, err := unmarshalRawParameters(parametersFrom.Value.Raw)
		if err != nil {
			return nil, err
		}
		params = p
	}
	if parametersFrom.SecretRef != nil {
		p, err := fetchSecretParameters(kubeClient, namespace, parametersFrom.SecretRef)
		if err != nil {
			return nil, err
		}
		params = p
	}
	if parametersFrom.SecretKeyRef != nil {
		data, err := fetchSecretKeyValue(kubeClient, namespace, parametersFrom.SecretKeyRef)
		if err != nil {
			return nil, err
		}
		p, err := unmarshalValue([]byte(data), v1alpha1.ValueTypeJSON)
		if err != nil {
			return nil, err
		}
		params = p.(map[string]interface{})

	}
	return params, nil
}

func fetchParameter(kubeClient kubernetes.Interface, namespace string, parameter *v1alpha1.Parameter) (interface{}, error) {
	if parameter.Value != "" {
		return unmarshalValue([]byte(parameter.Value), parameter.Type)
	}
	if parameter.ValueFrom != nil {
		source := parameter.ValueFrom
		if source.SecretKeyRef != nil {
			data, err := fetchSecretKeyValue(kubeClient, namespace, source.SecretKeyRef)
			if err != nil {
				return nil, err
			}
			return unmarshalValue([]byte(data), parameter.Type)
		}
	}
	return "", nil
}

func unmarshalRawParameters(in []byte) (map[string]interface{}, error) {
	parameters := make(map[string]interface{})
	if len(in) > 0 {
		if err := yaml.Unmarshal(in, &parameters); err != nil {
			return parameters, err
		}
	}
	return parameters, nil
}

// fetchSecretParameters requests and returns the contents of the given secret as a map
func fetchSecretParameters(kubeClient kubernetes.Interface, namespace string, secretRef *v1alpha1.SecretReference) (map[string]interface{}, error) {
	// TODO: add caching to avoid fetching the same secret many times?
	secret, err := kubeClient.CoreV1().Secrets(namespace).Get(secretRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{}, len(secret.Data))
	for k, v := range secret.Data {
		result[k], err = unmarshalValue(v, secretRef.Type)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// fetchSecretKeyValue requests and returns the contents of the given secret key as a string
func fetchSecretKeyValue(kubeClient kubernetes.Interface, namespace string, secretKeyRef *v1alpha1.SecretKeyReference) ([]byte, error) {
	secret, err := kubeClient.CoreV1().Secrets(namespace).Get(secretKeyRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return secret.Data[secretKeyRef.Key], nil
}

func unmarshalValue(in []byte, valueType v1alpha1.ParameterValueType) (interface{}, error) {
	switch valueType {
	case v1alpha1.ValueTypeString:
		return string(in), nil
	case v1alpha1.ValueTypeJSON:
		parameters := make(map[string]interface{})
		if err := json.Unmarshal(in, &parameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameters as JSON object: %v", err)
		}
		return parameters, nil
	default:
		return nil, fmt.Errorf("unsupported value type: %v", valueType)
	}
}
