package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
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
	Template      string
	templateCache *string
	Inputs        []string
	Output        string
	Filter        string
}

func (t *Target1) templateString() *string {
	if t.templateCache != nil {
		return t.templateCache
	}

	var templateBytes, err = os.ReadFile(t.Template)
	if err != nil {
		var msg = fmt.Sprintf("The template file %s does not exist!\n\n%s", t.Template, err.Error())
		panic(msg)
	}

	var s = string(templateBytes)

	t.templateCache = &s

	return &s
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
	var templateString = *t.templateString()
	// https://stackoverflow.com/questions/49933684/prevent-no-value-being-inserted-by-golang-text-template-library
	template, err := template.New(t.Template).Option("missingkey=error").Parse(templateString)
	if err != nil {
		panic(
			fmt.Sprintf("Error interpolating %s\n\n%s\n\nCurrent env was:\n%v", t.Template, err.Error(), env.Variables),
		)
	}

	var w = Check2(os.Create(t.Output))
	err = template.Execute(w, env.Variables)
	if err != nil {
		// We need to delete the output because it's likely corrupted, with a fresh
		// timestamp
		Check1(os.Remove(t.Output))
		panic(err)
	} else {
		Check1(w.Close())
	}
	return env
}

func (t Target1) MaybeBuild(env Env) (bool, Env, time.Time) {
	Trace("Evaluating if %s needs a build...\n", t.Output)
	var needsBuild = false
	var thisTime, ok = t.Age()
	if !ok {
		Trace("Target %s needs to be built because it does not exist\n", t.Output)
		needsBuild = true
	}

	if t.Template != "" {
		var templateTarget = Target2{Filename: t.Template}
		var templateTime, templateExists = templateTarget.Age()
		if templateExists == false {
			// TODO check if it's an empty string
			panic(fmt.Sprintf("The template \"%s\" for the target \"%s\" does not exist!", t.Template, t.Output))
		}

		if templateTime.After(thisTime) {
			Trace("Target %s needs to be built because template %s is newer\n", t.Output, t.Template)
			needsBuild = true
		}
	}
	if len(t.Inputs) > 0 {
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
				Trace("Target %s needs to be built because of input %s\n", t.Output, inputName)
				needsBuild = true
			}
		}
	}

	if needsBuild {
		// If we're not a template
		if t.Template == "" {
			// Check file already exists
			var _, err = os.Stat(t.Output)
			if err != nil {
				// TODO give a friendly message
				panic(err)
			}
			// we don't actually call .Build(), the interesting work happens in this method
		} else {
			env = t.Build(env)
		}
	}

	// TODO optimize, if we're not an input to another target, we can skip this
	var bytes = Check2(os.ReadFile(t.Output))
	var src string

	if t.Filter != "" {
		Trace("Applying the filter \"%s\" to %s\n", t.Filter, t.Output)
		var cmd = exec.Command("/bin/sh", "-c", t.Filter)
		var stdin = Check2(cmd.StdinPipe())
		var stdout = Check2(cmd.StdoutPipe())
		stdin.Write(bytes)
		Check1(stdin.Close())
		var err = cmd.Start()
		if err != nil {
			panic("TODO")
		}
		var readBuffer = make([]byte, 256)
		var stringBuffer = strings.Builder{}
		for {
			n, err := stdout.Read(readBuffer)
			if err == io.EOF {
				if n > 0 {
					stringBuffer.Write(readBuffer[:n])
				}
				break
			} else if err != nil {
				panic(err)
			}
			Check2(stringBuffer.Write(readBuffer[:n]))
		}
		Check1(cmd.Wait())

		src = strings.TrimSpace(stringBuffer.String())
	} else {
		src = strings.TrimSpace(string(bytes))
	}
	env.Variables[normalizePath(t.Output)] = src

	if needsBuild {
		fmt.Printf("Built target \"%s\"\n", t.Output)
	}
	return needsBuild, env, thisTime
}
