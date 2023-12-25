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
  <a href="https://github.com/kcmvp/gob/blob/main/.github/workflows/workkflow.yml" rel="nofollow">
     <img src="https://img.shields.io/github/actions/workflow/status/kcmvp/gob/workkflow.yml?branch=main" alt="Build" />
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
Whenever a new Golang project is started, it requires a series of installations、different command execution
to complete daily project development work; 
What’s worse is that whenever your switch the development environment, the same work has to be repeated! 
This project is built to solve this problem by providing a method similar to [Maven](https://maven.apache.org/) or [Gradle](https://gradle.com/) in the **Java** ecosystem.

<span id="nav-3"></span>

## Installation
Run below command to install this tool
```go
go install github.com/kcmvp/gob
```
## Features
This tool supply comprehensive help message, you can always get detail information & usage of each command by **-h** flag
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

