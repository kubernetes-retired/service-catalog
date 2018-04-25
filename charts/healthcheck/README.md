# Health Check

HealthCheck is a tool that can be used to monitor the basic health of the Service Catalog deployment.  It utilizes the User Provided Service Broker to perform basic end to end tests such as creating a Service Instance and Binding and verifying the operations are successful and then tearing it down.  It collects Prometheus metrics which can be scraped for analysis and alerting (ie error rate, execution time to create an instance or binding, etc).

For more information,
[visit the Service Catalog project on github](https://github.com/kubernetes-incubator/service-catalog).

## Installing the Chart

To install the chart with the release name `healthcheck`:

```bash
$ helm install charts/healthcheck --name healthcheck --namespace healthcheck --set imagePullPolicy=Never --set image=healthcheck:canary
```

## Uninstalling the Chart

To uninstall/delete the `healthcheck` deployment:

```bash
$ helm delete --purge healthcheck
```

The command removes all the Kubernetes components associated with the chart and
deletes the release.

## Configuration

The following tables lists the configurable parameters of the HealthCheck

Flag | Description
---- | ----
--broker-name string | Broker Name to test against - can only be ups-broker or osb-stub. | You must ensure the specified broker is deployed. (default "ups-broker")
--healthcheck-interval duration | How frequently the end to end health check should be performed (default 2m0s)
--alsologtostderr | log to standard error as well as files (default true)
--bind-address ip | The IP address on which to listen for the --secure-port port. The associated interface(s) must be reachable by the rest of the cluster, and by CLI/web clients. If blank, all interfaces will be used (0.0.0.0 for all IPv4 interfaces and :: for all IPv6 interfaces). (default 0.0.0.0)
--cert-dir string | The directory where the TLS certs are located. If --tls-cert-file and --tls-private-key-file are provided, this flag will be ignored. (default "/var/run/service-catalog-healthcheck")
--http2-max-streams-per-connection int | The limit that the server gives to clients for the maximum number of streams in an HTTP/2 connection. Zero means to use golang's default.
--kubernetes-config string | Path to config containing embedded authinfo for kubernetes. Default value is from environment variable KUBECONFIG
--kubernetes-context string | config context to use for kuberentes. If unset, will use value from 'current-context'
--kubernetes-host string | The kubernetes host, or apiserver, to connect to (default "http://127.0.0.1:8080")
--log_backtrace_at traceLocation | when logging hits line file:N, emit a stack trace (default :0)
--log_dir string | If non-empty, write log files in this directory
--logtostderr | log to standard error instead of files
--secure-port int | The port on which to serve HTTPS with authentication and authorization. If 0, don't serve HTTPS at all. (default 443)
--stderrthreshold severity | logs at or above this threshold go to stderr (default 2)
--tls-cert-file string | File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert). If HTTPS serving is enabled, and --tls-cert-file and --tls-private-key-file are not provided, a self-signed certificate and key are generated for the public address and saved to the directory specified by --cert-dir.
--tls-cipher-suites stringSlice | Comma-separated list of cipher suites for the server. Values are from tls package constants (https://golang.org/pkg/crypto/tls/#pkg-constants). If omitted, the default Go cipher suites will be used
--tls-min-version string | Minimum TLS version supported. Value must match version names from https://golang.org/pkg/crypto/tls/#pkg-constants.
--tls-private-key-file string | File containing the default x509 private key matching --tls-cert-file.
--tls-sni-cert-key namedCertKey | A pair of x509 certificate and private key file paths, optionally suffixed with a list of domain patterns which are fully qualified domain names, possibly with prefixed wildcard segments. If no domain patterns are provided, the names of the certificate are extracted. Non-wildcard matches trump over wildcard matches, explicit domain patterns trump over extracted names. For multiple key/certificate pairs, use the --tls-sni-cert-key multiple times. Examples: "example.crt,example.key" or "foo.crt,foo.key:*.foo.com,foo.com". (default [])
-v, --v Level | log level for V logs
--vmodule moduleSpec | comma-separated list of pattern=N settings for file-filtered logging



Specify each parameter using the `--set key=value[,key=value]` argument to
`helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be
provided while installing the chart. For example:

```bash
$ helm install charts/healthcheck --name healthcheck --namespace healthcheck \
  --values values.yaml
```
