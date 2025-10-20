package pbadmin

import (
	"context"
	"net/http"
)

// DeleteCollection removes a collection by id.
func (c *client) DeleteCollection(ctx context.Context, id string) error {
	req, _ := http.NewRequestWithContext(ctx, "DELETE", c.base+"/api/collections/"+id, nil)
	c.auth(req)
	res, err := c.h.Do(req)
	if err != nil { return err }
	defer res.Body.Close()
	if res.StatusCode >= 300 { return fmt.Errorf("DELETE collection %s: %d", id, res.StatusCode) }
	return nil
}
