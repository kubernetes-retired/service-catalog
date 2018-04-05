package generator

import (
	"fmt"

	"math/rand"

	"sort"

	"github.com/golang/glog"
	"github.com/pmorie/go-open-service-broker-client/v2"
)

// GetCatalog will produce a valid GetCatalog response based on the generator settings.
func (g *Generator) GetCatalog() (*v2.CatalogResponse, error) {
	if len(g.Services) == 0 {
		return nil, fmt.Errorf("no services defined")
	}

	services := make([]v2.Service, len(g.Services))

	for s, gs := range g.Services {
		services[s].Plans = make([]v2.Plan, len(gs.Plans))
		service := &services[s]
		service.Name = g.ClassPool[s+g.ClassPoolOffset]
		service.Description = g.description(s)
		service.ID = IDFrom(g.ClassPool[s])
		service.DashboardClient = g.dashboardClient(service.Name)

		for property, count := range gs.FromPool {
			switch property {
			case Tags:
				service.Tags = g.tagNames(s, count)
			case Metadata:
				service.Metadata = g.metaNames(s, count)
			case Bindable:
				service.Bindable = count > 0
			case BindingsRetrievable:
				service.BindingsRetrievable = count > 0
			case Requires:
				service.Requires = g.requiresNames(s, count)
			}
		}

		planNames := g.planNames(s, len(service.Plans))
		for p, gp := range gs.Plans {
			plan := &service.Plans[p]
			plan.Name = planNames[p]
			plan.Description = g.description(1000 + 1000*s*p)
			plan.ID = IDFrom(planNames[p])

			for property, count := range gp.FromPool {
				switch property {
				case Metadata:
					plan.Metadata = g.metaNames(1000+1000*s*p, count)
				case Free:
					isFree := count > 0
					plan.Free = &isFree
				}
			}
		}
	}

	return &v2.CatalogResponse{
		Services: services,
	}, nil
}

func getSliceWithoutDuplicates(count int, seed int64, list []string) []string {

	if len(list) < count {
		glog.Error("not enough items in list")
		return []string{""}
	}

	rand.Seed(seed)

	set := map[string]int32{}

	// Get strings from list without duplicates
	for len(set) < count {
		x := rand.Int31n(int32(len(list)))
		set[list[x]] = x
	}

	keys := []string(nil)
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (g *Generator) description(seed int) string {
	return getSliceWithoutDuplicates(1, int64(seed), g.DescriptionPool)[0]
}

func (g *Generator) planNames(seed, count int) []string {
	return getSliceWithoutDuplicates(count, int64(seed), g.PlanPool)
}

func (g *Generator) tagNames(seed, count int) []string {
	return getSliceWithoutDuplicates(count, int64(seed*1000+1000), g.TagPool)
}

func (g *Generator) requiresNames(seed, count int) []string {
	return getSliceWithoutDuplicates(count, int64(seed*1000+2000), g.RequiresPool)
}

func (g *Generator) metaNames(seed, count int) map[string]interface{} {
	key := getSliceWithoutDuplicates(count, int64(seed*1000+3000), g.MetadataPool)
	value := getSliceWithoutDuplicates(count, int64(seed*3000+4000), g.MetadataPool)
	meta := make(map[string]interface{}, count)
	for i := 0; i < len(key); i++ {
		meta[key[i]] = value[i]
	}
	return meta
}

func (g *Generator) dashboardClient(name string) *v2.DashboardClient {
	return &v2.DashboardClient{
		ID:          IDFrom(fmt.Sprintf("%s%s", name, "id")),
		Secret:      IDFrom(fmt.Sprintf("%s%s", name, "secret")),
		RedirectURI: "http://localhost:1234",
	}
}

//
//const okCatalogBytes = `{
//  "services": [{
//    "name": "fake-service",
//    "id": "acb56d7c-XXXX-XXXX-XXXX-feb140a59a66",
//    "description": "fake service",
//    "tags": ["tag1", "tag2"],
//    "requires": ["route_forwarding"],
//    "bindable": true,
//    "bindings_retrievable": true,
//    "metadata": {
//    	"a": "b",
//    	"c": "d"
//    },
//    "dashboard_client": {
//      "id": "398e2f8e-XXXX-XXXX-XXXX-19a71ecbcf64",
//      "secret": "277cabb0-XXXX-XXXX-XXXX-7822c0a90e5d",
//      "redirect_uri": "http://localhost:1234"
//    },
//    "plan_updateable": true,
//    "plans": [{
//      "name": "fake-plan-1",
//      "id": "d3031751-XXXX-XXXX-XXXX-a42377d3320e",
//      "description": "description1",
//      "metadata": {
//      	"b": "c",
//      	"d": "e"
//      }
//    }]
//  }]
//}`
