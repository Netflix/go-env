# go-env

![Build Status](https://github.com/Netflix/go-env/actions/workflows/build.yml/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/Netflix/go-env.svg)](https://pkg.go.dev/github.com/Netflix/go-env)
[![NetflixOSS Lifecycle](https://img.shields.io/osslifecycle/Netflix/go-expect.svg)]()


Package env provides an `env` struct field tag to marshal and unmarshal environment variables.

## Usage

```go
package main

import (
	"log"
	"time"

	"github.com/Netflix/go-env"
)

type Environment struct {
	Home string `env:"HOME"`

	Jenkins struct {
		BuildId     *string `env:"BUILD_ID"`
		BuildNumber int     `env:"BUILD_NUMBER"`
		Ci          bool    `env:"CI"`
	}

	Node struct {
		ConfigCache *string `env:"npm_config_cache,NPM_CONFIG_CACHE"`
	}

	Extras env.EnvSet

	Duration      time.Duration `env:"TYPE_DURATION"`
	DefaultValue  string        `env:"MISSING_VAR,default=default_value"`
	RequiredValue string        `env:"IM_REQUIRED,required=true"`
	ArrayValue    []string      `env:"ARRAY_VALUE,default=value1|value2|value3"`
}

func main() {
	var environment Environment
	es, err := env.UnmarshalFromEnviron(&environment)
	if err != nil {
		log.Fatal(err)
	}
	// Remaining environment variables.
	environment.Extras = es

	// ...

	es, err = env.Marshal(&environment)
	if err != nil {
		log.Fatal(err)
	}

	home := "/tmp/edgarl"
	cs := env.ChangeSet{
		"HOME":         &home,
		"BUILD_ID":     nil,
		"BUILD_NUMBER": nil,
	}
	es.Apply(cs)

	environment = Environment{}
	if err = env.Unmarshal(es, &environment); err != nil {
		log.Fatal(err)
	}

	environment.Extras = es
}
```

This will initially throw an error if `IM_REQUIRED` is not set in the environment as part of the env struct validation.

This error can be resolved by setting the `IM_REQUIRED` environment variable manually in the environment or by setting it in the 
code prior to calling `UnmarshalFromEnviron` with:
```go
os.Setenv("IM_REQUIRED", "some_value")
```

## Custom Marshaler/Unmarshaler

There is limited support for dictating how a field should be marshaled or unmarshaled. The following example
shows how you could marshal/unmarshal from JSON

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	
	"github.com/Netflix/go-env"
)

type SomeData struct {
    SomeField int `json:"someField"`
}

func (s *SomeData) UnmarshalEnvironmentValue(data string) error {
    var tmp SomeData
	if  err := json.Unmarshal([]byte(data), &tmp); err != nil {
		return err
	}
	*s = tmp 
	return nil
}

func (s SomeData) MarshalEnvironmentValue() (string, error) {
	bytes, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

type Config struct {
    SomeData *SomeData `env:"SOME_DATA"`
}

func main() {
	var cfg Config
	if _, err := env.UnmarshalFromEnviron(&cfg); err != nil {
		log.Fatal(err)
	}

    if cfg.SomeData != nil && cfg.SomeData.SomeField == 42 {
        fmt.Println("Got 42!")
    } else {
        fmt.Printf("Got nil or some other value: %v\n", cfg.SomeData)
    }

    es, err := env.Marshal(&cfg)
	if err != nil {
		log.Fatal(err)
	}
    fmt.Printf("Got the following: %+v\n", es)
}
```
