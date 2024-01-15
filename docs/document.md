## About gob.yaml configuration
To unleash the powerful capabilities of gob, it is necessary to familiarize yourself with and understand its configuration file: gob.yaml. 
This file is to gob what pom.xml is to Maven, or settings.gradle.kts is to Gradle. 
All settings about the project are configured in this file. The gob program itself reads this file to perform the corresponding operations. 
By combining different configurations, you can almost achieve any powerful function you want. 
The following configuration is the configuration of the [gob.yaml](../gob.yaml) project itself (it is eating its own dog food)
This configuration file is divided into two sections: 

- plugins 
- exec. 

They respectively define the plugins used in the project and when they will be invoked. 
Any tool can be referenced as a plugin in gob, just keep the basic principle in mind: define before use, the same is true in gob! (Same as coding)

```yaml
exec:
    commit-msg-hook: ^#[0-9]+:\s*.{10,}$
    pre-commit-hook:
        - lint
        - test
    pre-push-hook:
        - test
plugins:
    golangci-lint:
        alias: lint
        args: run ./...
        url: github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
    gotestsum:
        alias: test
        args: --format testname -- -coverprofile=target/cover.out ./...
        url: gotest.tools/gotestsum@v1.11.0
```

### Define a plugin

```yaml
    golangci-lint:
        alias: lint
        args: run ./...
        url: github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
```
above is a plugin declaration, which define golangci-lint as a plugin in the project. You don't need to manually edit `gob.yaml` to add a plugin.
below command would take all stuffs for you.

```shell 
gob plugin install xxxxxx
``` 
Below are the meanings of each section
- golangci-lint : plugin's unique name
- alias: lint :  alias of this command, you can use `gob lint` to invoke this plugin
- args: run ./... : The parameters passed to `gob lint`
- url: github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2 : Use golangci-lint@v1.55.2 as the plugin
> Compare to `go install`,`gob plugin install` would keep the version as suffix of the binary name. it's an enhanced version `go install`

### Use a plugin
There are **two** ways to use a defined plugin. Below is the first way, which will be invoked from an external program. Here are the git hooks.
- commit-msg
- pre-commit
- pre-push

As git local hooks are common requirements, gob implements them as built-in functionalities.
```yaml
exec:
    commit-msg-hook: ^#[0-9]+:\s*.{10,}$ // commit message must start with '#' then follows an issue number, the message length must bigger than 10 characters 
    pre-commit-hook:  // execute gob lint & gob test during pre-commit phrase. An error will stop the commit
        - lint
        - test
    pre-push-hook: // execue gob test during pre-push phrase. An error will stop the push
        - test
```
The second way to use declared plugins is invoking plugin directly by using gob. for example if you want to call `golangci-lint@v1.55.2` you can
issue below command
```shell 
gob lint
```

## Built-in commands

> principles apply to all commands: 
> every command must be executed from project root directory

### init
```shell
gob init
```
1. Initialize gob for the project, this is the first command you need to execute.
2. This command would set up built-in plugins
   3. golangci-lint (always with the latest version, you can change the version in `gob.yaml`)
   4. gotestsum
5. Generate a default `.golangci.yaml` in project root directory
6. Generate `gob.yaml` in project root directory
7. Set up git local hooks if current project has been version controlled

### build
```shell
gob build
```
1. This command would build binaries in the project.(If there are more than one main method in main package)
2. Final binaries will be built in the ${project_home}/target folder and named the same as go source name which contains main method in main package

### clean
```shell
gob clean
```
This command will clean all stuffs in the ${project_home}/target folder

### test
```shell
gob test
```
This command will test the project and generate coverage report

### lint
```shell
gob lint
```
This command will run lint against project based on the configuration. 
> For best practice, lint only check the changed files
```yaml
new-from-rev: HEAD
```

### gob plugin install (Enhanced version `go install`)
```shell
gob plugin install 
```
1. Install a tool as gob plugin and gob would generate the corresponding part in `gob.yaml`
2. If you don't specify version, gob would install the latest version of tool
> you can update `alias` and `version` at anytime. 

for example

```shell
gob plugin install github.com/golangci/golangci-lint/cmd/golangci-lint
```

### gob plugin list
```shell
gob plugin list
```
list all the definied plugins in project










