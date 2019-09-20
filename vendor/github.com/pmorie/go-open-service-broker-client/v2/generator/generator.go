package generator

import (
	"math/rand"
)

func CreateGenerator(serviceCount int, params Parameters) *Generator {
	rand.Seed(params.Seed)
	g := Generator{}
	g.Services = make(Services, serviceCount)
	for s, _ := range g.Services {
		service := &g.Services[s]
		// Fill out the service.
		service.FromPool = Pull{}
		if params.Services.Tags > 0 {
			service.FromPool[Tags] = randn(params.Services.Tags)
		}
		if params.Services.Metadata > 0 {
			service.FromPool[Metadata] = randn(params.Services.Metadata)
		}
		if params.Services.Requires > 0 {
			service.FromPool[Requires] = randn(params.Services.Requires)
		}
		if params.Services.Bindable > 0 {
			service.FromPool[Bindable] = randn(params.Services.Bindable)
		}
		if params.Services.BindingsRetrievable > 0 {
			service.FromPool[BindingsRetrievable] = randn(params.Services.BindingsRetrievable)
		}

		// How many plans will this service have? Needs at least one.
		planCount := randn(params.Services.Plans)
		if planCount == 0 {
			planCount = 1
		}
		service.Plans = make(Plans, planCount)

		// Fill out the plan.
		for p, _ := range service.Plans {
			plan := &service.Plans[p]
			plan.FromPool = Pull{}
			if params.Plans.Metadata > 0 {
				plan.FromPool[Metadata] = randn(params.Plans.Metadata)
			}
			if params.Plans.Bindable > 0 {
				plan.FromPool[Bindable] = randn(params.Plans.Bindable)
			}
			if params.Plans.Free > 0 {
				plan.FromPool[Free] = randn(params.Plans.Free)
			}
		}
	}
	return &g
}

// [0-n)
func randn(n int) int {
	return int(rand.Int31n(int32(n)))
}
