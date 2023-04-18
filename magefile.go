//go:build mage

package main

import (
	"os"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	name    = "hued"
	version = "1.0.0"
)

func init() {
	os.Setenv("MAGEFILE_VERBOSE", "true")
}

func Env() map[string]string {
	hash, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	return map[string]string{
		"VERSION":     version,
		"COMMIT_HASH": hash,
		"BUILD_DATE":  time.Now().Format("2006-01-02T15:04:05Z0700"),
	}
}

func Clean() {
	os.RemoveAll(name)
}

func Build() error {
	mg.Deps(Clean)
	ldflags := "--ldflags=-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.Hash=${COMMIT_HASH} "
	env := Env()
	env["GOOS"] = "linux"
	env["GOARCH"] = "amd64"
	env["CGO_ENABLED"] = "0"
	return sh.RunWith(env, "go", "build", ldflags, "-o", name)
}
