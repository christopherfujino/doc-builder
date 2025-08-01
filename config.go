package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"
)

type Config struct {
	Targets []Target
	Env     map[string]string
}

func ConfigOfBytes(data []byte) Config {
	var c Config
	Check1(json.Unmarshal(data, &c))
	return c
}

type Target struct {
	Source string
	Inputs []Target
	Output string
}

func (t Target) Age() time.Time {
	var stat = Check2(os.Stat(t.Output))
	return stat.ModTime()
}

func normalizePath(path string) string {
	var buf = strings.Builder{}
	for _, c := range path {
		switch c {
		case '.':
			fallthrough
		case ' ':
			fallthrough
		case '/':
			fallthrough
		case '\\':
			buf.WriteRune('_')
		default:
			buf.WriteRune(c)
		}
	}
	return buf.String()
}

func (t Target) Build(env *map[string]string) *map[string]string {
	fmt.Printf("Building %s...\n", t.Output)
	var sourceBytes = Check2(os.ReadFile(t.Source))
	if strings.HasPrefix(t.Output, "#") {
		switch t.Output {
		case "#env":
			(*env)[normalizePath(t.Source)] = strings.TrimSpace(string(sourceBytes))
		default:
			panic(fmt.Sprintf("Unknown magic var %s", t.Output))
		}
	} else {
		var template, err = template.New(t.Source).Parse(string(sourceBytes))
		if err != nil {
			panic(
				fmt.Sprintf("Error interpolating %s\n\n%s\n\nCurrent env was:\n%v", t.Source, err.Error(), *env),
			)
		}

		var w = Check2(os.Create(t.Output))
		Check1(template.Execute(w, env))
	}
	return env
}

func (t Target) MaybeBuild(env *map[string]string) (bool, *map[string]string) {
	if env == nil {
		env = &(map[string]string{})
	}
	fmt.Printf("Evaluating if %s needs a build...\n", t.Output)
	var needsBuild = false

	if len(t.Inputs) == 0 {
		t.Build(env)
		fmt.Println("Early return")
		return true, env
	}

	for _, input := range t.Inputs {
		var didBuild bool
		didBuild, env = input.MaybeBuild(env)
		if didBuild {
			needsBuild = true
		}
	}

	if needsBuild {
		env = t.Build(env)
	}

	return needsBuild, env
}
