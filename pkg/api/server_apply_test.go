package api_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joeblew999/pbm-schema-cli/pkg/api"
	"github.com/joeblew999/pbm-schema-cli/pkg/bundle"
	"github.com/joeblew999/pbm-schema-cli/pkg/pbadmin"
	"github.com/joeblew999/pbm-schema-cli/pkg/pbadmin/mock"
)

func buildPBZFromFixture(t *testing.T) string {
	t.Helper()
	schema, _ := os.ReadFile("../testdata/bundles/v0.1.0/schema.yaml")
	meta, _ := os.ReadFile("../testdata/bundles/v0.1.0/meta.json")
	var m bundle.Meta; _ = json.Unmarshal(meta, &m)
	f := &bundle.File{Meta:m, SchemaYAML: schema, Overlays: map[string][]byte{}, Seeds: map[string][]byte{}}
	tmp, _ := os.CreateTemp("", "pbm-*.pbz"); tmp.Close(); _ = bundle.Write(tmp.Name(), f)
	return tmp.Name()
}

func TestApplyEndpoint_DryRunAndApply(t *testing.T){
	// pbz
	p := buildPBZFromFixture(t)
	bz, _ := os.ReadFile(p)
	b64 := base64.StdEncoding.EncodeToString(bz)

	// mock state: only users/email
	live, _ := mock.LoadYAML("../testdata/live/minimal.yaml")
	var applied bool
	prov := func(u, t string) pbadmin.Client { return live }
	s := api.New(prov)

	// dry-run
	body, _ := json.Marshal(map[string]any{
		"target_url": "http://x", "token": "t", "bundle_b64": b64, "dry_run": true,
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/v1/apply", bytes.NewReader(body))
	s.Routes().ServeHTTP(w, r)
	if w.Code != 200 { t.Fatalf("dry-run status=%d", w.Code) }

	// apply
	body2, _ := json.Marshal(map[string]any{
		"target_url": "http://x", "token": "t", "bundle_b64": b64, "dry_run": false,
	})
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodPost, "/v1/apply", bytes.NewReader(body2))
	s.Routes().ServeHTTP(w2, r2)
	if w2.Code != 200 { t.Fatalf("apply status=%d", w2.Code) }
	_ = applied
}
