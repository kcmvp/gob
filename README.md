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
or [Gradle](https://gradle.com/) in the **Java** ecosystem. Please refer [document](./docs/document.md) for details

<span id="nav-3"></span>

## Features

1. **Everything is a plugin, simple yet powerful !**
2. Build a tool chain and workflow without a line.
3. Better user experience

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

3. Make some changes in your source code and try to commit the code, you will see below screenshot.
> <img src="https://github.com/kcmvp/gob/blob/main/docs/commit_hook.gif" height="245" width="425">


4: Check project dependency
```shell
gob deps
```
> <img src="https://github.com/kcmvp/gob/blob/main/docs/dependency_tree.png" height="245" width="425">




## FAQ

