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

### Style sugestions for steps

**Inputs**

- We usually prefix input names with `in` prefix (short for input) to distinguish these constants from everything else. 


## Error management 

A bootstrapping error received by the platform engineer shoudl allow:

1. understand the step that has failed
2. the reason and context of the failure
3. the actions to take to recover

To achieve this:

1) At internal layers like `util`, return the err. For example `CreateSecret`:
```
	err := client.Create(context.Background(), secret, &k8s_client.CreateOptions{})
	if err != nil {
		return err
	}

```
2) At step implementation: wrapping error with convenient error message in the step implementation for user like invalidEntitlementMsg. 
These messages will provide extra information that's not provided by errors like contacting sales / information about flux download:

```
	ent, err := entitlement.VerifyEntitlement(strings.NewReader(string(publicKey)), string(secret.Data["entitlement"]))
	if err != nil || time.Now().Compare(ent.IssuedAt) <= 0 {
		return fmt.Errorf("%s: %v", invalidEntitlementSecretMsg, err)
	}

```

Use custom errors when required for better handling like [this](https://github.com/weaveworks/weave-gitops-enterprise/blob/6b1c1db9dc0512a9a5c8dd03ddb2811a897849e6/pkg/bootstrap/steps/entitlement.go#L65)

3) Special case for cases where we could recover from the error and don't need to terminate

for example [here](https://github.com/weaveworks/weave-gitops-enterprise/blob/80667a419c286ee7d45178b639e36a2015533cb6/pkg/bootstrap/steps/flux.go#L39)

flux is not bootstrapped, but in the process we can bootstrap flux. in this case we could log the failure and continue the execution

```go
	out, err := runner.Run("flux", "check")
	if err != nil {
		c.Logger.Failuref("flux installed error: %v. %s", string(out), fluxRecoverMsg)
		return []StepOutput{}, nil
	}
```

## Logging Actions

For sharing progress with the user, the following levels are used:

- `c.Logger.Waitingf()`: to identify the step. or a subtask that's taking a long time. like reconciliation
- `c.Logger.Actionf()`: to identify subtask of a step. like Writing file to repo.
- `c.Logger.Warningf`: to show warnings. like admin creds already existed.
- `c.Logger.Successf`: to show that subtask/step is done successfully.

## Testing

Tend to follow the following levels

### Unit Testing

This level to ensure each component meets their expected contract for the happy and unhappy scenarios.
You will see them in the expected form `*_test.go`

### Integration Testing

This level to ensure some integrations with bootstrapping dependencies like flux, git, etc ... 

We currently have a gap to cover in the following features.

### Acceptance testing 

You could find it in [cmd_acceptance_test.go](../../cmd/gitops/app/bootstrap/cmd_acceptance_test.go) with the aim of
having a small set of bootstrapping journeys that we code for acceptance and regression on the bootstrapping workflow.

Dependencies are:
- flux
- kube cluster via envtest
- git

Environment Variables Required:

Entitlement stage

- `WGE_ENTITLEMENT_USERNAME`: entitlements username  to use for creating the entitlement before running the test.
- `WGE_ENTITLEMENT_PASSWORD`: entitlements password  to use for creating the entitlement before running the test.
- `WGE_ENTITLEMENT_ENTITLEMENT`: valid entitlements token to use for creating the entitlement before running the test.
- `OIDC_CLIENT_SECRET`: client secret for oidc flag
- `GIT_PRIVATEKEY_PATH`: path to the private key to do the git operations.
- `GIT_REPO_URL_SSH`: git ssh url for the repo wge configuration repo.
- `GIT_REPO_URL_HTTPS`: git https url for the repo wge configuration repo.
- `GIT_USERNAME`: git username for testing https auth
- `GIT_PASSWORD`: git password for testing https auth
- `GIT_BRANCH`: git branch for testing with flux bootstrap
- `GIT_REPO_PATH`: git repo path for default cluster for testing with flux bootstrap


Run it via `make cli-acceptance-tests`

## How to 

### How can I add a global behaviour around input management?

For example `silent` flag that affects how we resolve inputs. To be added out of the work in https://github.com/weaveworks/weave-gitops-enterprise/issues/3465

### How can I add a global behaviour around output management?
See the following examples:

- https://github.com/weaveworks/weave-gitops-enterprise/tree/cli-dry-run
- https://github.com/weaveworks/weave-gitops-enterprise/tree/cli-export


## How generated manifests are kept up to date beyond cli lifecycle?

This will be addressed in the following [ticket](https://github.com/weaveworks/weave-gitops-enterprise/issues/3405)

## Enable/Disable one or more input from step inputs

Field [`Enabled`](https://github.com/weaveworks/weave-gitops-enterprise/blob/80667a419c286ee7d45178b639e36a2015533cb6/pkg/bootstrap/steps/ask_bootstrap_flux.go#L14) is added to the step input to allow/disallow this input from being processd

This field should receive a function that takes the step input, config object and returns boolean value 

example:

- step input

	```go
	var bootstrapFLuxQuestion = StepInput{
		Name:    inBootstrapFlux,
		Type:    confirmInput,
		Msg:     bootstrapFluxMsg,
		Enabled: canAskForFluxBootstrap,
	}
	```

- function

	```go
	func canAskForFluxBootstrap(input []StepInput, c *Config) bool {
		return !c.FluxInstallated
	}
	```

This input will be processed only if `Enabled` field is equal to `true`