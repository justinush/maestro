package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/justinush/maestro/pkg/maestro"
	"github.com/justinush/maestro/pkg/validate"
)

// LoadDir reads workflow definition files from dir (non-recursive), validates each
// with vopts, and registers them in a new Registry.
//
// Supported extensions: .yaml, .yml, .json. Files are processed in lexical path order.
// Fails fast: the first decode, validation, or duplicate-key error aborts and returns
// (nil, err) with no registry for the caller.
func LoadDir(dir string, vopts validate.Options) (*Registry, error) {
	paths, err := listWorkflowFiles(dir)
	if err != nil {
		return nil, err
	}

	reg := NewRegistry()
	for _, path := range paths {
		rt, err := maestro.LoadWithValidate(path, vopts)
		if err != nil {
			return nil, fmt.Errorf("workflow: load %q: %w", path, err)
		}
		if err := reg.Register(rt); err != nil {
			return nil, fmt.Errorf("workflow: load %q: %w", path, err)
		}
	}
	return reg, nil
}

func listWorkflowFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("workflow: read dir %q: %w", dir, err)
	}

	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		switch ext {
		case ".yaml", ".yml", ".json":
			paths = append(paths, filepath.Join(dir, name))
		}
	}
	sort.Strings(paths)
	return paths, nil
}
