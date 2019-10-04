// This is a generated file. Do not edit directly.
// Run ./contrib/hack/pin-dependency.sh to change pinned dependency versions.
// Run ./contrib/hack/update-vendor.sh to update go.mod files and the vendor directory.

module github.com/kubernetes-sigs/service-catalog

go 1.13

require (
	github.com/Azure/go-autorest v11.1.2+incompatible
	github.com/dgrijalva/jwt-go v3.0.1-0.20160705203006-01aeca54ebda+incompatible // indirect
	github.com/emicklei/go-restful v1.1.4-0.20170410110728-ff4f55a20633 // indirect
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/spec v0.19.3
	github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d // indirect
	github.com/google/gofuzz v1.0.0
	github.com/google/uuid v1.1.1 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/hcl v0.0.0-20160711231752-d8c773c4cba1 // indirect
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/kubernetes-sigs/go-open-service-broker-client v0.0.0-20190909175253-906fa5f9c249
	github.com/magiconair/properties v1.7.1-0.20160816085511-61b492c03cf4 // indirect
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/mattn/go-runewidth v0.0.3 // indirect
	github.com/mitchellh/mapstructure v1.3.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20190414153302-2ae31c8b6b30 // indirect
	github.com/olekukonko/tablewriter v0.0.1
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/pelletier/go-toml v1.2.0 // indirect
	github.com/peterbourgon/mergemap v0.0.0-20130613134717-e21c03b7a721
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.2
	github.com/spf13/cast v0.0.0-20160730092037-e31f36ffc91a // indirect
	github.com/spf13/cobra v0.0.2-0.20180319062004-c439c4fa0937
	github.com/spf13/jwalterweatherman v0.0.0-20160311093646-33c24e77fb80 // indirect
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v0.0.0-20160820190039-7fb2782df3d8
	github.com/stretchr/testify v1.4.0
	github.com/vrischmann/envconfig v1.1.0
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/net v0.0.0-20200506145744-7e3656a0809f // indirect
	gomodules.xyz/jsonpatch/v2 v2.0.1
	k8s.io/api v0.15.7
	k8s.io/apiextensions-apiserver v0.15.7
	k8s.io/apimachinery v0.15.7
	k8s.io/apiserver v0.15.7
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.15.7
	k8s.io/component-base v0.15.7
	k8s.io/gengo v0.0.0-20190128074634-0689ccc1d7d6 // indirect
	k8s.io/klog v0.3.1
	k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
	k8s.io/kubectl v0.0.0-20190312160839-d28510b1b750
	k8s.io/utils v0.0.0-20190607212802-c55fbcfc754a // indirect
	sigs.k8s.io/controller-runtime v0.0.0-00010101000000-000000000000
	sigs.k8s.io/yaml v1.1.0
)

replace (
	k8s.io/api => k8s.io/api v0.15.7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.15.7
	k8s.io/apimachinery => k8s.io/apimachinery v0.15.7
	k8s.io/apiserver => k8s.io/apiserver v0.15.7
	k8s.io/client-go => k8s.io/client-go v0.15.7
	k8s.io/code-generator => k8s.io/code-generator v0.15.7
	k8s.io/component-base => k8s.io/component-base v0.15.7
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20190312160839-d28510b1b750
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.2.0
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v0.0.0-20190302045857-e85c7b244fd2
)
