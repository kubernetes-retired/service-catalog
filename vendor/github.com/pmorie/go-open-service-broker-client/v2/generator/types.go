package generator

// generator holds the parameters for generated responses.
type Generator struct {
	Services        Services
	ClassPoolOffset int
	ClassPool       Pool
	DescriptionPool Pool
	PlanPool        Pool
	TagPool         Pool
	MetadataPool    Pool
	RequiresPool    Pool
}

type Pool []string

type Services []Service

type Service struct {
	FromPool Pull
	Plans    Plans
}

type Plans []Plan

type Plan struct {
	FromPool Pull
}

type Pull map[Property]int

type Property string

const (
	Tags                Property = "tags"
	Metadata            Property = "metadata"
	Requires            Property = "Requires"
	Bindable            Property = "bindable"
	BindingsRetrievable Property = "bindings_retrievable"
	Free                Property = "free"
)

type Parameters struct {
	Seed     int64
	Services ServiceRanges
	Plans    PlanRanges
}

type ServiceRanges struct {
	// Plans will default to 1. Range will be [1-Plans)
	Plans               int
	Tags                int
	Metadata            int
	Requires            int
	Bindable            int
	BindingsRetrievable int
}

type PlanRanges struct {
	Metadata int
	Bindable int
	Free     int
}

// Classes can have:
// - tags
// - metadata
// - bindable
// - requires
// - bindings retrievable

// Plans can have:
// - metadata
// - free
// - bindable
