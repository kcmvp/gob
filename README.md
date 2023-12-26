<p align="center">
Golang Project Boot
  <br/>
  <br/>
  <a href="https://github.com/kcmvp/gob/blob/master/LICENSE">
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

## Table of Contents

<details>
  <summary>Click to Open/Close the directory listing</summary>

- [1. Table of Contents](#nav-1)
- [2. Introduction](#nav-2)
- [3. Features](#nav-3)
- [4. FAQ](#nav-4)

</details>

<span id="nav-2"></span>

## Introduction

Although the Golang programming ecosystem is becoming more and more mature,
these tools and frameworks exist independently to solve specific problems.
Whenever a new Golang project is started, it requires a series of initialization;
What’s worse is that whenever your switch the development environment, same process have to be repeated!
This project is built to solve this problem by providing a method similar to [Maven](https://maven.apache.org/)
or [Gradle](https://gradle.com/) in the **Java** ecosystem.

<span id="nav-3"></span>

## Features

**Everything is a plugin, simple yet powerful !**

You just need to use **three** commands to achieve whatever you want

1. Initialization project with

```shell
gob init
```

2. Install a tool as a plugin

```shell
gob install github.com/golangci/golangci-lint/cmd/golangci-lint lint
``` 

- This command install the **latest**  [golangci-lint](https://golangci-lint.run/) as a plugin with alias **lint**
- You can also install a tool with specified version as

```shell
gob install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.0 lint
```

- Compare to go default ```go install```, gob support multiple version tool installation. eg

```shell
ls -althr golangci-lint-v1.55*
-rwxr-xr-x@ 1 kcmvp  staff    40M Dec 25 16:08 golangci-lint-v1.55.1
-rwxr-xr-x  1 kcmvp  staff    41M Dec 25 16:10 golangci-lint-v1.55.0
```

This will make your project is build with constant tools set

3. Execute the tool as a gob plugin(execute golangci-lint)

```shell
gob lint
```

4. Run `gob -h` get comprehensive and beauty help information

## Quick Start

- Install `gob` with below command

```go
    go install github.com/kcmvp/gob
```

- Navigate to project directory, initialize project with below command

```go
  gob init
```

This command will create a configuration named `gob.yaml` in the project root directory as below:

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
    command: run, ./...
    url: github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.1
```

1. The `exec` section is designed to be executed by external; the `init` command will setup three
   git local hook: `commit-msg-hook`, `pre-commit-hook` and `pre-push-hook`
2. The `plugins` section define all the plugins(tools) used by this project. **Any tool can be installed as a gob plugin
   **,
   `init` command would install [golangci-lint](https://golangci-lint.run/) as a gob `plugin`

## Features

This tool supply comprehensive help message, you can always get detail information & usage of each command by **-h**
flag
The main features mainly categorize as below

#### Build & Package

There are mainly 4 built-in commands for project building: **clean, test, lint, build**.

#### Setup (gob setup)

- Setup Git Hook
- Setup onion architecture
- Scaffold of popular frameworks

#### Plugin(gob plugin)

If you want to chain a tool into your project phrase you can install it as a plugin. for example
gb has builtin **golangci-lint** as the part of git hook.

#### Diagram(on the roadmap)

<span id="nav-4"></span>

## FAQ

