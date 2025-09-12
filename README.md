# go-env (with prefix support)

![Build Status](https://github.com/dimahluodba96/go-env/actions/workflows/build.yml/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/dimahluodba96/go-env.svg)](https://pkg.go.dev/github.com/dimahluodba96/go-env)
[![NetflixOSS Lifecycle](https://img.shields.io/osslifecycle/Netflix/go-expect.svg)]()

> **Note**: This is a fork of [Netflix's go-env](https://github.com/Netflix/go-env) with additional support for **prefix directives** in struct tags.
Package env provides an `env` struct field tag to marshal and unmarshal environment variables.

## Usage

```go
package main

import (
 "log"
 "time"

 "github.com/dimahluodba96/go-env"
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

## 🆕 Prefix Support

This fork adds support for **prefix directives** that automatically prepend prefixes to environment variable names for nested structs. This is particularly useful for grouping related configuration settings.

### Basic Prefix Usage

```go
package main

import (
 "log"
 "os"

 "github.com/dimahluodba96/go-env"
)

// Config holds all configuration for our application
type Config struct {
 Server   ServerConfig   `env:"prefix=SERVER_"`
 Database DatabaseConfig `env:"prefix=DATABASE_"`
 JWT      JWTConfig      // No prefix - uses env var names as-is
}

// ServerConfig holds server configuration
type ServerConfig struct {
 Host string `env:"HOST,default=localhost"`     // Reads SERVER_HOST
 Port string `env:"PORT,default=8080"`          // Reads SERVER_PORT
 Mode string `env:"MODE,default=debug"`         // Reads SERVER_MODE
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
 Host     string `env:"HOST,default=localhost"`     // Reads DATABASE_HOST
 Port     string `env:"PORT,default=5432"`          // Reads DATABASE_PORT
 Username string `env:"USERNAME,default=postgres"`  // Reads DATABASE_USERNAME
 Password string `env:"PASSWORD"`                   // Reads DATABASE_PASSWORD
 Database string `env:"DATABASE,default=myapp"`     // Reads DATABASE_DATABASE
}

// JWTConfig holds JWT configuration (no prefix)
type JWTConfig struct {
 Secret     string `env:"JWT_SECRET"`                    // Reads JWT_SECRET
 ExpiryTime string `env:"JWT_EXPIRY,default=24h"`        // Reads JWT_EXPIRY
}

func main() {
 // Set some environment variables for demonstration
 os.Setenv("SERVER_HOST", "api.example.com")
 os.Setenv("SERVER_PORT", "3000")
 os.Setenv("DATABASE_HOST", "db.example.com")
 os.Setenv("DATABASE_USERNAME", "admin")
 os.Setenv("DATABASE_PASSWORD", "secret123")
 os.Setenv("JWT_SECRET", "super-secret-key")

 var config Config
 _, err := env.UnmarshalFromEnviron(&config)
 if err != nil {
  log.Fatal(err)
 }

 // config.Server.Host = "api.example.com" (from SERVER_HOST)
 // config.Server.Port = "3000" (from SERVER_PORT)
 // config.Server.Mode = "debug" (default, SERVER_MODE not set)
 // config.Database.Host = "db.example.com" (from DATABASE_HOST)
 // config.JWT.Secret = "super-secret-key" (from JWT_SECRET)
}
```

### Nested Prefixes

Prefixes can be nested and will accumulate:

```go
type Root struct {
 Level1 Level1Config `env:"prefix=LEVEL1_"`
}

type Level1Config struct {
 Nested Level2Config `env:"prefix=LEVEL2_"`
 Direct string       `env:"DIRECT"`  // Reads LEVEL1_DIRECT
}

type Level2Config struct {
 Value string `env:"VALUE"`  // Reads LEVEL1_LEVEL2_VALUE
}
```

### Key Features

- **Automatic Prefixing**: Environment variables are automatically prefixed for nested structs
- **Backward Compatible**: All existing functionality works exactly as before
- **Flexible**: Mix prefixed and non-prefixed structs in the same configuration
- **Nested Support**: Prefixes accumulate for deeply nested structures
- **Marshal Support**: Prefixes work with both unmarshaling and marshaling

This will initially throw an error if `IM_REQUIRED` is not set in the environment as part of the env struct validation.This error can be resolved by setting the `IM_REQUIRED` environment variable manually in the environment or by setting it in the
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

 "github.com/dimahluodba96/go-env"
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

## Changes from Original

This fork adds the following enhancements to the original Netflix go-env library:

### ✨ New Features

- **Prefix Support**: Use `prefix=VALUE` in struct tags to automatically prepend prefixes to environment variable names
- **Nested Prefixes**: Prefixes accumulate for deeply nested structures
- **Marshal Support**: Prefixes work in both directions (unmarshal and marshal)

### 🔧 Improvements

- **Fixed Deletion Logic**: Environment variables are now properly deleted from EnvSet after processing
- **Enhanced Tests**: Added comprehensive test coverage for prefix functionality

### 📝 Attribution

Original work by Netflix, Inc. under Apache License 2.0.
Prefix functionality added by @dimahluodba96.

---

**License**: Apache License 2.0
