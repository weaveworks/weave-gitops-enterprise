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
## How can I add a new step?

Steps are the units of business logic for bootstarpping. you will need to add one if you want to extend what the bootstrap workflow 
is able to do. For example, release 1 supported bootstrapping for flux environments, while release 2 supported bootstrapping for 
non-flux environment via bootstrapping flux as part of the workflow. This is an example of step.

To add a new step follow these steps as guidance. We will be using an outside-in approach and `admin password` step as example:

1. `Acceptance`: add or extend acceptance testing [test case](../../cmd/gitops/app/bootstrap/cmd_acceptance_test.go) to define how the experience looks like
for the users.
```
		{
			name: "journey flux exists: should bootstrap with valid arguments",
			flags: []string{kubeconfigFlag,
			    ...
				"--password=admin123",
```
2. `User Input`: add the user flags to either `rootCmd` (if global) or [cmd/gitops/app/bootstrap/cmd.go](../../cmd/gitops/app/bootstrap/cmd.go) (if local):

```
rootCmd.PersistentFlags().StringVarP(&options.Username, "username", "u", "", "The Weave GitOps Enterprise username for authentication can be set with `WEAVE_GITOPS_USERNAME` environment variable")
rootCmd.PersistentFlags().StringVarP(&options.Password, "password", "p", "", "The Weave GitOps Enterprise password for authentication can be set with `WEAVE_GITOPS_PASSWORD` environment variable")
```
 
3. `Configuration`: 

**Create a configuration struct for the step**

It should include user input configuration and values from the existing state.
```go
type ClusterUserAuthConfig struct {
    Username         string
    Password         string
    ExistCredentials bool
}
```
**Propagate User Input via Config Builder**

[cmd/gitops/app/bootstrap/cmd.go](../../cmd/gitops/app/bootstrap/cmd.go) 

```
		c, err := steps.NewConfigBuilder().
			WithPassword(opts.Password).
						Build()
```

**Build the configuration**

This should include to resolve any value required so the configuration logic is encapsulated in this layer. 

For example, for `ClusterUserAuthConfig` we require to configure the step differently in case cluster user auth 
credentials already exist via `isExistingAdminSecret`. We resolve it in this layer:

```
// NewClusterUserAuthConfig creates new configuration out of the user input and discovered state
func NewClusterUserAuthConfig(password string, client k8s_client.Client) (ClusterUserAuthConfig, error) {
	...
	return ClusterUserAuthConfig{
		...
		ExistCredentials: isExistingAdminSecret(client),
	}, nil
}


func (cb *ConfigBuilder) Build() (Config, error) {
...
	clusterUserAuthConfig, err := NewClusterUserAuthConfig(cb.password, kubeHttp.Client)
	if err != nil {
		return Config{}, fmt.Errorf("cannot configure cluster user auth:%v", err)
	}
	return Config{
		ClusterUserAuth:         clusterUserAuthConfig,
	}, nil
}
```

4. `Step`: create the step and add it as part of the workflow [pkg/bootstrap/bootstrap.go](../../pkg/bootstrap/bootstrap.go)

```
// Bootstrap initiated by the command runs the WGE bootstrap workflow
func Bootstrap(config steps.Config) error {

	adminCredentials, err := steps.NewAskAdminCredsSecretStep(config.ClusterUserAuth, config.Silent)
	if err != nil {
		return fmt.Errorf("cannot create ask admin creds step: %v", err)
	}

	// TODO have a single workflow source of truth and documented in https://docs.gitops.weave.works/docs/0.33.0/enterprise/getting-started/install-enterprise/
	var steps = []steps.BootstrapStep{
	    ...
		adminCredentials,
		...
	}
    ...
}
```
Identify the different scenarios that the steps should support. Common ones are interactive and non-ineractive session, 
create or update scenarios, etc ... This would be useful for:

**Add Test Cases**

Create a unit test [admin_password_test.go](../../pkg/bootstrap/steps/admin_password_test.go) for the step that includes
the contract of the step based on the different scenarios.

For example, cluster user auth step includes case for support create and update scenarios for interactive and non-interactive sessions:

```
func TestAskAdminCredsSecretStep_Execute(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() (BootstrapStep, Config)
		config     Config
		wantOutput []StepOutput
	}{
		{
			name: "should create cluster user non-interactive",
			setup: func() (BootstrapStep, Config) {
				config := MakeTestConfig(t, Config{})
				step, err := NewAskAdminCredsSecretStep(config.ClusterUserAuth, true)
				assert.NoError(t, err)
				return step, config
			},
			wantOutput: []StepOutput{
				{
					Name: "cluster-user-auth",
					Type: "secret",
					Value: v1.Secret{
						ObjectMeta: metav1.ObjectMeta{Name: "cluster-user-auth", Namespace: "flux-system"},
					},
				},
			},
		},

```

**NewXXStep function**

This function should contain the logic needed to determine what inputs we required from the user given the
configuration coming upstream. For example, for `admin password` we would need to ask the user for input in 
case that credentials already exist in the cluster (update scenarios).

```
func NewAskAdminCredsSecretStep(config ClusterUserAuthConfig, silent bool) (BootstrapStep, error) {
	inputs := []StepInput{}
	if !silent {
		// current state layer
		if !config.ExistCredentials {
			// insert
			if config.Password == "" {
				inputs = append(inputs, getPasswordInput)
			}
		} else {
			// update
			inputs = append(inputs, getPasswordWithExistingAndUserInput)
		}
	}
	return BootstrapStep{
		Name:  "user authentication",
		Input: inputs,
		Step:  createCredentials,
	}, nil
}
```

**Implement Step**

You have defined the step function, in our case `createCredentials` in the previous:

```
return BootstrapStep{
		Name:  "user authentication",
		Input: inputs,
		Step:  createCredentials,
	}, nil
```

This function should only contain specific business logic. For our example, the business logic to create or update 
the cluster user credentials

```
func createCredentials(input []StepInput, c *Config) ([]StepOutput, error) {
	
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(c.ClusterUserAuth.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	data := map[string][]byte{
		"username": []byte(defaultAdminUsername),
		"password": encryptedPassword,
	}
	c.Logger.Actionf("dashboard admin username: %s is configured", defaultAdminUsername)

	secret := corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      adminSecretName,
			Namespace: WGEDefaultNamespace,
		},
		Data: data,
	}
	c.Logger.Successf(secretConfirmationMsg)

```

Note: If you find yourself adding common behaviour in this function think on where within [step.go](../../pkg/bootstrap/steps/step.go) should go.

### Style suggestions for steps

**Inputs**

- We usually prefix input names with `in` prefix (short for input) to distinguish these constants from everything else.

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

## Default Behaviours (default value in inputs)

CLI take the decisions that considered safe to user by using the information provided by user in which no mutation could happen on the user's cluster.

The default values in the step input will be used while silent mode is on by providing `-s`, `--silent`

Examples:
- Using existing credentials this will not replace the user's data and it's safe
- Not to install extra controllers unless provided otherwise
- Not to install OIDC unless provided otherwise

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
- `GIT_PRIVATEKEY_PASSWORD`: password protecting access to private key
- `GIT_REPO_URL_SSH`: git ssh url for the repo wge configuration repo.
- `GIT_REPO_URL_SSH_NO_SCHEME`: git ssh url for the repo wge configuration repo without scheme like `git@github.com:weaveworks/cli-dev.git` 
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

Field [`Enabled`](https://github.com/weaveworks/weave-gitops-enterprise/blob/80667a419c286ee7d45178b639e36a2015533cb6/pkg/bootstrap/steps/ask_bootstrap_flux.go#L14) is added to the step input to allow/disallow this input from being processed

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