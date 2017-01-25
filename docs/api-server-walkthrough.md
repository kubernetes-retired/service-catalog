

kubernetes genericapiserver base

CURRENTLY TRACKING MASTER in k8s. Earliest possible release to
directly have the required binary compatibility may be k8s v1.6. K8s
changes on a daily basis so thing may break w/o being updated as K8s
changes.


Invoking `make apiserver` in the root directory will result in `apiserver` binary in the root directory.

OR

Compile `cmd/service-catalog/server.go` with `go build -o apiserver -v`

Start with:

```
# run etcd locally on the default port
$ etcd 
# switch to another shell and run
$ ./apiserver --etcd-servers localhost:2379
```

An etcd server is not hooked into yet, and is not required to be running.

A certificate will be generated in `/var/run/kubernetes/` so that directory must be creatable & writable by the running user.

In another term check for response from curl.
```
$ curl --cacert /var/run/kubernetes/apiserver.crt https://localhost:6443
{
  "paths": [
    "/apis",
    "/apis/servicecatalog.k8s.io",
    "/apis/servicecatalog.k8s.io/v1alpha1",
    "/healthz",
    "/healthz/ping",
    "/swaggerapi/",
    "/version"
  ]
}
```


Let's take a look at apis

```
# curl --cacert /var/run/kubernetes/apiserver.crt https://localhost:6443/apis
{
  "kind": "APIGroupList",
  "groups": [
    {
      "name": "servicecatalog.k8s.io",
      "versions": [],
      "preferredVersion": {
        "groupVersion": "servicecatalog.k8s.io/v1alpha1",
        "version": "v1alpha1"
      },
      "serverAddressByClientCIDRs": [
        {
          "clientCIDR": "0.0.0.0/0",
          "serverAddress": "9.52.233.169:6443"
        }
      ]
    }
  ]
}
```

And some of ours:
```
# curl --cacert /var/run/kubernetes/apiserver.crt https://localhost:6443/apis/servicecatalog.k8s.io
{
  "kind": "APIGroup",
  "apiVersion": "v1",
  "name": "servicecatalog.k8s.io",
  "versions": [],
  "preferredVersion": {
    "groupVersion": "servicecatalog.k8s.io/v1alpha1",
    "version": "v1alpha1"
  },
  "serverAddressByClientCIDRs": null
}
```

```
# curl --cacert /var/run/kubernetes/apiserver.crt https://localhost:6443/apis/servicecatalog.k8s.io/v1alpha1
{
  "kind": "APIResourceList",
  "apiVersion": "v1",
  "groupVersion": "servicecatalog.k8s.io/v1alpha1",
  "resources": null
}
```

kubectl seems happy enough:
```
$ kubectl --certificate-authority=/var/run/kubernetes/apiserver.crt --server=https://localhost:6443 version
Client Version: version.Info{Major:"1", Minor:"4", GitVersion:"v1.4.6+e569a27", GitCommit:"e569a27d02001e343cb68086bc06d47804f62af6", GitTreeState:"not a git tree", BuildDate:"2016-11-12T09:26:56Z", GoVersion:"go1.7.3", Compiler:"gc", Platform:"darwin/amd64"}
Server Version: version.Info{Major:"", Minor:"", GitVersion:"v0.0.0-master+$Format:%h$", GitCommit:"$Format:%H$", GitTreeState:"not a git tree", BuildDate:"1970-01-01T00:00:00Z", GoVersion:"go1.7.3", Compiler:"gc", Platform:"darwin/amd64"}
```
no version resource exists so this is to be expected.

```
$ kubectl --certificate-authority=/var/run/kubernetes/apiserver.crt --server=https://localhost:6443 get foo
the server doesn't have a resource type "foo"
```
no foo resource exists either.

```
$ kubectl --certificate-authority=/var/run/kubernetes/apiserver.crt --server=https://localhost:6443 api-versions
```
blank response. apiserver has no public apis. no errors either.



```
# kubectl --certificate-authority=/var/run/kubernetes/apiserver.crt --server=https://localhost:6443 create -f instance.yaml
instance "test-instance" created
```
query
```
kubectl --certificate-authority=/var/run/kubernetes/apiserver.crt --server=https://localhost:6443 get instance test-instance -o yaml
apiVersion: servicecatalog.k8s.io/v1alpha1
kind: Instance
metadata:
  creationTimestamp: 2017-01-25T21:57:48Z
  name: test-instance
  resourceVersion: "9"
  selfLink: /apis/servicecatalog.k8s.io/v1alpha1/namespaces//instances/test-instance
  uid: 4f88bd75-e349-11e6-8096-fa163e9a471d
spec:
  osbCredentials: ""
  osbDashboardURL: ""
  osbGuid: ""
  osbInternalID: ""
  osbLastOperation: ""
  osbPlanID: ""
  osbServiceID: ""
  osbSpaceGUID: ""
  osbType: ""
  parameters: null
  planName: ""
  serviceClassName: dugs awesome service instance
status:
  conditions: []
```

cleanup
```
 kubectl --certificate-authority=/var/run/kubernetes/apiserver.crt --server=https://localhost:6443 delete instance test-instance
instance "test-instance" deleted
```



