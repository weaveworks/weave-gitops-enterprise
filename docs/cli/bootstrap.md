# Bootstrap cli command 

The same as flux bootstrap, gitopsee bootstrap could be considered as one of the most important and complex comamnds that we have as part of our cli.

Given the expectations of evolution for this command, this document provides background 
and guidance on the design considerations taken for you to be in a succesful extension path.

## How can I add a new step?

1. Add a new file into [pkg/bootstrap/commands](/Users/enekofb/projects/github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/commands)
Example /pkg/bootstrap/commands/admin_password.go
you expect to have the following:
- Constants with the user messages.
- Constants with default configuration
- Step function like: 
```go
func (c *Config) AskAdminCredsSecret() error {
	
...
	// search for existing admin credentials in secret cluster-user-auth
	secret, err := utils.GetSecret(c.KubernetesClient, adminSecretName, WGEDefaultNamespace)
	if secret != nil && err == nil {
		existingCreds := utils.GetConfirmInput(existingCredsMsg)
		if existingCreds == confirmYes {
			return nil
		} else {
			c.Logger.Warningf(existingCredsExitMsg, adminSecretName, WGEDefaultNamespace)
			os.Exit(0)
		}
	}
...
}

```
- configuration: setup your configuration structs in the [config file](/Users/enekofb/projects/github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/commands/config.go)
- utils: leverage utils package for helper functions

2. Add your step to the [bootstrap command flow](/Users/enekofb/projects/github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/cmd.go)

```go

func bootstrap(opts *config.Options, logger logger.Logger) error {
	...
	config := commands.Config{}
	config.Username = flags.username
	config.Password = flags.password
	config.WGEVersion = flags.version
	config.KubernetesClient = kubernetesClient
	config.Logger = logger

	..

	if err := config.AskAdminCredsSecret(); err != nil {
		return err
	}

	...

	return nil
}
```
## How bootstrap configuration works?

We support the following levels of configuration:

1. User introduces configuration via flags
[bootstrapFlags](/Users/enekofb/projects/github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/cmd.go)
`gitops bootstrap --username=wego-admin`

```go
type bootstrapFlags struct {
	username string
	password string
	version  string
}
```
2. User input via interactive dialog

3.Default values 

The resolutions of the configuartion based on the previous 
three levels as follow:

1. if user passes the flag we use the flag 
2. if not, then ask user in interactive session with a default value
3. user introduces custom value
4. otherwise default configuration is taken

example could be seen here given `gitops bootstrap`

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

## How can I add a global or common behaviour for the command?

### silent // example of input common behaviour

1. add `silent` to the bootstrap flags struct
2. add `silent` to the config struct so it could be passed downstream
3. options:

    a)  go to your steps implementation `AskAdminCredsSecret` and add the 
    custom logic to handle the behaviour.
    ```go
        // handle silent behaviour 
        if c.silent &&  c.Username != "" {
            //use default value 
            
        }
    ```
    
    b) extend the user input method with the `silent`
    
    ```go
        if c.Username == "" {
            c.Username, err = utils.GetStringInput(adminUsernameMsg, DefaultAdminUsername, silent)
            if err != nil {
                return err
            }
        }
    ```

### export // example of output common behaviour

1. add `export` to the bootstrap flags struct
2. add `export` to the config struct so it could be passed downstream
3. options:

   a)  go to your steps implementation `AskAdminCredsSecret` and add the
   custom logic to handle the behaviour.
    ```go
   // handle export behaviour 
	if c.export {
		//write to stdout
	}
	else {
		if err := utils.CreateSecret(c.KubernetesClient, adminSecretName, WGEDefaultNamespace, data); err != nil {
			return err
		}	
	}
    ```

   b) extend the user input method with the `export`
    ```go
   // handle export behaviour 
		if err := utils.CreateSecret(c.KubernetesClient, adminSecretName, WGEDefaultNamespace, data, export); err != nil {
			return err
		}
    ```

### other option ~> using generic struct 

```go

type BootstrapStep struct {
   Name  string
   Input func(map,config) config
   Transform func(config)
   Output     func()
}

var (
   //generic
   AskAdminCredentialsStep = BootstrapStep{
      Input: defaultInputStep,
      Output:      defaultOutpuStep,
      Transform:     askAdminCredentialsStep,
   }
   WgeInstallCredentialsStep = BootstrapStep{
      Input: wgeInstallInputStep,
      Output:      defaultOutpuStep,
      Transform:     askAdminCredentialsStep,
   }
})
func defaultInput(params map, config *Config)  *Config {
   // process the config
   if config.silent {
   }
}
func defaultOutput(*Config)  *Config {
   // process the config
   if config.export {
   }
}

```

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
## How should i handle errors in the command and/or step?







