

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
$ ./apiserver --etcd-servers localhost
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



