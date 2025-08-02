package main

import (
	"flag"
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

func selectTarget(targets []Target1) Target1 {
	const noSuchTarget = "__NO_SUCH_TARGET__"
	var targetPtr = flag.String("target", noSuchTarget, "usage")

	flag.Parse()

	if targetPtr == nil {
		panic("NPE")
	}

	if *targetPtr == noSuchTarget {
		return targets[0]
	}

	for _, target := range targets {
		// TODO strip leading "./"
		if *targetPtr == target.Output {
			return target
		}
	}
	panic(fmt.Sprintf("There is no target named %s", *targetPtr))
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

	if len(config.Targets) == 0 {
		fmt.Fprintf(
			os.Stderr,
			"The config file at \"%s\" does not have any \"targets\" specified\n",
			configFile,
		)
		os.Exit(1)
	}

	var selectTarget = selectTarget(config.Targets)

	didBuild, _, _ := selectTarget.MaybeBuild(Env{
		Targets:   targetsMap,
		Variables: config.Variables,
	})
	//fmt.Println(env)
	var exitCode = 0
	if didBuild {
		// Calling `doc-builder` in a CI script is a way to ensure everything is
		// built, lest it returns 1
		exitCode = 1
	}

	os.Exit(exitCode)
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
