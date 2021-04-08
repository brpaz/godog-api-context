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
	godog.BindCommandLineFlags("godog.", &opts)
}

func TestMain(m *testing.M) {
	flag.Parse()
	opts.Paths = flag.Args()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("header token: %+v",r.URL.Query())
		w.Header().Set("X-Some-Header", "world")
	}))

	defer ts.Close()
	apiContext := apicontext.New(ts.URL).WithDebug(true)

	status := godog.TestSuite{
		Name:                "godogs",
		ScenarioInitializer: apiContext.InitializeScenario,
		Options:             &opts,
	}.Run()
	fmt.Printf("status %d", status)
	os.Exit(status)
}
