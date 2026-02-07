package main

import (
	"encoding/json"
	"errors"
)

type Config struct {
	Targets   []Target1
	Variables map[string]string
}

func ConfigOfBytes(data []byte) (Config, error) {
	var c Config
	Check1(json.Unmarshal(data, &c))
	if c.Variables == nil {
		c.Variables = map[string]string{}
	}
	for _, target := range c.Targets {
		if target.Output == "" {
			return Config{}, errors.New("Target does not have an output configured")
		}
	}
	return c, nil
}
