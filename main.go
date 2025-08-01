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
	// Find curring working directory
	var cwd = Check2(os.Getwd())
	// Find config file
	var configFile = findConfigFile(cwd)
	var configBytes = Check2(os.ReadFile(configFile))
	var config = ConfigOfBytes(configBytes)
	var targetsMap = map[string]Target{}
	for _, target := range config.Targets {
		targetsMap[target.Output] = target
	}

	_, _, _ = config.Targets[0].MaybeBuild(Env{
		Targets: targetsMap,
		Variables: &config.Variables,
	})
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
