// +build mage

package main

import (
	"os"
	"os/exec"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	name = "hued"
)

func init() {
	os.Setenv("MAGEFILE_VERBOSE", "true")
}

func Env() map[string]string {
	version := "development"
	hash, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	return map[string]string{
		"VERSION":     version,
		"COMMIT_HASH": hash,
		"BUILD_DATE":  time.Now().Format("2006-01-02T15:04:05Z0700"),
	}
}

func Clean() {
	os.RemoveAll("rice-box.go")
	os.RemoveAll(name)
}

func Dev() error {
	mg.Deps(Prerequisites)
	mg.Deps(Clean)
	mg.Deps(Generate)
	ldflags := "--ldflags=-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.Hash=${COMMIT_HASH} "
	env := Env()
	env["GOOS"] = "darwin"
	env["GOARCH"] = "amd64"
	env["CGO_ENABLED"] = "0"
	return sh.RunWith(env, "go", "build", ldflags, "-o", name)
}

func Build() error {
	mg.Deps(Prerequisites)
	mg.Deps(Clean)
	mg.Deps(Generate)
	ldflags := "--ldflags=-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.Hash=${COMMIT_HASH} "
	env := Env()
	env["GOOS"] = "linux"
	env["GOARCH"] = "amd64"
	env["CGO_ENABLED"] = "0"
	return sh.RunWith(env, "go", "build", ldflags, "-o", name)
}

func Generate() error {
	err := exec.Command("go", "generate").Run()
	if err != nil {
		return err
	}
	return nil
}

func Prerequisites() error {
	err := exec.Command("go", "get", "github.com/GeertJohan/go.rice").Run()
	if err != nil {
		return err
	}
	err = exec.Command("go", "get", "github.com/GeertJohan/go.rice/rice").Run()
	if err != nil {
		return err
	}
	err = exec.Command("go", "mod", "tidy").Run()
	if err != nil {
		return err
	}
	return nil
}
