# go-env

[![Build Status](https://travis-ci.org/Netflix/go-env.svg?branch=master)](https://travis-ci.org/Netflix/go-env)
[![GoDoc](https://godoc.org/github.com/Netflix/go-env?status.svg)](https://godoc.org/github.com/Netflix/go-env)
[![NetflixOSS Lifecycle](https://img.shields.io/osslifecycle/Netflix/go-expect.svg)]()


Package env provides an `env` struct field tag to marshal and unmarshal environment variables.

## Usage

```go
package main

import (
  "log"

  env "github.com/Netflix/go-env"
)

type Environment struct {
  Home string `env:"HOME"`

  Jenkins struct {
    BuildId     *string `env:"BUILD_ID"`
    BuildNumber int    `env:"BUILD_NUMBER"`
    Ci          bool   `env:"CI"`
  }

  Extras env.EnvSet
}

func main() {
  var environment env.Environment
  es, err := env.UnmarshalFromEnviron(&environment)
  if err != nil {
    log.Fatal(err)
  }
  // Remaining environment variables.
  environment.Extras = es

  // ...

  es, err = env.Marshal(environment)
  if err != nil {
    log.Fatal(err)
  }

  cs := env.ChangeSet{
    "HOME": "/tmp/edgarl",
    "BUILD_ID": nil,
    "BUILD_NUMBER": nil,
  }
  es.Apply(cs)

  environment = env.Environment{}
  err = env.Unmarshal(es, &environment)
  if err != nil {
    log.Fatal(err)
  }

  environment.Extras = es
}
```
