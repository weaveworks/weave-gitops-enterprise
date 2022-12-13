package templates

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	gapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/gitopstemplate/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/server"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/estimation"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
)

type templateCommandFlags struct {
	parameterValues []string
	export          bool
}

var flags templateCommandFlags

var CreateCommand = &cobra.Command{
	Use:   "template",
	Short: "Create template resources",
	Example: `
	  export or apply rendered resources of template to cluster or path
	  gitops create template.yaml --values key=value key=value --export > clusters/management/template.yaml
	`,
	RunE: templatesCmdRunE(),
}

func init() {
	CreateCommand.Flags().StringSliceVar(&flags.parameterValues, "values", []string{}, "Set parameter values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	CreateCommand.Flags().BoolVar(&flags.export, "export", false, "export in YAML format to stdout")
}

func templatesCmdRunE() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("template file is required")
		}

		parsedTemplate, err := parseTemplate(args[0])
		if err != nil {
			return fmt.Errorf("failed to parse template file %s: %w", args[0], err)
		}

		ctx := context.Background()
		vals := make(map[string]string)

		// parse parameter values
		for _, v := range flags.parameterValues {
			kv := strings.SplitN(v, "=", 2)
			if len(kv) == 2 {
				vals[kv[0]] = kv[1]
			}
		}

		getFilesRequest := server.GetFilesRequest{
			ParameterValues: vals,
			TemplateName:    parsedTemplate.Name,
			TemplateKind:    parsedTemplate.Kind,
		}

		templateResources, err := server.GetFiles(
			ctx, nil, logr.Discard(), estimation.NilEstimator(), nil,
			types.NamespacedName{}, types.NamespacedName{}, parsedTemplate, getFilesRequest, nil)
		if err != nil {
			return fmt.Errorf("failed to get template resources: %w", err)
		}

		renderedTemplate := templateResources.RenderedTemplate.Content

		if flags.export {
			err := export(*renderedTemplate, os.Stdout)
			if err != nil {
				return fmt.Errorf("failed to export rendered template: %w", err)
			}

			return nil
		}

		return nil
	}
}

// parse parses a template file and returns a GitOpsTemplate object
func parseTemplate(filename string) (*gapiv1.GitOpsTemplate, error) {
	gitOpsTemplate := gapiv1.GitOpsTemplate{}

	templateYAML, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", filename, err)
	}

	scheme := runtime.NewScheme()
	if err := gapiv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add GitOpsTemplate to scheme: %w", err)
	}

	var codecs = serializer.NewCodecFactory(scheme)
	decoder := codecs.UniversalDecoder(gapiv1.GroupVersion)

	_, _, err = decoder.Decode(templateYAML, nil, &gitOpsTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to decode template file %s: %w", filename, err)
	}

	return &gitOpsTemplate, nil
}

// export writes the rendered template to the specified output
func export(template string, out io.Writer) error {
	_, err := fmt.Fprintf(out, "%s", template)
	if err != nil {
		return fmt.Errorf("failed to write rendered template to output: %w", err)
	}

	return nil
}
