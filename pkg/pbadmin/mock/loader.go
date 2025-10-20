package mock

import (
	"os"

	"gopkg.in/yaml.v3"
	"github.com/joeblew999/pbm-schema-cli/pkg/model"
)

func LoadYAML(path string) (*Client, error) {
	b, err := os.ReadFile(path)
	if err != nil { return nil, err }
	var d model.Desired
	if err := yaml.Unmarshal(b, &d); err != nil { return nil, err }
	return NewFromCollections(d.Collections), nil
}
