package mutators

import (
	"testing"

	"github.com/cohenjo/waste/go/config"
)

func TestDropTable(t *testing.T) {
	config.Config = config.LoadConfiguration()
	cng := &DropChange{Cluster: "dbdev", DatabaseName: "pavlok_test", TableName: "_jonyTest_del"}
	_, err := cng.RunChange()
	if err != nil {
		t.Fatalf("this is bad... %v", err)
	}

	t.Logf("###############################################################################################")
	t.Logf("after rich: %+v", cng)
}
