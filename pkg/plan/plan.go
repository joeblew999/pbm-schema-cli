package plan

import "github.com/joeblew999/pbm-schema-cli/pkg/model"

type Kind string

const (
	CreateCollection Kind = "create_collection"
	AddField         Kind = "add_field"
)

type Step struct {
	Kind Kind
	Msg  string

	NewCollection *model.Collection
	AddField      struct {
		CollectionName string
		Field          model.Field
	}
}

type Plan struct{ Steps []Step }

func Compute(current []model.Collection, desired model.Desired) Plan {
	curByName := map[string]model.Collection{}
	for _, c := range current { curByName[c.Name] = c }
	var steps []Step
	for _, want := range desired.Collections {
		cur, ok := curByName[want.Name]
		if !ok {
			col := want
			steps = append(steps, Step{Kind: CreateCollection, Msg: "Create collection "+want.Name, NewCollection: &col})
			continue
		}
		curFields := map[string]model.Field{}
		for _, f := range cur.Schema { curFields[f.Name] = f }
		for _, wf := range want.Schema {
			if _, ok := curFields[wf.Name]; !ok {
				st := Step{Kind: AddField, Msg: "Add field "+wf.Name+" to "+want.Name}

				st.AddField.CollectionName = want.Name
				st.AddField.Field = wf
				steps = append(steps, st)
			}
		}
	}
	return Plan{Steps: steps}
}
