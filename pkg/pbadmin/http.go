package pbadmin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/joeblew999/pbm-schema-cli/internal/httpx"
	"github.com/joeblew999/pbm-schema-cli/pkg/model"
)

// Client is the PocketBase Admin API surface used by the tool.
type Client interface {
	ListCollections(ctx context.Context) ([]model.Collection, error)
	CreateCollection(ctx context.Context, col model.Collection) (model.Collection, error)
	UpdateCollection(ctx context.Context, id string, col model.Collection) (model.Collection, error)
	DeleteCollection(ctx context.Context, id string) error
}

// New returns a default HTTP-backed client.
func New(baseURL, adminToken string, opts ...Option) Client {
	c := &client{
		base: strings.TrimRight(baseURL, "/"),
		tok:  adminToken,
		h:    httpx.New(30 * time.Second),
	}
	for _, o := range opts { o(c) }
	return c
}

// --- implementation ---

type Option func(*client)

type client struct {
	base string
	tok  string
	h    *httpx.Client
}

func (c *client) ListCollections(ctx context.Context) ([]model.Collection, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", c.base+"/api/collections?perPage=200", nil)
	c.auth(req)
	res, err := c.h.Do(req)
	if err != nil { return nil, err }
	defer res.Body.Close()
	if res.StatusCode >= 300 { b, _ := io.ReadAll(res.Body); return nil, fmt.Errorf("GET collections: %d %s", res.StatusCode, string(b)) }
	var out struct{ Items []model.Collection `json:"items"` }
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil { return nil, err }
	for i := range out.Items {
		sort.Slice(out.Items[i].Schema, func(a,b int) bool { return out.Items[i].Schema[a].Name < out.Items[i].Schema[b].Name })
		sort.Strings(out.Items[i].Indexes)
	}
	sort.Slice(out.Items, func(i,j int) bool { return out.Items[i].Name < out.Items[j].Name })
	return out.Items, nil
}

func (c *client) CreateCollection(ctx context.Context, col model.Collection) (model.Collection, error) {
	body, _ := json.Marshal(col)
	req, _ := http.NewRequestWithContext(ctx, "POST", c.base+"/api/collections", strings.NewReader(string(body)))
	c.auth(req); req.Header.Set("Content-Type", "application/json")
	res, err := c.h.Do(req)
	if err != nil { return model.Collection{}, err }
	defer res.Body.Close()
	if res.StatusCode >= 300 { b, _ := io.ReadAll(res.Body); return model.Collection{}, fmt.Errorf("POST %q: %d %s", col.Name, res.StatusCode, string(b)) }
	var created model.Collection
	if err := json.NewDecoder(res.Body).Decode(&created); err != nil { return model.Collection{}, err }
	return created, nil
}

func (c *client) UpdateCollection(ctx context.Context, id string, col model.Collection) (model.Collection, error) {
	body, _ := json.Marshal(col)
	req, _ := http.NewRequestWithContext(ctx, "PATCH", c.base+"/api/collections/"+id, strings.NewReader(string(body)))
	c.auth(req); req.Header.Set("Content-Type", "application/json")
	res, err := c.h.Do(req)
	if err != nil { return model.Collection{}, err }
	defer res.Body.Close()
	if res.StatusCode >= 300 { b, _ := io.ReadAll(res.Body); return model.Collection{}, fmt.Errorf("PATCH %q: %d %s", col.Name, res.StatusCode, string(b)) }
	var updated model.Collection
	if err := json.NewDecoder(res.Body).Decode(&updated); err != nil { return model.Collection{}, err }
	return updated, nil
}

func (c *client) DeleteCollection(ctx context.Context, id string) error {
	req, _ := http.NewRequestWithContext(ctx, "DELETE", c.base+"/api/collections/"+id, nil)
	c.auth(req)
	res, err := c.h.Do(req)
	if err != nil { return err }
	defer res.Body.Close()
	if res.StatusCode >= 300 { return fmt.Errorf("DELETE collection %s: %d", id, res.StatusCode) }
	return nil
}

func (c *client) auth(req *http.Request) { if c.tok != "" { req.Header.Set("Authorization", "Bearer "+c.tok) } }
