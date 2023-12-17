<p align="center">
Golang Project Boot
  <br/>
  <br/>
  <a href = "https://github.com/kcmvp/gb/blob/main/LICENSE">
   <img src="https://img.shields.io/github/license/kcmvp/gb"/>
  </a>
  <a href="https://goreportcard.com/report/github.com/kcmvp/gb">
    <img src="https://goreportcard.com/badge/github.com/kcmvp/gb" />
  </a>
  <a href="https://pkg.go.dev/github.com/kcmvp/gb">
    <img src="https://pkg.go.dev/badge/github.com/kcmvp/gb.svg" alt="Go Reference"/>
  </a>
  <a href="https://github.com/kcmvp/gb/blob/main/.github/workflows/go.yml" rel="nofollow">
     <img src="https://img.shields.io/github/actions/workflow/status/kcmvp/gb/go.yml?branch=main" alt="Build" />
  </a>
  <a href="https://app.codecov.io/gh/kcmvp/gb" ref="nofollow">
    <img src ="https://img.shields.io/codecov/c/github/kcmvp/gb" alt="coverage"/>
  </a>

</p>

<span id="nav-1"></span>

## Table of Contents

<details>
  <summary>Click me to Open/Close the directory listing</summary>

- [1. Table of Contents](#nav-1)
- [2. Introduction](#nav-2)
- [3. Features](#nav-3)
- [4. Document](#nav-4)
- [5. FAQ](#nav-5)

</details>

<span id="nav-2"></span>

## Introduction
Although the Golang programming ecosystem is becoming more and more mature, 
these tools and frameworks exist independently to solve specific problems. 
Whenever a new Golang project is started, it requires a series of installations、different command execution
to complete daily project development work; 
What’s worse is that whenever your switch the development environment, the same work has to be repeated! 
This project is built to solve this problem by providing a method similar to [Maven](https://maven.apache.org/) or [Gradle](https://gradle.com/) in the **Java** ecosystem.
o
<span id="nav-3"></span>

## Features

Here are just outstanding features, for details please refer to the [documents](https://github.com/kcmvp/gb/wiki)

### [Commands](https://github.com/kcmvp/gb/wiki#commands)

- [x] Uniform build scripts(Test, Build and delivery) without shell on all platform (build go with go).
  In most case, you don't need to write any code to build the project. from both command line as well as CI
- [x] Git Hook: for code format and quality. it supports [golangci-linter](https://golangci-lint.run/) by default and
  generate beautify report.

<span id="nav-4"></span>

## Document

Detail can be found here  [Document](https://github.com/kcmvp/gb/wiki)

<span id="nav-5"></span>

## FAQ

- As we can define server side hooks easily, do I need a local git hook?

  Keep the principle:Don't let the bad smell comes into code repositories.
  As server side hooks happens after code have been pushed into repositories, a local hook can help you prevent issues
  slipping into repository. 
  
- Can I call the script from git server hook or piple line?

  Yes, you can call it. In fact this project's piple line is calling the **builder.go** directly. [builder workflow](https://github.com/kcmvp/gb/blob/main/.github/workflows/build.yml) 


