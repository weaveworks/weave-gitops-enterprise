package templates

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	securejoin "github.com/cyphar/filepath-securejoin"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	gapiv1 "github.com/weaveworks/templates-controller/apis/gitops/v1alpha2"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/server"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
	clitemplates "github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/templates"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/estimation"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops/core/logger"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/repo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
)

type Config struct {
	ParameterValues []string `mapstructure:"values"`
	Export          bool     `mapstructure:"export"`
	OutputDir       string   `mapstructure:"output-dir"`
	TemplateFile    string   `mapstructure:"template-file"`
	HelmRepoName    string   `mapstructure:"helm-repo-name"`
	Profiles        []string `mapstructure:"profiles"`
}

var config Config
var configPath string

var DefaultCluster = "default"
var DefaultHelmRepoNamespace = "flux-system"

var CreateCommand = &cobra.Command{
	Use:   "template",
	Short: "Create template resources",
	Example: `
	  # export rendered resources of template to stdout
 	  gitops create template.yaml --values key1=value1,key2=value2 --export

	  # apply rendered resources of template to path
	  gitops create template.yaml --values key1=value1,key2=value2 --output-dir ./out 

	  # specify a template file and a config file
	  gitops create template.yaml --config config.yaml --output-dir ./out

	  # specify template file and values in a config file
	  gitops create --config config.yaml --output-dir ./out
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
	flags.StringArray("profiles", []string{}, "Set profiles values files on the command line (--profile 'name=foo-profile,version=0.0.1,namespace=foo-system' --profile 'name=bar-profile,namespace=bar-system,values=bar-values.yaml')")
	flags.String("helm-repo-name", "weaveworks-charts", "name of the helm repo in the helm local cache")
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
	return func(_ *cobra.Command, args []string) error {
		log, err := logger.New(logger.DefaultLogLevel, true)
		if err != nil {
			return fmt.Errorf("failed to create logger: %w", err)
		}

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
			return fmt.Errorf("failed to parse template file %s: %w", templateFile, err)
		}

		params := make(map[string]string)

		// parse parameter values
		for _, v := range config.ParameterValues {
			kv := strings.SplitN(v, "=", 2)
			if len(kv) == 2 {
				params[kv[0]] = kv[1]
			}
		}

		profilesValues, err := clitemplates.ParseProfileFlags(config.Profiles)
		if err != nil {
			return fmt.Errorf("error parsing profiles: %w", err)
		}
		capiProfileValues := []*capiv1_proto.ProfileValues{}
		for _, profile := range profilesValues {
			capiProfileValues = append(capiProfileValues, &capiv1_proto.ProfileValues{
				Name:      profile.Name,
				Namespace: profile.Namespace,
				Version:   profile.Version,
				Values:    profile.Values,
				HelmRepository: &capiv1_proto.HelmRepositoryRef{
					Name:      config.HelmRepoName,
					Namespace: DefaultHelmRepoNamespace,
				},
			})
		}

		files, err := generateFilesLocally(parsedTemplate, params, config.HelmRepoName, capiProfileValues, cli.New(), log)
		if err != nil {
			return fmt.Errorf("failed to generate files locally: %w", err)
		}

		if config.Export {
			renderedTemplate := ""
			for _, file := range files {
				renderedTemplate += fmt.Sprintf("# path: %s\n---\n%s\n\n", file.Path, *file.Content)
			}

			err := export(renderedTemplate, os.Stdout)
			if err != nil {
				return fmt.Errorf("failed to export rendered template: %w", err)
			}

			return nil
		}

		if config.OutputDir != "" {
			for _, res := range files {
				filePath, err := securejoin.SecureJoin(config.OutputDir, res.Path)
				if err != nil {
					return fmt.Errorf("failed to join %s to %s: %w", config.OutputDir, res.Path, err)
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

// implement another helmrepofetcher

// implement a HelmRepoFetcher that fetches from the cluster
type localHelmRepoFetcher struct {
	helmRepo *sourcev1.HelmRepository
}

func (c *localHelmRepoFetcher) GetHelmRepository(ctx context.Context, helmRepo types.NamespacedName) (*sourcev1.HelmRepository, error) {
	// throw error if you request a helmrepo that is not the default
	if helmRepo.Name != c.helmRepo.Name {
		return nil, fmt.Errorf("helm repo %s not found, only single helm repo supported: %s", helmRepo.Name, c.helmRepo.Name)
	}

	return c.helmRepo, nil
}

func generateFilesLocally(tmpl *gapiv1.GitOpsTemplate, params map[string]string, helmRepoName string, profiles []*capiv1_proto.ProfileValues, settings *cli.EnvSettings, log logr.Logger) ([]git.CommitFile, error) {
	templateHasRequiredProfiles, err := templates.TemplateHasRequiredProfiles(tmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to check if template has required profiles: %w", err)
	}

	var chartsCache helm.ProfilesGeneratorCache = helm.NilProfilesGeneratorCache{}
	var helmRepo *sourcev1.HelmRepository
	if len(profiles) > 0 || templateHasRequiredProfiles {
		entry, index, err := localHelmRepo(helmRepoName, settings)
		if err != nil {
			return nil, fmt.Errorf(
				"template has profiles and loading local helm repo data failed, try `helm repo add`. (RepositoryConfig: %s, RepositoryCache: %s): %w",
				settings.RepositoryConfig,
				settings.RepositoryCache,
				err,
			)
		}
		helmRepo = fluxHelmRepo(entry)
		chartsCache = helm.NewHelmIndexFileReader(index)
	}

	helmRepoFetcher := &localHelmRepoFetcher{helmRepo: helmRepo}

	templateResources, err := server.GetFiles(
		context.Background(),
		nil, // no need for a kube client as we're providing the helm repo no
		helmRepoFetcher,
		nil,
		log,
		estimation.NilEstimator(),
		chartsCache,
		types.NamespacedName{Name: DefaultCluster},
		tmpl,
		server.GetFilesRequest{
			ParameterValues: params,
			TemplateName:    tmpl.Name,
			Profiles:        profiles,
			DefaultHelmRepository: types.NamespacedName{
				Name:      helmRepoName,
				Namespace: DefaultHelmRepoNamespace,
			},
		},
		nil, // FIXME: no create message request, generated resources won't be "editable" in the UI
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get template resources: %w", err)
	}

	files := templateResources.RenderedTemplate
	files = append(files, templateResources.ProfileFiles...)
	files = append(files, templateResources.KustomizationFiles...)

	return files, nil
}

func localHelmRepo(repoName string, settings *cli.EnvSettings) (*repo.Entry, *repo.IndexFile, error) {
	if settings == nil {
		return nil, nil, fmt.Errorf("helm settings missing for repo %s", repoName)
	}

	f, err := repo.LoadFile(settings.RepositoryConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load helm repo file: %w", err)
	}

	r := f.Get(repoName)
	if r == nil {
		return nil, nil, fmt.Errorf("failed to find helm repo %s", repoName)
	}

	indexPath := filepath.Join(settings.RepositoryCache, helmpath.CacheIndexFile(r.Name))
	index, err := repo.LoadIndexFile(indexPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load helm repo index file: %w", err)
	}

	return r, index, nil
}

func fluxHelmRepo(r *repo.Entry) *sourcev1.HelmRepository {
	return &sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: DefaultHelmRepoNamespace,
		},
		Spec: sourcev1.HelmRepositorySpec{
			Interval: metav1.Duration{Duration: 10 * time.Minute},
			URL:      r.URL,
		},
	}
}
