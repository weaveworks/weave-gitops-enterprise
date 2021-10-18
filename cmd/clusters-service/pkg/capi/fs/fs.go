package fs

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/capi"
)

// New creates and returns a new Filesystem based template library.
func New(fs fs.FS, base string) *FSLibrary {
	return &FSLibrary{FS: fs, base: base}
}

type FSLibrary struct {
	fs.FS
	base string
}

// Flavours lists the flavours in a tree structure.
//
// The base directory is searched for directories for example:
// └── flavours
//     ├── 1.2.3
//     │   ├── template1.yaml
//     │   └── template2.yaml
//     └── 2.1.0
//         └── template1.yaml
//
// Would result in 3 templates, with two different versions 1.2.3 and 2.1.0.
func (f FSLibrary) Flavours() ([]*capi.Flavour, error) {
	dirs, err := fs.ReadDir(f.FS, f.base)
	if err != nil {
		return nil, fmt.Errorf("failed to ReadDir in Flavours(): %w", err)
	}
	var found []*capi.Flavour
	for _, v := range dirs {
		if v.IsDir() {
			versionFlavours, err := f.flavoursFromDir(filepath.Join(f.base, v.Name()), v.Name())
			if err != nil {
				return nil, err
			}
			found = append(found, versionFlavours...)
		}
	}
	sort.Slice(found, func(i, j int) bool { return found[i].Name+found[i].Version < found[j].Name+found[i].Version })
	return found, nil
}

// This parses the templates in the directory, and sets the version to the
// directory being parsed.
func (f FSLibrary) flavoursFromDir(dir, version string) ([]*capi.Flavour, error) {
	dirs, err := fs.ReadDir(f.FS, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to ReadDir listing flavours from %s: %w", dir, err)
	}
	var found []*capi.Flavour
	for _, v := range dirs {
		if !v.IsDir() {
			t, err := capi.ParseFileFromFS(f.FS, filepath.Join(dir, v.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to parse: %w", err)
			}

			params, err := capi.ParamsFromSpec(t.Spec)
			if err != nil {
				return nil, err
			}
			found = append(found, &capi.Flavour{
				Name:        t.ObjectMeta.Name,
				Description: t.Spec.Description,
				Params:      params,
				Version:     version,
			})
		}
	}
	return found, nil
}
