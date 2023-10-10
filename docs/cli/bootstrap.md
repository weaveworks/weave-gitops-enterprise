# Bootstrap cli command 

The same as flux bootstrap, gitopsee bootstrap could be considered as one of the most important and complex commands that we have as part of our cli.

Given the expectations of evolution for this command, this document provides background 
and guidance on the design considerations taken for you to be in a successful extension path.

## Glossary

- Bootstrap: the process of installing weave gitops enterprise app and configure a management cluster.
- Step: each of the bootstrapping stages or activities the workflow goes through. For example, checking entitlements.

## What is the bootstrapping command architecture?

It follows a regular cli structure where:

- [cmd/gitops/app/bootstrap/cmd.go](../../cmd/gitops/app/bootstrap/cmd.go): represents the presentation layer
- [pkg/bootstrap/bootstrap.go](../../pkg/bootstrap/bootstrap.go): domain layer for bootstrapping
- [pkg/bootstrap/steps](../../pkg/bootstrap/steps): domain layer for bootstrapping steps
- [pkg/bootstrap/steps/config.go](../../pkg/bootstrap/steps/config.go): configuration for bootstrapping

## How the bootstrapping workflow looks like?

You could find it in [pkg/bootstrap/bootstrap.go](../../pkg/bootstrap/bootstrap.go) as a sequence of steps:

```go
	var steps = []steps.BootstrapStep{
		steps.CheckEntitlementSecret,
		steps.VerifyFluxInstallation,
		steps.NewSelectWgeVersionStep(config),
		steps.NewAskAdminCredsSecretStep(config),
		steps.NewSelectDomainType(config),
		steps.NewInstallWGEStep(config),
		steps.CheckUIDomainStep,
	}

```

## How configuration works ?

The following chain of responsibility applies for config:

1. Users introduce command flags values [cmd/gitops/app/bootstrap/cmd.go](../../cmd/gitops/app/bootstrap/cmd.go)
2. We use builder pattern for configuration [pkg/bootstrap/steps/config.go](../../pkg/bootstrap/steps/config.go): 
    - builder: so we propagate user flags
    - build: we build the configuration object
3. Configuration is then used to create the workflow steps [pkg/bootstrap/bootstrap.go](../../pkg/bootstrap/bootstrap.go)
```
		steps.NewSelectWgeVersionStep(config),
```
4. Steps use configuration for execution (for example [wge_version.go](../../pkg/bootstrap/steps/wge_version.go))
```
// selectWgeVersion step ask user to select wge version from the latest 3 versions.
func selectWgeVersion(input []StepInput, c *Config) ([]StepOutput, error) {
	for _, param := range input {
		if param.Name == WGEVersion {
			version, ok := param.Value.(string)
			if !ok {
				return []StepOutput{}, errors.New("unexpected error occurred. Version not found")
			}
			c.WGEVersion = version
		}

```
## How can I add a new step?

Follow these indications:

1. Add or extend an existing [test case](../../cmd/gitops/app/bootstrap/cmd_integration_test.go)
2. Add the user flags to [cmd/gitops/app/bootstrap/cmd.go](../../cmd/gitops/app/bootstrap/cmd.go)
3. Add the config to [pkg/bootstrap/steps/config.go](../../pkg/bootstrap/steps/config.go):
   - Add config values to the builder
   - Resolves the configuration business logic in the build function. Ensure that validation happens to fail fast. 
4. Add the step as part of the workflow [pkg/bootstrap/bootstrap.go](../../pkg/bootstrap/bootstrap.go)
5. Add the new step [pkg/bootstrap/steps](../../pkg/bootstrap/steps)


An example could be seen here given `gitops bootstrap`

1. if user passes the flag we use the flag
```go
   cmd.Flags().StringVarP(&flags.username, "username", "u", "", "Dashboard admin username")
```
- this is empty so we go to the next level
2. if not, then ask user in interactive session with a default value
```go
func (c *Config) AskAdminCredsSecret() error {

	if c.Username == "" {
		c.Username, err = utils.GetStringInput(adminUsernameMsg, DefaultAdminUsername)
		if err != nil {
			return err
		}
	}
	
	return nil
}
```
User has not introduce a custom value so we take the custom value

```go
type Config struct {
	Username         string
	Password         string
	KubernetesClient k8s_client.Client
	WGEVersion       string
	UserDomain       string
	Logger           logger.Logger
}

```

## Error management 

// TBD with waleed

- Errors should provide the user with
   - step that failed
   - reason of the failure
   - how the user could recover from the failure
- Return the error in the method
- if the error is meaningless create a custom one that provides the user:
```go
if err != nil {
	return errors.New(fluxInstallationErrorMsgFormat)
}

```

## Logging Actions

// TBD with waleed


## How to 

### How can I add a global behaviour around input management?

For example `silent` flag that affects how we resolve inputs: if silent, the we take defaults and we dont ask 
the user. if not silent and no configuration exists, we ask via input the user. 

// TBD

### How can I add a global behaviour around output management?
See the following examples:

- https://github.com/weaveworks/weave-gitops-enterprise/tree/cli-dry-run
- https://github.com/weaveworks/weave-gitops-enterprise/tree/cli-export


## How generated manifests are kept up to date beyond cli lifecycle?

This will be addressed in the following [ticket](https://github.com/weaveworks/weave-gitops-enterprise/issues/3405)

Currently, we do the following in the install wge step

/Users/enekofb/projects/github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/commands/install_wge.go

we have a valuesFile struct with the helm values that
we will be writing:

```go
type valuesFile struct {
	Config             ValuesWGEConfig        `json:"config,omitempty"`
	Ingress            map[string]interface{} `json:"ingress,omitempty"`
	TLS                map[string]interface{} `json:"tls,omitempty"`
	PolicyAgent        map[string]interface{} `json:"policy-agent,omitempty"`
	PipelineController map[string]interface{} `json:"pipeline-controller,omitempty"`
	GitOpsSets         map[string]interface{} `json:"gitopssets-controller,omitempty"`
	EnablePipelines    bool                   `json:"enablePipelines,omitempty"`
	EnableTerraformUI  bool                   `json:"enableTerraformUI,omitempty"`
	Global             global                 `json:"global,omitempty"`
	ClusterController  clusterController      `json:"cluster-controller,omitempty"`
}
```

that we build in the command logic

```go
values := valuesFile{
	Ingress: constructIngressValues(c.UserDomain),
	TLS: map[string]interface{}{
	"enabled": false,
	},
	GitOpsSets:        gitOpsSetsValues,
	EnablePipelines:   true,
	ClusterController: clusterController,
}
```
that we build with the following code

```go
wgeHelmRelease, err := constructWGEhelmRelease(values, c.WGEVersion)
if err != nil {
	return err
}

if err := utils.CreateFileToRepo(wgeHelmReleaseFileName, wgeHelmRelease, pathInRepo, wgeHelmReleaseCommitMsg); err != nil {
	return err
}
```




