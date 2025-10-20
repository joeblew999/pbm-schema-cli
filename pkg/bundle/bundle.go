package bundle

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/joeblew999/pbm-schema-cli/pkg/model"
)

type Meta struct {
	Name     string `json:"name" yaml:"name"`
	Version  string `json:"version" yaml:"version"`
	Created  string `json:"created" yaml:"created"`
	Checksum string `json:"checksum" yaml:"checksum"`
	Engine   string `json:"engine" yaml:"engine"`
	Notes    string `json:"notes" yaml:"notes"`
}

type File struct {
	Meta       Meta
	SchemaYAML []byte
	Overlays   map[string][]byte
	Seeds      map[string][]byte
}

func Read(path string) (*File, error) {
	r, err := zip.OpenReader(path)
	if err != nil { return nil, err }
	defer r.Close()
	f := &File{Overlays: map[string][]byte{}, Seeds: map[string][]byte{}}
	for _, zf := range r.File {
		rc, _ := zf.Open(); b, _ := io.ReadAll(rc); rc.Close()
		switch {
		case zf.Name == "meta.json":
			if err := json.Unmarshal(b, &f.Meta); err != nil { return nil, err }
		case zf.Name == "schema.yaml":
			f.SchemaYAML = b
		case strings.HasPrefix(zf.Name, "overlays/") && strings.HasSuffix(zf.Name, ".yaml"):
			name := strings.TrimSuffix(strings.TrimPrefix(zf.Name, "overlays/"), ".yaml")
			f.Overlays[name] = b
		case strings.HasPrefix(zf.Name, "data/seed/") && strings.HasSuffix(zf.Name, ".jsonl"):
			coll := strings.TrimSuffix(strings.TrimPrefix(zf.Name, "data/seed/"), ".jsonl")
			f.Seeds[coll] = b
		}
	}
	if len(f.SchemaYAML) == 0 { return nil, fmt.Errorf("bundle missing schema.yaml") }
	sum := sha256.Sum256(f.SchemaYAML)
	got := fmt.Sprintf("sha256:%x", sum[:])
	if f.Meta.Checksum != "" && f.Meta.Checksum != got { return nil, fmt.Errorf("checksum mismatch: %s != %s", got, f.Meta.Checksum) }
	return f, nil
}

func Write(path string, f *File) error {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	w := func(name string, b []byte) error { z, err := zw.Create(name); if err != nil { return err }; _, err = z.Write(b); return err }
	meta, _ := json.MarshalIndent(f.Meta, "", "  ")
	if err := w("meta.json", meta); err != nil { return err }
	if err := w("schema.yaml", f.SchemaYAML); err != nil { return err }
	for k, v := range f.Overlays { _ = w("overlays/"+k+".yaml", v) }
	for k, v := range f.Seeds { _ = w("data/seed/"+k+".jsonl", v) }
	if err := zw.Close(); err != nil { return err }
	return os.WriteFile(path, buf.Bytes(), 0644)
}

func ParseDesired(yml []byte) (model.Desired, error) {
	var d model.Desired
	if err := yaml.Unmarshal(yml, &d); err != nil { return d, err }
	sort.Slice(d.Collections, func(i,j int) bool { return d.Collections[i].Name < d.Collections[j].Name })
	for i := range d.Collections {
		sort.Slice(d.Collections[i].Schema, func(a,b int) bool { return d.Collections[i].Schema[a].Name < d.Collections[i].Schema[b].Name })
		sort.Strings(d.Collections[i].Indexes)
	}
	return d, nil
}
