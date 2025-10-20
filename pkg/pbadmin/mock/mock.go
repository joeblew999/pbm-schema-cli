package mock

import (
	"context"
	"slices"

	"github.com/joeblew999/pbm-schema-cli/pkg/model"
)

// Client is an in-memory implementation of pbadmin.Client for tests.
type Client struct { Collections []model.Collection }

func NewFromCollections(cols []model.Collection) *Client { return &Client{Collections: cols} }

func (m *Client) ListCollections(context.Context) ([]model.Collection, error) {
	out := make([]model.Collection, len(m.Collections))
	copy(out, m.Collections)
	return out, nil
}

func (m *Client) CreateCollection(ctx context.Context, col model.Collection) (model.Collection, error) {
	if col.ID == "" { col.ID = "c_" + col.Name }
	m.Collections = append(m.Collections, col)
	return col, nil
}

func (m *Client) UpdateCollection(ctx context.Context, id string, col model.Collection) (model.Collection, error) {
	i := slices.IndexFunc(m.Collections, func(c model.Collection) bool { return c.ID==id || c.Name==col.Name })
	if i < 0 { m.Collections = append(m.Collections, col); return col, nil }
	m.Collections[i] = col
	return col, nil
}

func (m *Client) DeleteCollection(ctx context.Context, id string) error {
	i := slices.IndexFunc(m.Collections, func(c model.Collection) bool { return c.ID==id })
	if i < 0 { return nil }
	m.Collections = append(m.Collections[:i], m.Collections[i+1:]...)
	return nil
}
