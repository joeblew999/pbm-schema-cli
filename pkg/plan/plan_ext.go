package plan

import "github.com/joeblew999/pbm-schema-cli/pkg/model"

// Extend kinds to include destructive and metadata changes.
type Kind string

const (
	CreateCollection Kind = "create_collection"
	AddField         Kind = "add_field"
	DropField        Kind = "drop_field"
	DropCollection   Kind = "drop_collection"
	AddIndex         Kind = "add_index"
	DropIndex        Kind = "drop_index"
	UpdateRules      Kind = "update_rules"
)

// Options tune the planner behavior.
type Options struct {
	AllowDestructive bool
}

type Step struct {
	Kind Kind
	Msg  string

	NewCollection *model.Collection
	AddField      struct { CollectionName string; Field model.Field }
	DropField     struct { CollectionName string; FieldName string }
	DropCollection struct { Name string }
	AddIndex      struct { CollectionName string; Index string }
	DropIndex     struct { CollectionName string; Index string }
	UpdateRules   struct { CollectionName string }
}

type Plan struct{ Steps []Step }

func Compute(current []model.Collection, desired model.Desired, opts ...Options) Plan {
	var cfg Options
	if len(opts)>0 { cfg = opts[0] }
	curByName := map[string]model.Collection{}
	for _, c := range current { curByName[c.Name] = c }
	desByName := map[string]model.Collection{}
	for _, c := range desired.Collections { desByName[c.Name] = c }
	var steps []Step
	// creates & additive 
	for _, want := range desired.Collections {
		cur, ok := curByName[want.Name]
		if !ok { col := want; steps = append(steps, Step{Kind: CreateCollection, Msg: "Create collection "+want.Name, NewCollection: &col}); continue }
		curFields := map[string]model.Field{}; for _, f := range cur.Schema { curFields[f.Name] = f }
		for _, wf := range want.Schema { if _, ok := curFields[wf.Name]; !ok { st := Step{Kind: AddField, Msg: "Add field "+wf.Name+" to "+want.Name}; st.AddField = struct{CollectionName string; Field model.Field}{want.Name, wf}; steps = append(steps, st) } }
		// indexes additive
		curIdx := map[string]bool{}; for _, idx := range cur.Indexes { curIdx[idx] = true }
		for _, idx := range want.Indexes { if !curIdx[idx] { steps = append(steps, Step{Kind: AddIndex, Msg: "Add index to "+want.Name+": "+idx, AddIndex: struct{CollectionName string; Index string}{want.Name, idx}}) } }
		// rules change
		if ruleChanged(cur, want) { steps = append(steps, Step{Kind: UpdateRules, Msg: "Update rules for "+want.Name, UpdateRules: struct{CollectionName string}{want.Name}}) }
	}
	// destructive: drops
	if cfg.AllowDestructive {
		for _, cur := range current {
			if _, ok := desByName[cur.Name]; !ok { steps = append(steps, Step{Kind: DropCollection, Msg: "Drop collection "+cur.Name, DropCollection: struct{ Name string }{cur.Name}}); continue }
			want := desByName[cur.Name]
			wantFields := map[string]bool{}; for _, f := range want.Schema { wantFields[f.Name] = true }
			for _, f := range cur.Schema { if !wantFields[f.Name] { steps = append(steps, Step{Kind: DropField, Msg: "Drop field "+f.Name+" from "+cur.Name, DropField: struct{CollectionName, FieldName string}{cur.Name, f.Name}}) } }
			// drop indexes
			wantIdx := map[string]bool{}; for _, x := range want.Indexes { wantIdx[x] = true }
			for _, x := range cur.Indexes { if !wantIdx[x] { steps = append(steps, Step{Kind: DropIndex, Msg: "Drop index from "+cur.Name+": "+x, DropIndex: struct{CollectionName, Index string}{cur.Name, x}}) } }
		}
	}
	return Plan{Steps: steps}
}

func ruleChanged(a, b model.Collection) bool {
	return strp(a.ListRule)!=strp(b.ListRule) || strp(a.ViewRule)!=strp(b.ViewRule) || strp(a.CreateRule)!=strp(b.CreateRule) || strp(a.UpdateRule)!=strp(b.UpdateRule) || strp(a.DeleteRule)!=strp(b.DeleteRule)
}
func strp(s *string) string { if s==nil { return "" }; return *s }
