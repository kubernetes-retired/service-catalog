

kubernetes genericapiserver base

CURRENTLY TRACKING MASTER in k8s

Start with:
```
$ go run server.go
```

In another term check for response
```
$ curl --cacert /var/run/kubernetes/apiserver.crt https://localhost:6443
{
  "paths": [
    "/apis",
    "/healthz",
    "/healthz/ping",
    "/swaggerapi/"
  ]
}
```

kubectl seems happy enough:
```
$ kubectl --certificate-authority=/var/run/kubernetes/apiserver.crt --server=https://localhost:6443 version
Client Version: version.Info{Major:"1", Minor:"4", GitVersion:"v1.4.6+e569a27", GitCommit:"e569a27d02001e343cb68086bc06d47804f62af6", GitTreeState:"not a git tree", BuildDate:"2016-11-12T09:26:56Z", GoVersion:"go1.7.3", Compiler:"gc", Platform:"darwin/amd64"}
Couldn't read server version from server: the server could not find the requested resource
```
no version resource exists

```
$ kubectl --certificate-authority=/var/run/kubernetes/apiserver.crt --server=https://localhost:6443 get foo
the server doesn't have a resource type "foo"
```
no foo resource.

```
$ kubectl --certificate-authority=/var/run/kubernetes/apiserver.crt --server=https://localhost:6443 api-versions
```
blank response. apiserver has no public apis. no errors either.
```



