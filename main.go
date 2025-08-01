package main

import (
	"fmt"
	"os"
)

// Doc Builder Config
const configFileName = "dbc.json"

func Check2[T any](t T, e error) T {
	Check1(e)
	return t
}

func Check1(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	fmt.Println("Hello, world!")

	// Find curring working directory
	var cwd = Check2(os.Getwd())
	// Find config file
	var configFile = findConfigFile(cwd)
	var configBytes = Check2(os.ReadFile(configFile))
	var config = ConfigOfBytes(configBytes)
	fmt.Println(len(config.Targets))

	_, _ = config.Targets[0].MaybeBuild(&config.Env)
	//fmt.Println(env)
}

func findConfigFile(dir string) string {
	var entries = Check2(os.ReadDir(dir))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name() == configFileName {
			return entry.Name()
		}
	}
	panic(
		fmt.Sprintf(
			"The config file %s not found in %s", configFileName, dir,
		),
	)
}
