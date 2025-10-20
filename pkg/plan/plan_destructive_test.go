package plan_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/joeblew999/pbm-schema-cli/pkg/model"
	"github.com/joeblew999/pbm-schema-cli/pkg/plan"
)

// current: users{email,display_name}, invoices{amount}
// desired: users{email} (drop display_name), no invoices (drop collection)
func TestComputePlan_DestructiveDrops(t *testing.T){
	current := []model.Collection{
		{ Name: "users", Type: "base", Schema: []model.Field{
			{Name: "email", Type: "email", Required: true, Unique: true},
			{Name: "display_name", Type: "text"},
		}},
		{ Name: "invoices", Type: "base", Schema: []model.Field{ {Name: "amount", Type: "number", Required: true} } },
	}
	desired := model.Desired{ Collections: []model.Collection{
		{ Name: "users", Type: "base", Schema: []model.Field{ {Name: "email", Type: "email", Required: true, Unique: true} } },
	}}

	// without destructive, no drops
	plSafe := plan.Compute(current, desired, plan.Options{})
	for _, s := range plSafe.Steps {
		if s.Kind == plan.DropField || s.Kind == plan.DropCollection {
			t.Fatalf("unexpected destructive step in safe plan: %v", s.Kind)
		}
	}

	// with destructive, expect 2 drops
	plForce := plan.Compute(current, desired, plan.Options{AllowDestructive: true})
	var got []string
	for _, s := range plForce.Steps { got = append(got, s.Msg) }
	wantContains := []string{
		"Drop field display_name from users",
		"Drop collection invoices",
	}
	for _, want := range wantContains {
		found := false
		for _, g := range got { if g == want { found = true; break } }
		if !found { t.Fatalf("missing step: %s
plan: %s", want, cmp.Diff(wantContains, got)) }
	}
}
