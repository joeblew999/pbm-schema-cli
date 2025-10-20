package apply_test

import (
	"context"
	"os"
	"testing"

	"github.com/joeblew999/pbm-schema-cli/pkg/apply"
	"github.com/joeblew999/pbm-schema-cli/pkg/bundle"
	mockpb "github.com/joeblew999/pbm-schema-cli/pkg/pbadmin/mock"
		"github.com/joeblew999/pbm-schema-cli/pkg/plan"
)

func TestApply_AddsCollectionsAndFields(t *testing.T){
	live, _ := mockpb.LoadYAML("../../testdata/live/minimal.yaml")
	desiredYML, _ := os.ReadFile("../../testdata/bundles/v0.1.0/schema.yaml")
	d, _ := bundle.ParseDesired(desiredYML)
	pl := plan.Compute(live.Collections, d)
	if err := apply.Execute(context.Background(), live, pl, live.Collections); err != nil { t.Fatal(err) }
	// naive checks
	foundInvoices := false
	for _, c := range live.Collections { if c.Name == "invoices" { foundInvoices = true } }
	if !foundInvoices { t.Fatalf("expected invoices collection") }
}
