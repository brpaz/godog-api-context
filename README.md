
# godog-api-context

> A set of step definitions for testing REST APIs with [Godog](https://github.com/DATA-DOG/godog)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](LICENSE)
[![Commitizen friendly](https://img.shields.io/badge/commitizen-friendly-brightgreen.svg?style=for-the-badge)](http://commitizen.github.io/cz-cli/)
[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg?style=for-the-badge)](https://github.com/semantic-release/semantic-release?style=for-the-badge)
[![Actions Status](https://github.com/brpaz/godog-api-context/workflows/CI/badge.svg?style=for-the-badge)](https://github.com/brpaz/godog-api-context/actions)

## Pre-requisites

* [godog](https://github.com/DATA-DOG/godog)

## Usage

```go
package main

import (
	"flag"
	apiContext "github.com/brpaz/godog-api-context"
	"os"
	"testing"

	"github.com/DATA-DOG/godog"
)

var opt = godog.Options{
	Output: os.Stdout,
	Format: "progress", // can define default values,
	Strict: false,
}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opt)
}

func TestMain(m *testing.M) {
	flag.Parse()

	opt.Paths = flag.Args()

	status := godog.RunWithOptions("godogs", func(s *godog.Suite) {
		apiContext.NewAPIContext(s, os.Getenv("APP_BASE_URL"))
	}, opt)

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}
```

## Available step definitions

`^I send "([^"]*)" request to "([^"]*)"$`

`^The response code should be (\d+)$`

`^The response should match json:$`

`^I set header "([^"]*)" with value "([^"]*)"$`

`^The response should match json schema "([^"]*)"$`

`^I send "([^"]*)" request to "([^"]*)" with body:$`

`^The response should be a valid json$`

`^I set query param "([^"]*)" with value "([^"]*)"$`

`^I set headers to:$`

`^I set query params to:$`

More to come.


## 🤝 Contributing

Contributions, issues and feature requests are welcome!

## Author

👤 **Bruno Paz**

* Website: [https://github.com/brpaz](https://github.com/brpaz)
* Github: [@brpaz](https://github.com/brpaz)

## 📝 License

Copyright © 2019 [Bruno Paz](https://github.com/brpaz).

This project is [MIT](LICENSE) licensed.