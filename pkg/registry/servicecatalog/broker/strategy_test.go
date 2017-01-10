package broker

import (
	"testing"
)

// TestBrokerStrategyTrivial is the testing of the trivial hardcoded
// boolean flags.
func TestBrokerStrategyTrivial(t *testing.T) {
	if createStrategy.NamespaceScoped() {
		t.Errorf("broker create must be namespace scoped")
	}
	if updateStrategy.NamespaceScoped() {
		t.Errorf("broker update must be namespace scoped")
	}
	if updateStrategy.AllowCreateOnUpdate() {
		t.Errorf("Job should not allow create on update")
	}
	if updateStrategy.AllowUnconditionalUpdate() {
		t.Errorf("Job should not allow create on update")
	}
}
