package templates

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	gapiv1 "github.com/weaveworks/templates-controller/apis/gitops/v1alpha2"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/server"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/estimation"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
)

type Config struct {
	ParameterValues []string `mapstructure:"values"`
	Export          bool     `mapstructure:"export"`
	OutputDir       string   `mapstructure:"output-dir"`
	TemplateFile    string   `mapstructure:"template-file"`
}

var config Config
var configPath string

var CreateCommand = &cobra.Command{
	Use:   "template",
	Short: "Create template resources",
	Example: `
	  # export rendered resources of template to stdout
 	  gitops create template.yaml --values key1=value1,key2=value2 --export

	  # apply rendered resources of template to path
	  gitops create template.yaml --values key1=value1,key2=value2 --output-dir ./out 
	`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig(cmd)
	},
	RunE: templatesCmdRunE(),
}

func init() {
	flags := CreateCommand.Flags()
	flags.StringSlice("values", []string{}, "Set parameter values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	flags.Bool("export", false, "export in YAML format to stdout")
	flags.String("output-dir", "", "write YAML format to file")
	flags.String("template-file", "", "template file to use")
	flags.StringVar(&configPath, "config", "", "config file to use")
}

// initializeConfig reads in config file.
func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	v.AutomaticEnv()

	err := v.BindPFlags(cmd.Flags())
	if err != nil {
		return err
	}

	if configPath != "" {
		v.SetConfigFile(configPath)
		if err = v.ReadInConfig(); err != nil {
			return err
		}
	}

	err = v.Unmarshal(&config)
	if err != nil {
		return fmt.Errorf("error unmarshalling flags and env into config struct %w", err)
	}

	return nil
}

func templatesCmdRunE() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// set template name from args if not set in flags
		templateFile := config.TemplateFile
		if len(args) > 0 {
			templateFile = args[0]
		}

		if templateFile == "" {
			return errors.New("must specify template file")
		}

		parsedTemplate, err := parseTemplate(templateFile)
		if err != nil {
			return fmt.Errorf("failed to parse template file %s: %w", args[0], err)
		}

		ctx := context.Background()
		vals := make(map[string]string)

		// parse parameter values
		for _, v := range config.ParameterValues {
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

		renderedTemplate := ""
		for _, file := range templateResources.RenderedTemplate {
			renderedTemplate += *file.Content
		}

		if config.Export {
			renderedTemplate := ""
			for _, file := range templateResources.RenderedTemplate {
				renderedTemplate += "\n# path: " + *file.Path + "\n---\n" + *file.Content
			}

			err := export(renderedTemplate, os.Stdout)
			if err != nil {
				return fmt.Errorf("failed to export rendered template: %w", err)
			}

			return nil
		}

		if config.OutputDir != "" {
			for _, res := range templateResources.RenderedTemplate {
				filePath, err := securejoin.SecureJoin(flags.outputDir, *res.Path)
				if err != nil {
					return fmt.Errorf("failed to join %s to %s: %w", flags.outputDir, *res.Path, err)
				}
				directoryPath := filepath.Dir(filePath)

				err = os.MkdirAll(directoryPath, 0755)
				if err != nil {
					return fmt.Errorf("failed to create directory: %w", err)
				}

				file, err := os.Create(filePath)
				if err != nil {
					return fmt.Errorf("failed to create file: %w", err)
				}

				defer file.Close()

				_, err = file.Write([]byte(*res.Content))
				if err != nil {
					return fmt.Errorf("failed to write to file: %w", err)
				}
			}
			return nil
		}

		return errors.New("please provide either --export or --output-dir")
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
