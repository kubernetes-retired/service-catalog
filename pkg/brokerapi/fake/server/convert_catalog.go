package server

import (
	pkgbrokerapi "github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"github.com/pivotal-cf/brokerapi"
)

// ConvertCatalog converts a (github.com/kubernetes-incubator/service-catalog/pkg/brokerapi).Catalog
// to an array of brokerapi.Services
func ConvertCatalog(cat *pkgbrokerapi.Catalog) []brokerapi.Service {
	ret := make([]brokerapi.Service, len(cat.Services))
	for i, svc := range cat.Services {
		ret[i] = convertService(svc)
	}
	return ret
}

func convertService(svc *pkgbrokerapi.Service) brokerapi.Service {
	return brokerapi.Service{
		ID:            svc.ID,
		Name:          svc.Name,
		Description:   svc.Description,
		Bindable:      svc.Bindable,
		Tags:          svc.Tags,
		PlanUpdatable: svc.PlanUpdateable,
		Plans:         convertPlans(svc.Plans),
		// TODO: convert Requires, Metadata, DashboardClient
	}
}

func convertPlans(plans []pkgbrokerapi.ServicePlan) []brokerapi.ServicePlan {
	ret := make([]brokerapi.ServicePlan, len(plans))
	for i, plan := range plans {
		ret[i] = brokerapi.ServicePlan{
			ID:          plan.ID,
			Name:        plan.Name,
			Description: plan.Description,
			Free:        &plan.Free,
			Bindable:    plan.Bindable,
			// TODO: convert Metadata
		}
	}
	return ret
}
