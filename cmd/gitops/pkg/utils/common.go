package utils

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/tomwright/dasel"
	"gopkg.in/yaml.v2"
)

const (
	coreManifestCount = 2
	coreManifestName  = "ww-gitops"
)

type ConfigStatus int

const (
	Missing ConfigStatus = iota
	Partial
	Embedded
	Valid
)

func (cs ConfigStatus) String() string {
	switch cs {
	case Missing:
		return "Missing"
	case Partial:
		return "Partial"
	case Embedded:
		return "Embedded"
	case Valid:
		return "Valid"
	default:
		return "UnknownStatus"
	}
}

type WalkResult struct {
	Status ConfigStatus
	Path   string
}

func (wr WalkResult) Error() string {
	return fmt.Sprintf("found %s: with status: %s", wr.Path, wr.Status)
}

func FindCoreConfig(dir string) WalkResult {
	err := filepath.WalkDir(dir,
		func(path string, _ fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if !strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml") {
				return nil
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			r := bytes.NewReader(data)
			decoder := yaml.NewDecoder(r)
			docs := []map[string]interface{}{}

			for {
				var entry map[string]interface{}
				if err := decoder.Decode(&entry); err == io.EOF {
					break
				}

				docs = append(docs, entry)
			}

			rootNode := dasel.New(docs)
			foundPartial := false

			_, err = rootNode.QueryMultiple(fmt.Sprintf(".(kind=HelmRelease)(.metadata.name=%s)", coreManifestName))
			if err == nil {
				foundPartial = true
			}

			_, err = rootNode.QueryMultiple(fmt.Sprintf(".(kind=HelmRepository)(.metadata.name=%s)", coreManifestName))
			if err != nil {
				if foundPartial {
					return WalkResult{Status: Partial, Path: path}
				}

				return nil
			}

			// retrieve the number of top-level entries from the file
			val, err := rootNode.Query(".[#]")
			if err != nil {
				return nil
			}

			if val.InterfaceValue() != coreManifestCount {
				return WalkResult{Status: Embedded, Path: path}
			}

			return WalkResult{Status: Valid, Path: path}
		})

	if val, ok := err.(WalkResult); ok {
		return val
	}

	return WalkResult{Status: Missing, Path: ""}
}
