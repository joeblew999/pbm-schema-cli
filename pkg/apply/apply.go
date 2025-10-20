package apply

import (
	"context"
	"errors"
	"sort"

	"github.com/joeblew999/pbm-schema-cli/pkg/model"
	"github.com/joeblew999/pbm-schema-cli/pkg/pbadmin"
	"github.com/joeblew999/pbm-schema-cli/pkg/plan"
)

func Execute(ctx context.Context, c pbadmin.Client, p plan.Plan, current []model.Collection) error {
	cur := current
	find := func(name string) (model.Collection, bool) {
		for _, x := range cur { if x.Name == name { return x, true } }
		return model.Collection{}, false
	}
	for _, s := range p.Steps {
		switch s.Kind {
		case plan.CreateCollection:
			created, err := c.CreateCollection(ctx, *s.NewCollection)
			if err != nil { return err }
			cur = append(cur, created)
		case plan.AddField:
			col, ok := find(s.AddField.CollectionName)
			if !ok { return errors.New("collection missing: "+s.AddField.CollectionName) }
			has := false
			for _, f := range col.Schema { if f.Name == s.AddField.Field.Name { has = true } }
			if !has {
				col.Schema = append(col.Schema, s.AddField.Field)
				sort.Slice(col.Schema, func(a,b int) bool { return col.Schema[a].Name < col.Schema[b].Name })
				updated, err := c.UpdateCollection(ctx, col.ID, col)
				if err != nil { return err }
				for i := range cur { if cur[i].ID == updated.ID { cur[i] = updated; break } }
			}
		}
	}
	return nil
}
