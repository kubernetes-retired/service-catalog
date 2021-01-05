// This is a generated file. Do not edit directly.
// Run ./contrib/hack/pin-dependency.sh to change pinned dependency versions.
// Run ./contrib/hack/update-vendor.sh to update go.mod files and the vendor directory.

module github.com/kubernetes-sigs/service-catalog

go 1.13

require (
	github.com/go-openapi/spec v0.19.3
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/google/gofuzz v1.1.0
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/go-multierror v1.0.0
	github.com/kubernetes-sigs/go-open-service-broker-client v0.0.0-20200527163240-4406bd2cb6b8
	github.com/mattn/go-runewidth v0.0.3 // indirect
	github.com/olekukonko/tablewriter v0.0.1
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/peterbourgon/mergemap v0.0.0-20130613134717-e21c03b7a721
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.0.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.4.0
	github.com/vrischmann/envconfig v1.1.0
	golang.org/x/crypto v0.0.0-20200429183012-4b2356b1ed79 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	golang.org/x/sys v0.0.0-20200509044756-6aff5f38e54f // indirect
	gomodules.xyz/jsonpatch/v2 v2.0.1
	k8s.io/api v0.18.2
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/apiserver v0.18.2
	k8s.io/cli-runtime v0.18.2
	k8s.io/client-go v0.18.2
	k8s.io/code-generator v0.18.2
	k8s.io/component-base v0.18.2
	k8s.io/gengo v0.0.0-20200413195148-3a45101e95ac
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	sigs.k8s.io/controller-runtime v0.6.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.1
	github.com/kubernetes-sigs/go-open-service-broker-client => github.com/kubernetes-sigs/go-open-service-broker-client v0.0.0-20200527163240-4406bd2cb6b8
	go.etcd.io/bbolt => go.etcd.io/bbolt v1.3.3
	go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738 // 3cf2f69b5738 is the SHA for git tag v3.4.3
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	sigs.k8s.io/structured-merge-diff/v3 => sigs.k8s.io/structured-merge-diff/v3 v3.0.0
)
