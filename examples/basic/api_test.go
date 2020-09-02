// +build example

package main

import (
	"flag"
	"os"
	"testing"

	apicontext "github.com/brpaz/godog-api-context"
	"github.com/cucumber/godog"
)

var opts = godog.Options{
	Format:        "progress", // can define default values,
	StopOnFailure: true,
	NoColors:      true,
}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opts)
}

func TestMain(m *testing.M) {
	flag.Parse()
	opts.Paths = flag.Args()

	apiContext := apicontext.New("https://httpbin.org").WithDebug(true)

	status := godog.TestSuite{
		Name:                "godogs",
		ScenarioInitializer: apiContext.InitializeScenario,
		Options:             &opts,
	}.Run()

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}
