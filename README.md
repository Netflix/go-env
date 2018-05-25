# go-env

[![Build Status](https://travis-ci.com/Netflix-Skunkworks/go-env.svg?token=qVsub6qcmXEV63K5Cykm&branch=master)](https://travis-ci.com/Netflix-Skunkworks/go-env)

Package env provides an `env` struct field tag to marshal and unmarshal environment variables.

## Usage

```go
package main

import (
  "log"

  env "github.com/Netflix-Skunkworks/go-env"
)

type Environment struct {
	Home string `env:"HOME"`

	Jenkins struct {
		BuildId     *string `env:"BUILD_ID"`
		BuildNumber int    `env:"BUILD_NUMBER"`
		Ci          bool   `env:"CI"`
	}
}

func main() {
  var environment env.Environment
  err = env.UnmarshalFromEnviron(&environment)
  if err != nil {
    log.Fatal(err)
  }

  // ...

  es, err := env.Marshal(environment)
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
}
```
