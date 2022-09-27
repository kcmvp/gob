
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

Engineering methodology is very important to a project. [Spring Boot](https://spring.io/projects/spring-boot/) did
a good job in **Java** ecosystem. it supplies lots of libraries and best practices for a Java project. With it programmer
only need to focus on system business, it reduces programmer's mental overhead very much. This framework aims to supplies
the same functionalities as [Spring Boot](https://spring.io/projects/spring-boot/) did in Java ecosystem.

As a developer you **DON'T NEED** do any special configuration, you can get below two categories functionalities. It achieves this by reading **go.mod** and do the cumbersome configuration automatically.

- Functionalities supports SLDC. such as build, clean, git hooks and reporting.
- Environment sensitive configuration(Think about data source in Dev, Test, Prd environments) and code generation.

<span id="nav-3"></span>

## Features

Here are just outstanding features, for details please refer to the [documents](https://github.com/kcmvp/gob/wiki)
SLDC
 - [x] Uniform build scripts(Test, Build and delivery) without shell on all platform (build go with go).
In most case, you dont need to write any code to build the project. from both command line as well as CI
 - [x] Git Hook: for code format and quality. it support [golangci-linter](https://golangci-lint.run/) by default and generate beautify report.


<span id="nav-4"></span>

## Document
Detail can be found here  [Document](https://github.com/kcmvp/gob/wiki)

<span id="nav-5"></span>
## FAQ
@todo

