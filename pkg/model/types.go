package model

type Field struct {
	System      bool                   `json:"system" yaml:"system"`
	ID          string                 `json:"id,omitempty" yaml:"id,omitempty"`
	Name        string                 `json:"name" yaml:"name"`
	Type        string                 `json:"type" yaml:"type"`
	Required    bool                   `json:"required" yaml:"required"`
	Presentable bool                   `json:"presentable,omitempty" yaml:"presentable,omitempty"`
	Unique      bool                   `json:"unique" yaml:"unique"`
	Options     map[string]any         `json:"options,omitempty" yaml:"options,omitempty"`
}

type Collection struct {
	ID         string   `json:"id,omitempty" yaml:"id,omitempty"`
	Name       string   `json:"name" yaml:"name"`
	Type       string   `json:"type" yaml:"type"` // base|auth|view
	System     bool     `json:"system" yaml:"system"`
	ListRule   *string  `json:"listRule,omitempty" yaml:"listRule,omitempty"`
	ViewRule   *string  `json:"viewRule,omitempty" yaml:"viewRule,omitempty"`
	CreateRule *string  `json:"createRule,omitempty" yaml:"createRule,omitempty"`
	UpdateRule *string  `json:"updateRule,omitempty" yaml:"updateRule,omitempty"`
	DeleteRule *string  `json:"deleteRule,omitempty" yaml:"deleteRule,omitempty"`
	Indexes    []string `json:"indexes,omitempty" yaml:"indexes,omitempty"`
	Schema     []Field  `json:"schema" yaml:"schema"`
}

type Desired struct {
	Collections []Collection `yaml:"collections"`
}
