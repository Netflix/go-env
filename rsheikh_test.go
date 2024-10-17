package env

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
)

type SomeData struct {
	SomeField int `json:"someField"`
}

func (s *SomeData) UnmarshalEnvironmentValue(data string) error {
	var tmp SomeData
	if err := json.Unmarshal([]byte(data), &tmp); err != nil {
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

func TestRsheikh(t *testing.T) {
	var cfg Config
	if _, err := UnmarshalFromEnviron(&cfg); err != nil {
		log.Fatal(err)
	}

	if cfg.SomeData != nil && cfg.SomeData.SomeField == 42 {
		fmt.Println("Got 42!")
	} else {
		fmt.Printf("Got nil or some other value: %v\n", cfg.SomeData)
	}

	es, err := Marshal(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Got the following: %+v\n", es)
}
