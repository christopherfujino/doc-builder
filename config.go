package main

import (
	"encoding/json"
)

type Config struct {
	Targets   []Target1
	Variables map[string]string
}

func ConfigOfBytes(data []byte) Config {
	var c Config
	Check1(json.Unmarshal(data, &c))
	if c.Variables == nil {
		c.Variables = map[string]string{}
	}
	return c
}
