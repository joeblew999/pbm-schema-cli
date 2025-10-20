package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"

		"github.com/joeblew999/pbm-schema-cli/pkg/bundle"
)

// pbzgen packages a fixture dir containing schema.yaml and meta.json into a .pbz
func main(){
	in := flag.String("in", "", "fixture dir (must contain schema.yaml, meta.json)")
	out := flag.String("out", "", "output .pbz path")
	flag.Parse()
	if *in=="" || *out=="" { die("usage: pbzgen -in DIR -out FILE.pbz") }
	schema, err := os.ReadFile(*in+"/schema.yaml"); if err != nil { die(err) }
	metaBytes, err := os.ReadFile(*in+"/meta.json"); if err != nil { die(err) }
	var m bundle.Meta; if err := json.Unmarshal(metaBytes, &m); err != nil { die(err) }
	chk := sha256.Sum256(schema); if m.Checksum=="" { m.Checksum = fmt.Sprintf("sha256:%x", chk[:]) }
	f := &bundle.File{Meta:m, SchemaYAML:schema, Overlays: map[string][]byte{}, Seeds: map[string][]byte{}}
	if err := bundle.Write(*out, f); err != nil { die(err) }
	fmt.Println("wrote", *out)
}

func die(v interface{}){ fmt.Fprintln(os.Stderr, v); os.Exit(1) }
