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
	"github.com/joeblew999/pbm-schema-cli/pkg/pbadmin/mock"
)

func TestPlanEndpoint_WithBundleAndMock(t *testing.T){
	// build a pbz from fixtures
	schema, _ := os.ReadFile("../testdata/bundles/v0.1.0/schema.yaml")
	meta, _ := os.ReadFile("../testdata/bundles/v0.1.0/meta.json")
	var m bundle.Meta; _ = json.Unmarshal(meta, &m)
	f := &bundle.File{Meta:m, SchemaYAML: schema, Overlays: map[string][]byte{}, Seeds: map[string][]byte{}}
	tmp, _ := os.CreateTemp("", "pbm-*.pbz"); tmp.Close(); _ = bundle.Write(tmp.Name(), f)
	bz, _ := os.ReadFile(tmp.Name())
	b64 := base64.StdEncoding.EncodeToString(bz)

	// mock provider returns a live state with only users/email
	live, _ := mock.LoadYAML("../testdata/live/minimal.yaml")
	prov := func(u, t string) pbadmin.Client { return live }
	s := api.New(prov)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/v1/plan", bytes.NewReader([]byte(`{"target_url":"http://x","token":"t","bundle_b64":"`+b64+`"}`)))
	s.Routes().ServeHTTP(w, r)
	if w.Code != 200 { t.Fatalf("status=%d " , w.Code) }
}
