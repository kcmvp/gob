<p align="center">
Golang Project Boot
  <br/>
  <br/>
  <a href="https://github.com/kcmvp/gob/blob/main/LICENSE">
    <img alt="GitHub" src="https://img.shields.io/github/license/kcmvp/gob"/>
  </a>
  <a href="https://goreportcard.com/report/github.com/kcmvp/gob">
    <img src="https://goreportcard.com/badge/github.com/kcmvp/gob"/>
  </a>
  <a href="https://pkg.go.dev/github.com/kcmvp/gob">
    <img src="https://pkg.go.dev/badge/github.com/kcmvp/gob.svg" alt="Go Reference"/>
  </a>
  <a href="https://github.com/kcmvp/gob/blob/main/.github/workflows/workflow.yml" rel="nofollow">
     <img src="https://img.shields.io/github/actions/workflow/status/kcmvp/gob/workflow.yml?branch=main" alt="Build" />
  </a>
  <a href="https://app.codecov.io/gh/kcmvp/gob" ref="nofollow">
    <img src ="https://img.shields.io/codecov/c/github/kcmvp/gob" alt="coverage"/>
  </a>

</p>

<span id="nav-1"></span>

<span id="nav-2"></span>

## Introduction

Although the Golang programming ecosystem is becoming more and more mature,
these tools and frameworks exist independently to solve specific problems.
Whenever a new Golang project is started, it requires a series of initialization;
What’s worse is that whenever your switch the development environment, same process have to be repeated!
This project is built to solve this problem by providing a method similar to [Maven](https://maven.apache.org/)
or [Gradle](https://gradle.com/) in the **Java** ecosystem. Please refer [Document](#commands) for details

<span id="nav-3"></span>

## Features

1. **Everything is a plugin, simple yet powerful !**
2. Build a tool chain and workflow without a line.
3. Better user experience

## How gob works
`Gob` takes everything defined in the `gob.yaml` as plugin.
```mermaid
flowchart TD
        Gob --> gob.yaml 
        gob.yaml --> plugin1
        gob.yaml --> plugin2
        gob.yaml --> plugin3
```
You just need to tell `gob` 3W(where,when and what)

1. **Where** : where to download the tool
2. **When** : when to execute to command
2. **What** : what to do with the tool

## Quick Start
1. Install `gob` with below command
```shell
    go install github.com/kcmvp/gob
```
2. Initialize project with below command(in the project home directory)
```shell
  gob init
```
This command will generate two files
>- gob.yaml :  `gob` configuration
>- .golangci.yaml: [golangci-lint](https://golangci-lint.run/) configuration. gob supports `golangci-lint` butilin
>  These two files need to be checked in with your source code

| Git Hooks | Dependency Tree |
|-----------|-----------------|
|<img src="https://github.com/kcmvp/gob/blob/main/docs/commit_hook.gif" height="245" width="425">           |<img src="https://github.com/kcmvp/gob/blob/main/docs/dependency_tree.png" height="245" width="425">                 |


## Commands 
- Build related commands
  - [gob init](#gob-init)
  - [gob build](#gob-build)
  - [gob clean](#gob-clean)
  - [gob test](#gob-test)
  - [gob lint](#gob-lint)
  - [gob deps](#gob-deps)
- Plugin related commands
  - [gob plugin install](#gob-plugin-install)
  - [gob plugin list](#gob-plugin-list)

### gob init
Initialize gob for the project, it will do following initializations 
1. generate file `gob.yaml`
2. generate file `.golangci.yaml`, which is the configuration for [golangci-lint](https://github.com/golangci/golangci-lint)
3. setup `git hooks` if project in the source control.
   4. commit-msg
   5. pre-commit
   6. pre-push
```shell
gob init
```
`gob.yaml`
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
        alias: lint #When : when issue `gob lint`
        args: run ./... #What: execute `golangci-lint run ./...`
        url: github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2 #Where: where to download the plugin
    gotestsum:
        alias: test
        args: --format testname -- -coverprofile=target/cover.out ./...
        url: gotest.tools/gotestsum@v1.11.0
```
in most cases you don't need to edit the configuration manually. you can achieve this by [plugin commands](#gob-plugin-install) 

### gob build
### gob clean
### gob test
### gob lint
### gob deps
### gob plugin install
### gob plugin list


## FAQ

