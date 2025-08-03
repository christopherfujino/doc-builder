package main

import (
	"flag"
	"fmt"
	"hash/fnv"
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

func handleArgs(targets []Target1) Target1 {
	const noSuchTarget = "__FIRST_IN_DBC__"
	var targetName string
	flag.StringVar(&targetName, "target", noSuchTarget, "target to build")
	flag.BoolVar(&ensureMode, "ensure", false, "performs a build and returns non-zero exit code if the build changed the target")

	flag.Parse()

	if targetName == noSuchTarget {
		return targets[0]
	}

	for _, target := range targets {
		// TODO strip leading "./"
		if targetName == target.Output {
			return target
		}
	}
	panic(fmt.Sprintf("There is no target named %s", targetName))
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

	var preTargetHash uint64
	var selectTarget = handleArgs(config.Targets)
	if ensureMode {
		var h = fnv.New64a()
		var targetBytes, err = os.ReadFile(selectTarget.Output)
		if err != nil {
			panic(fmt.Sprintf("In ensure mode, but the target %s does not even exist", selectTarget.Output))
		} else {
			Check2(h.Write(targetBytes))
			preTargetHash = h.Sum64()
		}
	}

	didBuild, _, _ := selectTarget.MaybeBuild(Env{
		Targets:   targetsMap,
		Variables: config.Variables,
	})

	if ensureMode {
		if !didBuild {
			panic("Unreachable")
		}
		var h = fnv.New64a()
		var targetBytes, err = os.ReadFile(selectTarget.Output)
		if err != nil {
			panic(fmt.Sprintf("The target %s does not exist", selectTarget.Output))
		}
		Check2(h.Write(targetBytes))
		var postTargetHash = h.Sum64()
		if preTargetHash != postTargetHash {
			// TODO diff
			panic(fmt.Sprintf("The target %s changed after build", selectTarget.Output))
		} else {
			Trace("The target %s did not change after a build\n", selectTarget.Output)
		}
	}

	os.Exit(0)
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
