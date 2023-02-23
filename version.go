package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

var (
	//go:embed version.gotmpl
	versionTemplate string
)

func printVersion() error {
	t, err := template.New("version").Parse(versionTemplate)
	if err != nil {
		return err
	}

	data, err := buildVersionData()
	if err != nil {
		return err
	}

	out := new(strings.Builder)
	if err := t.Execute(out, data); err != nil {
		return err
	}

	fmt.Fprint(os.Stderr, out.String())

	return nil
}

func buildVersionData() (map[string]string, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, fmt.Errorf("error retrieving build info")
	}

	data := map[string]string{
		"ProgramName": progName,
		"Version":     info.Main.Version,
		"GoVersion":   info.GoVersion,
	}

	for _, s := range info.Settings {
		switch s.Key {
		case "GOOS":
			data["OS"] = s.Value
		case "GOARCH":
			data["Arch"] = s.Value
		case "vcs.modified":
			data["VCSDirty"] = s.Value
		case "vcs.revision":
			data["GitSha"] = s.Value[0:8]
		case "vcs.time":
			// 2022-05-17T21:41:36Z
			t, err := time.Parse(time.RFC3339, s.Value)
			if err != nil {
				log.Fatal(err)
			}
			data["VCSTime"] = t.Local().Format(time.RFC1123Z)
		}
	}

	return data, nil
}
