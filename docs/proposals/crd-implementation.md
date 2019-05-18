# Service Catalog migration to CRDs

## Implemented features

The API server custom features were migrated to the admission webhook approach.

The table converter API server feature was migrated to the [Additional Printer Columns](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/#additional-printer-columns) feature.

## Project changes

The architecture after the migration looks as follows:

<img src="/docs/images/current.png" width="75%" height="75%">

The Service Catalog resources are grouped under the `svcat` name. With that feature, you can list them with the `kubectl get svcat` command.

### Chart changes

The following files were added to the Service Catalog helm chart:

- **crds.yaml** - CustomResourceDefinitions for the Service Catalog resources
- **webhook-deployment.yaml** - webhook server deployment
- **webhook-register.yaml** - registration of the webhook server operations
- **webhook-service.yaml** - service for the webhook server
- **cleaner-job.yaml** - creates the `pre-delete` job which ensures that all CRDs and CRs are removed after helm release deletion

These files were deleted from the Service Catalog helm chart because they are not necessary anymore due to the API server removal:

- **apiregistration.yaml**
- **apiserver-deployment.yaml**
- **apiserver-service.yaml**
- **etcd-pvc.yaml**

All the webhook server configuration can be found under the `webhook` field in the chart's `values.yaml` file.

### Code changes

The following code changes were introduced to the Service Catalog:

- **pkg/webhookutil** - under this directory you can find the common logic used in the webhook server implementation
- **pkg/webhook/servicecatalog** - here you can find the webhook server logic for all of the Service Catalog resources
- **cmd/webhook/server** - the implementation of the webhook server where you can find the validation and mutation webhook registration
- **pkg/cleaner** - responsible for removing CRD and CR when removing a helm release
- **pkg/probe** - responsible for checking if webhook/controller is ready to use (readiness probe), especially if all required CRDs exist and have specific state

All of the webhook logic is covered by the unit tests. The API server tests were deleted.

## Implementation

Mutating and validating admission webhooks are registered in the chart's file **webhook-register.yaml**. For example, the registration of the Service Instances looks as follows:

```yaml
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: {{ template "fullname" . }}-webhook
  namespace: "{{ .Release.Namespace }}"
webhooks:
- name: mutating.serviceinstances.servicecatalog.k8s.io
  clientConfig:
    caBundle: {{ b64enc $ca.Cert }}
    service:
      name: {{ template "fullname" . }}-webhook
      namespace: "{{ .Release.Namespace }}"
      path: "/mutating-serviceinstances"
  failurePolicy: Fail
  rules:
  - operations: [ "CREATE", "UPDATE" ]
    apiGroups: ["servicecatalog.k8s.io"]
    apiVersions: ["v1beta1"]
    resources: ["serviceinstances"]

---

apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: {{ template "fullname" . }}-validating-webhook
  namespace: "{{ .Release.Namespace }}"
webhooks:
- name: validating.serviceinstances.servicecatalog.k8s.io
  clientConfig:
    caBundle: {{ b64enc $ca.Cert }}
    service:
      name: {{ template "fullname" . }}-webhook
      namespace: "{{ .Release.Namespace }}"
      path: "/validating-serviceinstances"
  failurePolicy: Fail
  rules:
  - operations: [ "CREATE", "UPDATE" ]
    apiGroups: ["servicecatalog.k8s.io"]
    apiVersions: ["v1beta1"]
    resources: ["serviceinstances"]
```

> **NOTE:** Each kind is registered separately in the **webhooks** array.

If the resource is registered, the webhook logic will be triggered when the registered operation on this resource occurs. The example of the webhook logic implementation looks as follows:

```go
// Handle admission requests.
func (h *CreateUpdateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	si := &sc.ServiceInstance{}
	webhookutil.MatchKinds(si, req.Kind)
	h.decoder.Decode(req, si)

	switch req.Operation {
	case admissionTypes.Create:
		h.mutateOnCreate(ctx, req, mutated)
	case admissionTypes.Update:
		oldObj := &sc.ServiceInstance{}
		h.decoder.DecodeRaw(req.OldObject, oldObj)
		h.mutateOnUpdate(ctx, req, oldObj, mutated)
	default:
		return admission.Allowed("action not taken")
	}

	rawMutated, err := json.Marshal(mutated)
	return admission.PatchResponseFromRaw(req.AdmissionRequest.Object.Raw, rawMutated)
}
```

> **NOTE:** The webhook implementation logic is common for all of the resources.

For the webhook server implementation, the [sigs.k8s.io/controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) dependency is used in the latest version - `v0.2.0-beta.0`.
