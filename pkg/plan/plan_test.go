package plan_test

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/joeblew999/pbm-schema-cli/pkg/bundle"
	mockpb "github.com/joeblew999/pbm-schema-cli/pkg/pbadmin/mock"
	"github.com/joeblew999/pbm-schema-cli/pkg/plan"
)

func TestComputePlan_Additive(t *testing.T) {
	live, err := mockpb.LoadYAML("../../testdata/live/minimal.yaml")
	if err != nil { t.Fatal(err) }
	desiredYML, _ := os.ReadFile("../../testdata/bundles/v0.1.0/schema.yaml")
	d, err := bundle.ParseDesired(desiredYML)
	if err != nil { t.Fatal(err) }
	pl := plan.Compute(live.Collections, d)
	want := []string{
		"Add field display_name to users",
		"Create collection invoices",
	}
	var got []string
	for _, s := range pl.Steps { got = append(got, s.Msg) }
	if diff := cmp.Diff(want, got); diff != "" { t.Fatalf("plan diff (-want +got):
%s", diff) }
}
