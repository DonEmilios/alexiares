package intel

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadSignatures walks dir for *.yaml/*.yml files, parses each as a
// Signature, and validates it. It returns every problem found across
// the whole tree (joined via errors.Join) rather than failing on the
// first bad file, so a maintainer can fix a batch of signatures in one
// pass. Duplicate IDs across files are reported as errors: signature
// identity must be unique across the entire repository.
//
// A directory that does not exist yet is not an error — it returns an
// empty set, so a fresh checkout with no signatures installed still
// runs.
func LoadSignatures(dir string) ([]Signature, error) {
	var sigs []Signature
	var errs []error
	seen := make(map[string]string) // id -> file it was first seen in

	err := walkYAML(dir, func(path string, data []byte) {
		var sig Signature
		if err := yaml.Unmarshal(data, &sig); err != nil {
			errs = append(errs, fmt.Errorf("%s: parsing: %w", path, err))
			return
		}
		if err := sig.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", path, err))
			return
		}
		if first, dup := seen[sig.ID]; dup {
			errs = append(errs, fmt.Errorf("%s: duplicate signature id %q (first defined in %s)", path, sig.ID, first))
			return
		}
		seen[sig.ID] = path
		sigs = append(sigs, sig)
	})
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return sigs, nil
}

// LoadObservations walks dir for *.yaml/*.yml files, parses each as an
// Observation, and validates it. Like LoadSignatures, it aggregates
// every problem found rather than stopping at the first.
func LoadObservations(dir string) ([]Observation, error) {
	var obs []Observation
	var errs []error

	err := walkYAML(dir, func(path string, data []byte) {
		var o Observation
		if err := yaml.Unmarshal(data, &o); err != nil {
			errs = append(errs, fmt.Errorf("%s: parsing: %w", path, err))
			return
		}
		if err := o.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", path, err))
			return
		}
		obs = append(obs, o)
	})
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return obs, nil
}

// walkYAML calls fn with the contents of every *.yaml/*.yml file under
// dir, recursively. A missing dir is treated as empty, not an error.
func walkYAML(dir string, fn func(path string, data []byte)) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("%s: reading: %w", path, err)
		}
		fn(path, data)
		return nil
	})
}
