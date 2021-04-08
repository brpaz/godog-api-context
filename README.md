
# godog-api-context

> A set of reusable step definitions for testing REST APIs with [Godog](https://github.com/DATA-DOG/godog).

![Go version](https://img.shields.io/github/go-mod/go-version/brpaz/godog-api-context?style=for-the-badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/brpaz/godog-api-context?style=for-the-badge)](https://goreportcard.com/report/github.com/brpaz/godog-api-context)
[![CI Status](https://github.com/brpaz/godog-api-context/workflows/CI/badge.svg?style=for-the-badge)](https://github.com/brpaz/godog-api-context/actions)
[![Coverage Status](https://img.shields.io/codecov/c/github/brpaz/godog-api-context/master.svg?style=for-the-badge)](https://codecov.io/gh/brpaz/godog-api-context)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](LICENSE)
[![Commitizen friendly](https://img.shields.io/badge/commitizen-friendly-brightgreen.svg?style=for-the-badge)](http://commitizen.github.io/cz-cli/)
[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg?style=for-the-badge)](https://github.com/semantic-release/semantic-release?style=for-the-badge)

## Pre-requisites

* [godog](https://github.com/cucumber/godog) > 0.10.0

## Usage

To recommended way is to integrate with Godog and Go test as specified in the [Godog documentation](https://github.com/cucumber/godog#running-godog-with-go-test)

```go
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
}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opts)
}

func TestMain(m *testing.M) {
	flag.Parse()
	opts.Paths = flag.Args()

	apiContext := apicontext.New("<base_url>")

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
```

You can see a complete example together with Feature files in [examples folder](examples).

## Available step definitions

`^I set query param "([^"]*)" with value "([^"]*)"$`

`^I set query params to:$`

`^I set header "([^"]*)" with value "([^"]*)"$`

`^I set headers to:$`

`^I send "([^"]*)" request to "([^"]*)" with form body::$`

`^I send "([^"]*)" request to "([^"]*)"$`

`^I send "([^"]*)" request to "([^"]*)" with body:$`

`^The response code should be (\d+)$`

`^The response should be a valid json$`

`^The response should match json:$`

`The response header "([^"]*)" should have value ([^"]*)$`

`^The response should match json schema "([^"]*)"$`

`^The json path "([^"]*)" should have value "([^"]*)"$`

`^wait for  (\d+) seconds$`

`^Store data in scope variable "([^"]*)" with value ([^"]*)`

`^I store the value of response header "([^"]*)" as ([^"]*) in scenario scope$`

`^I store the value of body path "([^"]*)" as "([^"]*)" in scenario scope$`

`^The scenario variable "([^"]*)" should have value "([^"]*)"$`


## Scope Values

This can also store the values from http response body and header and then use in subsequent requests. 
To use the value of scope variable, use this ``pattern: `##(keyname)` without parenthesis``

Example:
```
I store the value of response header "X-AUTH-TOKEN" as token in scenario scope
I set header "X-AUTH-TOKEN" with value "`##token`"
```

This can be used for Authentication headers.

## TODO

* Add steps for setting Cookies

## Contributing

Contributions are what make the open source community such an amazing place to be learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 💛 Support the project

If this project was useful to you in some form, I would be glad to have your support.  It will help to keep the project alive and to have more time to work on Open Source.

The sinplest form of support is to give a ⭐️ to this repo.

You can also contribute with [GitHub Sponsors](https://github.com/sponsors/brpaz).

[![GitHub Sponsors](https://img.shields.io/badge/GitHub%20Sponsors-Sponsor%20Me-red?style=for-the-badge)](https://github.com/sponsors/brpaz)


Or if you prefer a one time donation to the project, you can simple:

<a href="https://www.buymeacoffee.com/Z1Bu6asGV" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: auto !important;width: auto !important;" ></a>

## Author

👤 **Bruno Paz**

* Website: [brunopaz.dev](https://brunopaz.dev)
* Github: [@brpaz](https://github.com/brpaz)
* Twitter: [@brunopaz88](https://twitter.com/brunopaz88)

## 📝 License

Copyright  [Bruno Paz](https://github.com/brpaz).

This project is [MIT](LICENSE) licensed.
