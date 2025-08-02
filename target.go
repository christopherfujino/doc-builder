package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"
)

type Env struct {
	Targets   map[string]Target
	Variables map[string]string
}

type Target interface {
	Age() (time.Time, bool)
	Build(env Env) Env
	MaybeBuild(env Env) (bool, Env, time.Time)
}

// Specified by config
type Target1 struct {
	Template string
	Inputs []string
	Output string
}

// External dependency
type Target2 struct {
	Filename string
}

func (t Target2) Build(env Env) Env {
	panic("Unreachable")
}

func (t Target2) MaybeBuild(env Env) (bool, Env, time.Time) {
	var age, ok = t.Age()
	if !ok {
		panic(
			fmt.Sprintf(
				"Expected the target %s to exist, but it did not", t.Filename,
			),
		)
	}

	// We report false so dependees can potentially not rebuild,
	// but populate env in case they do
	var sourceBytes = Check2(os.ReadFile(t.Filename))
	env.Variables[normalizePath(t.Filename)] = strings.TrimSpace(string(sourceBytes))

	return false, env, age
}

var epoch = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

func age(path string) (time.Time, bool) {
	var stat, err = os.Stat(path)
	if err != nil {
		return epoch, false
	}
	return stat.ModTime(), true
}

func (t Target2) Age() (time.Time, bool) {
	return age(t.Filename)
}

func (t Target1) Age() (time.Time, bool) {
	return age(t.Output)
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

func (t Target1) Build(env Env) Env {
	fmt.Printf("Building %s...\n", t.Output)
	if t.Template == "" {
		var msg = fmt.Sprintf("The target %s does not have a template to build from\n", t.Output)
		panic(msg)
	}
	var templateBytes, err = os.ReadFile(t.Template)
	if err != nil {
		var msg = fmt.Sprintf("The template file %s does not exist!\n\n%s", t.Template, err.Error())
		panic(msg)
	}

	// https://stackoverflow.com/questions/49933684/prevent-no-value-being-inserted-by-golang-text-template-library
	template, err := template.New(t.Template).Option("missingkey=error").Parse(string(templateBytes))
	if err != nil {
		panic(
			fmt.Sprintf("Error interpolating %s\n\n%s\n\nCurrent env was:\n%v", t.Template, err.Error(), env.Variables),
		)
	}

	var w = Check2(os.Create(t.Output))
	Check1(template.Execute(w, env.Variables))
	return env
}

func (t Target1) MaybeBuild(env Env) (bool, Env, time.Time) {
	fmt.Printf("Evaluating if %s needs a build...\n", t.Output)
	var needsBuild = false
	var thisTime, ok = t.Age()
	if !ok {
		needsBuild = true
	}

	// TODO must check for source!

	if len(t.Inputs) == 0 {
		needsBuild = true
	} else {
		for _, inputName := range t.Inputs {
			var didBuild bool
			var inputTime time.Time
			var input, ok = env.Targets[inputName]
			if !ok {
				input = Target2{Filename: inputName}
			}
			didBuild, env, inputTime = input.MaybeBuild(env)
			if inputTime.Equal(thisTime) {
				// Unclear what to do in this case
				panic("Are you on Windows?!")
			}
			if didBuild || (inputTime.After(thisTime)) {
				needsBuild = true
			}
		}
	}

	if needsBuild {
		env = t.Build(env)
	}

	return needsBuild, env, thisTime
}
