package v1beta1

// For Listers

//go:generate mockgen -destination=./mocks/mock_clusterserviceplan.go -package=mockv1beta1 github.com/kubernetes-incubator/service-catalog/pkg/client/listers_generated/servicecatalog/v1beta1 ClusterServicePlanLister

// Generate like:
// > go generate github.com/kubernetes-incubator/service-catalog/pkg/client/...

// Note: https://github.com/golang/mock/issues/30 The vendor paths need to be cleaned up after generation.
