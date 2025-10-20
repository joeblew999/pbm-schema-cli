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
	find := func(name string) (int, model.Collection, bool) {
		for i, x := range cur { if x.Name == name { return i, x, true } }
		return -1, model.Collection{}, false
	}
	for _, s := range p.Steps {
		switch s.Kind {
		case plan.CreateCollection:
			created, err := c.CreateCollection(ctx, *s.NewCollection)
			if err != nil { return err }
			cur = append(cur, created)
		case plan.AddField:
			idx, col, ok := find(s.AddField.CollectionName); if !ok { return errors.New("collection missing: "+s.AddField.CollectionName) }
			has := false; for _, f := range col.Schema { if f.Name == s.AddField.Field.Name { has = true } }
			if !has { col.Schema = append(col.Schema, s.AddField.Field); sort.Slice(col.Schema, func(a,b int) bool { return col.Schema[a].Name < col.Schema[b].Name }) }
			updated, err := c.UpdateCollection(ctx, col.ID, col); if err != nil { return err }
			cur[idx] = updated
		case plan.AddIndex:
			idx, col, ok := find(s.AddIndex.CollectionName); if !ok { return errors.New("collection missing: "+s.AddIndex.CollectionName) }
			have := false; for _, x := range col.Indexes { if x == s.AddIndex.Index { have = true } }
			if !have { col.Indexes = append(col.Indexes, s.AddIndex.Index) }
			updated, err := c.UpdateCollection(ctx, col.ID, col); if err != nil { return err }
			cur[idx] = updated
		case plan.UpdateRules:
			idx, col, ok := find(s.UpdateRules.CollectionName); if !ok { return errors.New("collection missing: "+s.UpdateRules.CollectionName) }
			updated, err := c.UpdateCollection(ctx, col.ID, col); if err != nil { return err }
			cur[idx] = updated
		case plan.DropField:
			idx, col, ok := find(s.DropField.CollectionName); if !ok { return errors.New("collection missing: "+s.DropField.CollectionName) }
			var ns []model.Field; for _, f := range col.Schema { if f.Name != s.DropField.FieldName { ns = append(ns, f) } }
			col.Schema = ns; updated, err := c.UpdateCollection(ctx, col.ID, col); if err != nil { return err }
			cur[idx] = updated
		case plan.DropIndex:
			idx, col, ok := find(s.DropIndex.CollectionName); if !ok { return errors.New("collection missing: "+s.DropIndex.CollectionName) }
			var ni []string; for _, x := range col.Indexes { if x != s.DropIndex.Index { ni = append(ni, x) } }
			col.Indexes = ni; updated, err := c.UpdateCollection(ctx, col.ID, col); if err != nil { return err }
			cur[idx] = updated
		case plan.DropCollection:
			_, col, ok := find(s.DropCollection.Name); if !ok { continue }
			if err := c.DeleteCollection(ctx, col.ID); err != nil { return err }
			var nn []model.Collection; for _, c0 := range cur { if c0.Name != s.DropCollection.Name { nn = append(nn, c0) } }
			cur = nn
		}
	}
	return nil
}
