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











// Configuration


## How can I add a global or common behaviour for the command?

### silent

### export 



## How generated manifests are kept up to date beyond cli lifecycle?



