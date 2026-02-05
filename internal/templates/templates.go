package templates

import "embed"

//go:embed operator.md carnie.md issue-to-beads.md.tmpl workorder.md.tmpl
var FS embed.FS

// Load reads an embedded template file by name.
func Load(name string) (string, error) {
	data, err := FS.ReadFile(name)
	return string(data), err
}
