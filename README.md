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
		BuildId     string `env:"BUILD_ID"`
		BuildNumber int    `env:"BUILD_NUMBER"`
		Ci          bool   `env:"CI"`
	}
}

func main() {
  m, err := env.EnvironToMap(os.Environ())
  if err != nil {
    log.Fatal(err)
  }

  var environment env.Environment
  err = env.Unmarshal(m, &environment)
  if err != nil {
    log.Fatal(err)
  }

  // ...
}
```
