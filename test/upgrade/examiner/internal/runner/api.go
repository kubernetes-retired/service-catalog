package runner

type (
	// UpgradeTest is an interface to represent single upgrade test
	UpgradeTest interface {
		CreateResources(stop <-chan struct{}, namespace string) error
		TestResources(stop <-chan struct{}, namespace string) error
	}
)
