package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/joeblew999/pbm-schema-cli/pkg/apply"
	"github.com/joeblew999/pbm-schema-cli/pkg/bundle"
	"github.com/joeblew999/pbm-schema-cli/pkg/pbadmin"
	"github.com/joeblew999/pbm-schema-cli/pkg/plan"
)

// ClientProvider allows injecting mocks in tests.
type ClientProvider func(targetURL, token string) pbadmin.Client

type Server struct {
	router  *chi.Mux
	provide ClientProvider
}

func New(provide ClientProvider) *Server {
	if provide == nil { provide = func(u,t string) pbadmin.Client { return pbadmin.New(u,t) } }
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer, middleware.Timeout(60*time.Second))
	s := &Server{router: r, provide: provide}
	s.routes()
	return s
}

func (s *Server) Routes() http.Handler { return s.router }

func (s *Server) routes() {
	s.router.Get("/healthz", func(w http.ResponseWriter, r *http.Request){ w.WriteHeader(200); _,_ = w.Write([]byte("ok")) })
	s.router.Post("/v1/plan", s.handlePlan)
	s.router.Post("/v1/apply", s.handleApply)
}

// ---- payloads ----

type PlanRequest struct {
	TargetURL string `json:"target_url"`
	Token     string `json:"token"`
	BundleB64 string `json:"bundle_b64"`
}

type PlanResponse struct {
	Steps []struct{ Kind, Msg string } `json:"steps"`
}

func (s *Server) handlePlan(w http.ResponseWriter, r *http.Request) {
	var req PlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { httpError(w, err, 400); return }
	if req.TargetURL==""||req.Token==""||req.BundleB64=="" { httpError(w, errors.New("target_url, token, bundle_b64 required"),400); return }
	buf, err := base64.StdEncoding.DecodeString(req.BundleB64); if err != nil { httpError(w, err, 400); return }
	bz, err := readTempPBZ(buf); if err != nil { httpError(w, err, 400); return }
	defer bz.cleanup()
	f, err := bundle.Read(bz.path); if err != nil { httpError(w, err, 400); return }
	d, err := bundle.ParseDesired(f.SchemaYAML); if err != nil { httpError(w, err, 400); return }
	client := s.provide(req.TargetURL, req.Token)
	cur, err := client.ListCollections(r.Context()); if err != nil { httpError(w, err, 502); return }
	pl := plan.Compute(cur, d)
	out := PlanResponse{}
	for _, st := range pl.Steps { out.Steps = append(out.Steps, struct{Kind, Msg string}{Kind: string(st.Kind), Msg: st.Msg}) }
	jsonOK(w, out)
}

// Apply simplified (dry-run by client side)
type ApplyRequest struct {
	TargetURL string `json:"target_url"`
	Token     string `json:"token"`
	BundleB64 string `json:"bundle_b64"`
	DryRun    bool   `json:"dry_run"`
}

func (s *Server) handleApply(w http.ResponseWriter, r *http.Request) {
	var req ApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { httpError(w, err, 400); return }
	buf, err := base64.StdEncoding.DecodeString(req.BundleB64); if err != nil { httpError(w, err, 400); return }
	bz, err := readTempPBZ(buf); if err != nil { httpError(w, err, 400); return }
	defer bz.cleanup()
	f, err := bundle.Read(bz.path); if err != nil { httpError(w, err, 400); return }
	d, err := bundle.ParseDesired(f.SchemaYAML); if err != nil { httpError(w, err, 400); return }
	client := s.provide(req.TargetURL, req.Token)
	cur, err := client.ListCollections(r.Context()); if err != nil { httpError(w, err, 502); return }
	pl := plan.Compute(cur, d)
	if req.DryRun { jsonOK(w, map[string]any{"applied": false, "plan": render(pl)}); return }
	if err := apply.Execute(r.Context(), client, pl, cur); err != nil { httpError(w, err, 502); return }
	jsonOK(w, map[string]any{"applied": true, "plan": render(pl)})
}

// ---- helpers ----

func render(p plan.Plan) string { if len(p.Steps)==0 { return "PLAN: no changes" }; s:="PLAN:
"; for _, st := range p.Steps { s += "  - "+st.Msg+"
" }; return s }

func jsonOK(w http.ResponseWriter, v any) { w.Header().Set("Content-Type", "application/json"); _ = json.NewEncoder(w).Encode(v) }

func httpError(w http.ResponseWriter, err error, code int){ http.Error(w, err.Error(), code) }

// temp helper
type tempPBZ struct{ path string }
func readTempPBZ(b []byte) (*tempPBZ, error) { f, err := os.CreateTemp("", "pbm-*.pbz"); if err != nil { return nil, err }; if _, err := f.Write(b); err != nil { f.Close(); return nil, err }; f.Close(); return &tempPBZ{path:f.Name()}, nil }
func (t *tempPBZ) cleanup(){ if t!=nil && t.path!="" { _ = os.Remove(t.path) } }
