package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joeblew999/pbm-schema-cli/pkg/api"
	"github.com/joeblew999/pbm-schema-cli/pkg/apply"
	"github.com/joeblew999/pbm-schema-cli/pkg/bundle"
	"github.com/joeblew999/pbm-schema-cli/pkg/pbadmin"
	"github.com/joeblew999/pbm-schema-cli/pkg/plan"
	"gopkg.in/yaml.v3"
)

func main(){
	var (
		cmd    = flag.String("cmd", "", "export|plan|apply|serve")
		url    = flag.String("url", "", "PB admin URL")
		token  = flag.String("token", "", "Admin token")
		out    = flag.String("o", "", "output .pbz (export)")
		bdir   = flag.String("fixtures-bundle-dir", "", "dir with schema.yaml/meta.json to build a pbz")
		bundlePath = flag.String("bundle", "", "input bundle path (plan/apply)")
		bind   = flag.String("bind", ":8088", "api bind (serve)")
		force  = flag.Bool("force", false, "allow destructive operations (drop field/collection/index)")
	)
	flag.Parse()

	switch *cmd {
	case "export":
		if *out == "" { die("missing -o") }
		client := pbadmin.New(*url, *token)
		live, err := client.ListCollections(context.Background()); if err != nil { die(err) }
		desired := struct{ Collections any `yaml:"collections"` }{Collections: live}
		yml, err := yaml.Marshal(desired); if err != nil { die(err) }
		sum := sha256.Sum256(yml)
		meta := bundle.Meta{Name:"pocketbase", Version:"0.0.0", Created: time.Now().UTC().Format(time.RFC3339), Checksum: fmt.Sprintf("sha256:%x", sum[:]), Engine:"pocketbase>=0.22", Notes:"export"}
		f := &bundle.File{Meta: meta, SchemaYAML: yml, Overlays: map[string][]byte{}, Seeds: map[string][]byte{}}
		if err := bundle.Write(*out, f); err != nil { die(err) }
		fmt.Println("exported:", *out)
	case "plan", "apply":
		var pbz string
		if *bdir != "" {
			schema, err := os.ReadFile(*bdir+"/schema.yaml"); if err != nil { die(err) }
			metaBytes, err := os.ReadFile(*bdir+"/meta.json"); if err != nil { die(err) }
			var m bundle.Meta; if err := json.Unmarshal(metaBytes, &m); err != nil { die(err) }
			chk := sha256.Sum256(schema); if m.Checksum == "" { m.Checksum = fmt.Sprintf("sha256:%x", chk[:]) }
			f := &bundle.File{Meta: m, SchemaYAML: schema, Overlays: map[string][]byte{}, Seeds: map[string][]byte{}}
			tmp, _ := os.CreateTemp("", "pbm-*.pbz"); tmp.Close(); if err := bundle.Write(tmp.Name(), f); err != nil { die(err) }
			pbz = tmp.Name()
		} else { pbz = *bundlePath }
		b, err := bundle.Read(pbz); if err != nil { die(err) }
		d, err := bundle.ParseDesired(b.SchemaYAML); if err != nil { die(err) }
		client := pbadmin.New(*url, *token)
		cur, err := client.ListCollections(context.Background()); if err != nil { die(err) }
		pln := plan.Compute(cur, d, plan.Options{AllowDestructive: *force})
		if *cmd == "plan" { fmt.Println(render(pln)); return }
		if err := apply.Execute(context.Background(), client, pln, cur); err != nil { die(err) }
		fmt.Println("apply: done")
	case "serve":
		s := api.New(nil)
		fmt.Println("listening on", *bind)
		if err := http.ListenAndServe(*bind, s.Routes()); err != nil { die(err) }
	default:
		die("unknown -cmd")
	}
}

func render(p plan.Plan) string { if len(p.Steps)==0 { return "PLAN: no changes" }; s:="PLAN:
"; for _, st := range p.Steps { s += "  - "+st.Msg+"
" }; return s }

func die(v interface{}) { fmt.Fprintln(os.Stderr, "error:", v); os.Exit(1) }
