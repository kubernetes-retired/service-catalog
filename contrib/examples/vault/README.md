# Service Catalog Vault Open Service Broker

This document shows how to use the
[Cloud Foundry Vault Service Broker](https://www.hashicorp.com/blog/cloud-foundry-vault-service-broker/)
on Kubernetes by leveraging the Service Catalog. The Vault installation steps
are not hardened for production use and are provided for ease of use only.
See the [Vault docs](https://github.com/hashicorp/vault) for more information
on how to properly configure a production grade service.

## Step 0 - Prerequisites

### Implement the Service Catalog

You *must* have the service catalog running as directed by the
[demo walkthrough](../../../docs/walkthrough.md). The steps in this guide assume
that the service catalog is running in the `service-catalog` context and that
the `catalog` namespace is already created.

## Step 1 - Install & Configure Vault (development mode)

Follow these steps to install Vault in development mode onto Kubernetes. In
development mode vault data will be stored in memory only and will no longer
be available when the service restarts. If you are using an existing Vault
environment you'll still need to adapt steps 3-5 below to appropriate equivalents.

1. Create Vault resources on Kubernetes

  ```console
  kubectl create -n catalog -f contrib/examples/vault/vault-devmode.yaml
  configmap "cf-broker-hcl" created
  deployment "vaultdev" created
  service "vaultdev" created
  ```

2. Obtain the root vault token from stdout.

  ```console
  kubectl logs -n catalog -l app=vault --tail=100 | grep "Root Token"
  Root Token: a7469f5d-1991-a9de-34b9-2c3804ae2c39
  ```
  
  Save the Root Token for use in the next few steps.
  
3. Shell into the Vault container and authenticate using the root token from the previous step.
  ```console
  kubectl exec -it -n catalog `(kubectl get pods -n catalog -l app=vault --output=jsonpath={.items..metadata.name})` /bin/sh
  ```

  Inside the Vault container

  ```console
  export VAULT_ADDR=http://127.0.0.1:8200
  vault auth a7469f5d-1991-a9de-34b9-2c3804ae2c39
  ```

  Sample Ouput

  ```console
  / # export VAULT_ADDR=http://127.0.0.1:8200
  / # vault auth a7469f5d-1991-a9de-34b9-2c3804ae2c39
  Successfully authenticated! You are now logged in.
  token: a7469f5d-1991-a9de-34b9-2c3804ae2c39
  token_duration: 0
  token_policies: [root]
  / # 
  ```

4. Create the Vault access policy and initial token for the Vault Open Service Broker

  Inside the Vault container

  ```console
  vault write sys/policy/cf-broker rules=@/etc/hcl/cf-broker.hcl
  vault token-create -period="30m" -orphan -policy=cf-broker
  ```

  ```console
  / # vault write sys/policy/cf-broker rules=@/etc/hcl/cf-broker.hcl
  Success! Data written to: sys/policy/cf-broker
  / # vault token-create -period="30m" -orphan -policy=cf-broker
  Key            	Value
  ---            	-----
  token          	1195ea9c-b495-8578-b030-75a484f5d111
  token_accessor 	b587ba20-d845-0596-ba1b-fb00c987bc1c
  token_duration 	30m0s
  token_renewable	true
  token_policies 	[cf-broker default]

  / # 
  ```

5. Capture the value of this cf-broker temporary access token for use by the
Vault Open Service Broker in Step 2 below.

## Step 2 - Run the Vault Open Service Broker on Kubernetes

The version of the open service broker used in this guide was
[modified slightly](https://github.com/emaildanwilson/cf-vault-service-broker/commit/394e0955cb6e532a97d06270c61f9aed083db50b)
in order to work with the current implementation of the Kubernetes Service Catalog.
This should no longer be required in the next update. This change is delivered as a
[public container](https://hub.docker.com/r/danw/cf-vault-service-broker/)
for ease of use in this guide.

1. Run the Vault Open Service Broker using the temporary cf-broker token above.

  ```console
  kubectl run -n catalog cf-vault-service-broker --env VAULT_ADDR=http://vaultdev.catalog.svc.cluster.local:8200 --env VAULT_TOKEN=1195ea9c-b495-8578-b030-75a484f5d111 --env SECURITY_USER_NAME=admin --env SECURITY_USER_PASSWORD=pass --image=danw/cf-vault-service-broker --expose=true --port=8000
  service "cf-vault-service-broker" created
  deployment "cf-vault-service-broker" created
  ```

2. Check the logs to make sure it starts normally

  ```console
  kubectl logs -n catalog -l run=cf-vault-service-broker
  [INFO] starting broker
  [DEBUG] creating mounts map[string]string{"cf/broker":"generic"}
  [INFO] sleeping for 5m41.5s
  [DEBUG] restoring bindings
  [DEBUG] listing directory "cf/broker/"
  [INFO] listDir cf/broker/ has no secret data
  [INFO] restored 0 binds and 0 instances
  [INFO] starting server on :8000
  [INFO] listing services
  ```

## Step 3 - Configure the Kubernetes Service Catalog

1. Create a `Secret` for Kubernetes to access the Vault Open Service Broker API

  ```console
  kubectl create secret generic vault -n catalog --from-literal=username=admin --from-literal=password=pass
  secrets "vault" created
  ```

2. Create a `Broker` for Vault in the Kubernetes Service Catalog

  ```console
  kubectl --context=service-catalog create -f contrib/examples/vault/vault-openservicebroker.yaml
  brokers "vault-broker" created
  ```

3. Check the status of the broker

  ```console
  kubectl --context=service-catalog get brokers vault-broker --output=jsonpath={.status.conditions..message}
  Successfully fetched catalog entries from broker.
  ```

4. View the service classes

  ```console
  kubectl --context=service-catalog get serviceclasses
  ```

  We should see something like:
  ```console
  NAME                    KIND                                          ALPHA TAGS   BINDABLE     BROKER NAME
  hashicorp-vault         ServiceClass.v1alpha1.servicecatalog.k8s.io   1 item(s)    true         vault-broker
  ```

5. Run the following command to see the details of this service class offering:

  ```console
  kubectl --context=service-catalog get serviceclasses hashicorp-vault -o yaml
  ```

  We should see something like:

  ```yaml
  alphaTags:
  - ""
  apiVersion: servicecatalog.k8s.io/v1alpha1
  bindable: true
  brokerName: vault-broker
  description: HashiCorp Vault Service Broker
  externalID: 0654695e-0760-a1d4-1cad-5dd87b75ed99
  externalMetadata: null
  kind: ServiceClass
  metadata:
    creationTimestamp: 2017-07-17T18:58:52Z
    name: hashicorp-vault
    resourceVersion: "821"
    selfLink: /apis/servicecatalog.k8s.io/v1alpha1/serviceclasseshashicorp-vault
    uid: f9c2f9b3-6b21-11e7-9c24-0242ac11000b
  planUpdatable: false
  plans:
  - description: Secure access to Vault's storage and transit backends
    externalID: 0654695e-0760-a1d4-1cad-5dd87b75ed99-shared
    externalMetadata: null
    free: true
    name: shared
  ```

## Step 4 - Provisioning a New Vault Secret using the Broker

1. Create the new Vault secret in the namespace `test-ns` from the Walkthrough.

  ```console
  kubectl --context=service-catalog -n test-ns create -f contrib/examples/vault/consumer.yaml
  ```

  Console output

  ```console
  instance "mycreds" created
  binding "mycreds" created
  ```

2. Check the status of the instance and binding

  ```console
  kubectl --context=service-catalog get bindings,instance -n test-ns mycreds --output=jsonpath={.items..status.conditions..message}
  ```

  We should see something like this to indicate the bound secret is not yet available.

  ```console
  Binding cannot begin because referenced instance "test-ns/mycreds" is not ready The instance was provisioned successfully
  ```

  After a few seconds this should change to.

  ```console
  Injected bind result The instance was provisioned successfully
  ```

  This means our binding is ready to use and that a new secret is available.
  
3. View the contents of the secret:

  ```console
  kubectl get secrets -n test-ns mycreds -o yaml
  ```

  We should see something like:

  ```yaml
  apiVersion: v1
  data:
    address: aHR0cDovL3ZhdWx0ZGV2LmNhdGFsb2cuc3ZjLmNsdXN0ZXIubG9jYWw6ODIwMC8=
    auth: eyJhY2Nlc3NvciI6Ijc2MDZhOGQ4LTIzYTQtZDUxYS1iOTJlLTlmNzk0YjMxN2U0MyIsInRva2VuIjoiM2M5MTEyNTMtMDFjZi1hODM1LWRiYTgtMWQ3YzFkOTBkMDI3In0=
    backends: eyJnZW5lcmljIjoiY2YvMWEzMGM4MzMtYTVhMC00ZTM5LWIyZjMtOTY3ZjgwMjZiZDhkL3NlY3JldCIsInRyYW5zaXQiOiJjZi8xYTMwYzgzMy1hNWEwLTRlMzktYjJmMy05NjdmODAyNmJkOGQvdHJhbnNpdCJ9
    backends_shared: eyJvcmdhbml6YXRpb24iOiJjZi85YWU0NWYzYy02NzJiLTExZTctODg4ZC0wODAwMjdiOGYxNjMvc2VjcmV0Iiwic3BhY2UiOiJjZi85YWU0NWYzYy02NzJiLTExZTctODg4ZC0wODAwMjdiOGYxNjMvc2VjcmV0In0=
  kind: Secret
  metadata:
    creationTimestamp: 2017-07-17T19:32:00Z
    name: mycreds
    namespace: test-ns
    resourceVersion: "188936"
    selfLink: /api/v1/namespaces/test-ns/secrets/mycreds
    uid: 9a760eaa-6b26-11e7-a51f-080027b8f163
  type: Opaque
  ```

  The data included in the secret above contains all the information required
  to connect up to vault and begin reading\writing other data. You could consume
  this secret from another pod at this point.

## Step 5 - Cleanup

1. Delete the Binding and Instance resources

  ```console
  kubectl --context=service-catalog -n test-ns delete -f contrib/examples/vault/consumer.yaml
  instance "mycreds" deleted
  binding "mycreds" deleted
  ```

2. Delete the Kubernetes Broker object for Vault

  ```console
  kubectl --context=service-catalog delete -f contrib/examples/vault/vault-openservicebroker.yaml
  broker "vault-broker" deleted
  ```

3. Delete the Broker secret

  ```console
  kubectl delete secret vault -n catalog
  secret "vault" deleted
  ```

4. Delete the Vault Open Service Broker

  ```console
  kubectl delete svc,deployment -n catalog cf-vault-service-broker
  service "cf-vault-service-broker" deleted
  deployment "cf-vault-service-broker" deleted
  ```

5. Delete the Vault service and resources

  ```console
  kubectl delete -n catalog -f contrib/examples/vault/vault-devmode.yaml
  configmap "cf-broker-hcl" deleted
  deployment "vaultdev" deleted
  service "vaultdev" deleted
  ```