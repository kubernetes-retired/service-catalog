

kubernetes genericapiserver base

Start with:
```
$ go run server.go
```

In another term check for response
```
$ curl localhost:8080
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
$ kubectl -s localhost:8080 version
Client Version: version.Info{Major:"1", Minor:"4", GitVersion:"v1.4.6+e569a27", GitCommit:"e569a27d02001e343cb68086bc06d47804f62af6", GitTreeState:"not a git tree", BuildDate:"2016-11-12T09:26:56Z", GoVersion:"go1.7.3", Compiler:"gc", Platform:"darwin/amd64"}
Couldn't read server version from server: the server could not find the requested resource

$ kubectl -s localhost:8080 get foo
the server doesn't have a resource type "foo"

```

