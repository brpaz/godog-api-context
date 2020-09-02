// Package apicontext defines common godog step definitions for testing REST APIs
package apicontext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/cucumber/gherkin-go"
	"github.com/cucumber/godog"
	"github.com/xeipuuv/gojsonschema"
)

// DefaultSchemasPath The defaults path to json schema files for validating the responses
const defaultSchemasPath = "schemas"

// ApiContext The main struct
type ApiContext struct {
	baseURL         string
	jSONSchemasPath string
	debug           bool
	client          *http.Client
	headers         map[string]string
	queryParams     map[string]string
	lastResponse    *ApiResponse
	lastRequest     *http.Request
}

// ApiResponse Struct that wraps an API response.
// It contains common accessed fields like Status Code and the Payload as well as access to the raw http.Response object
type ApiResponse struct {
	StatusCode  int
	Body        string
	ResponseObj *http.Response
}

// New Creates a new instance of the API Context
func New(baseURL string) *ApiContext {
	return &ApiContext{
		baseURL:         baseURL,
		client:          &http.Client{},
		headers:         map[string]string{},
		queryParams:     map[string]string{},
		debug:           false,
		jSONSchemasPath: defaultSchemasPath,
	}
}

// WithBaseURL Configures context base URL
func (ctx *ApiContext) WithBaseURL(url string) *ApiContext {
	ctx.baseURL = url

	return ctx
}

// WithDebug Configures debug mode
func (ctx *ApiContext) WithDebug(debug bool) *ApiContext {
	ctx.debug = debug

	return ctx
}

// WithJSONSchemasPath Specifies the path to JSON schema files for doing response validation
func (ctx *ApiContext) WithJSONSchemasPath(path string) *ApiContext {
	ctx.jSONSchemasPath = path
	return ctx
}

// InitializeScenario this function should be called when starting the Test suite, to register the available steps.
func (ctx *ApiContext) InitializeScenario(s *godog.ScenarioContext) {
	s.BeforeScenario(ctx.reset)

	s.Step(`^I set header "([^"]*)" with value "([^"]*)"$`, ctx.ISetHeaderWithValue)
	s.Step(`^I set headers to:$`, ctx.ISetHeadersTo)
	s.Step(`^I send "([^"]*)" request to "([^"]*)" with body:$`, ctx.ISendRequestToWithBody)
	s.Step(`^I send "([^"]*)" request to "([^"]*)"$`, ctx.ISendRequestTo)
	s.Step(`^I set query param "([^"]*)" with value "([^"]*)"$`, ctx.ISetQueryParamWithValue)
	s.Step(`^I set query params to:$`, ctx.ISetQueryParamsTo)
	s.Step(`^The response code should be (\d+)$`, ctx.TheResponseCodeShouldBe)
	s.Step(`^The response should be a valid json$`, ctx.TheResponseShouldBeAValidJSON)
	s.Step(`^The response should match json:$`, ctx.TheResponseShouldMatchJSON)
	s.Step(`^The response header "([^"]*)" should have value ([^"]*)$`, ctx.TheResponseHeaderShouldHaveValue)
	s.Step(`^The response should match json schema "([^"]*)"$`, ctx.TheResponseShouldMatchJsonSchema)
	s.Step(`^The json path "([^"]*)" should have value "([^"]*)"$`, ctx.TheJSONPathShouldHaveValue)
}

// reset Reset the internal state of the API context
func (ctx *ApiContext) reset(*godog.Scenario) {
	ctx.headers = make(map[string]string)
	ctx.queryParams = make(map[string]string)
	ctx.lastResponse = nil
	ctx.lastRequest = nil
}

// ISetHeadersTo This step sets the request headers using a datatable as source.
// It allows to define multiple headers at the same time.
func (ctx *ApiContext) ISetHeadersTo(data *gherkin.DataTable) error {
	for i := 0; i < len(data.Rows); i++ {
		ctx.headers[data.Rows[i].Cells[0].Value] = data.Rows[i].Cells[1].Value
	}

	return nil
}

// ISetHeaderWithValue Step that add a new header to the current request.
func (ctx *ApiContext) ISetHeaderWithValue(name string, value string) error {
	ctx.headers[name] = value
	return nil
}

// ISetQueryParamWithValue Adds a new query param to the request
func (ctx *ApiContext) ISetQueryParamWithValue(name string, value string) error {
	ctx.queryParams[name] = value
	return nil
}

// ISetQueryParamsTo Set query params from a Data Table
func (ctx *ApiContext) ISetQueryParamsTo(data *gherkin.DataTable) error {
	for i := 0; i < len(data.Rows); i++ {
		ctx.queryParams[data.Rows[i].Cells[0].Value] = data.Rows[i].Cells[1].Value
	}

	return nil
}

// ISendRequestTo Sends a request to the specified endpoint using the specified method.
func (ctx *ApiContext) ISendRequestTo(method, uri string) error {
	reqURL := fmt.Sprintf("%s%s", ctx.baseURL, uri)

	req, err := http.NewRequest(method, reqURL, nil)

	if err != nil {
		return err
	}

	// Add headers to request
	for name, value := range ctx.headers {
		req.Header.Set(name, value)
	}

	// Add query string to request
	q := req.URL.Query()
	for name, value := range ctx.queryParams {
		q.Add(name, value)
	}

	req.URL.RawQuery = q.Encode()

	ctx.logRequest(req)

	ctx.lastRequest = req
	resp, err := ctx.client.Do(req)

	if err != nil {
		return err
	}

	ctx.logResponse(resp)

	body, err2 := ioutil.ReadAll(resp.Body)

	if err2 != nil {
		return err2
	}

	ctx.lastResponse = &ApiResponse{
		StatusCode:  resp.StatusCode,
		ResponseObj: resp,
		Body:        string(body),
	}

	return nil
}

// ISendRequestToWithBody Send a request with json body. Ex: a POST request.
func (ctx *ApiContext) ISendRequestToWithBody(method, uri string, requestBody *godog.DocString) error {

	reqURL := fmt.Sprintf("%s%s", ctx.baseURL, uri)

	var jsonStr = []byte(requestBody.Content)
	req, err := http.NewRequest(method, reqURL, bytes.NewBuffer(jsonStr))

	for name, value := range ctx.headers {
		req.Header.Set(name, value)
	}

	if err != nil {
		return err
	}

	ctx.logRequest(req)

	ctx.lastRequest = req
	resp, err := ctx.client.Do(req)

	if err != nil {
		return err
	}

	ctx.logResponse(resp)

	body, err2 := ioutil.ReadAll(resp.Body)

	if err2 != nil {
		return err2
	}

	ctx.lastResponse = &ApiResponse{
		StatusCode:  resp.StatusCode,
		ResponseObj: resp,
		Body:        string(body),
	}

	return nil
}

// TheResponseCodeShouldBe Check if the http status code of the response matches the specified value.
func (ctx *ApiContext) TheResponseCodeShouldBe(code int) error {
	if code != ctx.lastResponse.StatusCode {
		if ctx.lastResponse.StatusCode >= 400 {
			return fmt.Errorf("expected Response code to be: %d, but actual is: %d, Response message: %s", code, ctx.lastResponse.StatusCode, ctx.lastResponse.Body)
		}
		return fmt.Errorf("expected Response code to be: %d, but actual is: %d", code, ctx.lastResponse.StatusCode)
	}
	return nil
}

// TheResponseShouldBeAValidJSON checks if the response is a valid JSON.
func (ctx *ApiContext) TheResponseShouldBeAValidJSON() error {
	var data interface{}
	return json.Unmarshal([]byte(ctx.lastResponse.Body), &data)
}

// TheJSONPathShouldHaveValue Validates if the json object have the expected value at the specified path.
func (ctx *ApiContext) TheJSONPathShouldHaveValue(path string, value interface{}) error {
	var jsonData map[string]interface{}

	if err := json.Unmarshal([]byte(ctx.lastResponse.Body), &jsonData); err != nil {
		return err
	}

	res, err := jsonpath.Get(path, jsonData)

	if err != nil {
		return err
	}

	if res != value {
		return fmt.Errorf("expected json %v, does not match actual: %v", res, value)
	}

	return nil
}

// TheResponseShouldMatchJSON Check that response matches the expected JSON.
func (ctx *ApiContext) TheResponseShouldMatchJSON(body *godog.DocString) error {
	actual := strings.Trim(ctx.lastResponse.Body, "\n")

	expected := body.Content

	match, err := isEqualJson(actual, expected)
	if err != nil {
		return err
	}
	if !match {
		return fmt.Errorf("expected json %s, does not match actual: %s", expected, actual)
	}
	return nil
}

// TheResponseShouldMatchJsonSchema Checks if the response matches the specified JSON schema
func (ctx *ApiContext) TheResponseShouldMatchJsonSchema(path string) error {

	path = strings.Trim(path, "/")

	schemaPath := fmt.Sprintf("%s/%s", ctx.jSONSchemasPath, path)

	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		return fmt.Errorf("JSON schema file does not exist: %s", schemaPath)
	}

	schemaContents, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("cannot open json schema file: %s", err)
	}

	schemaLoader := gojsonschema.NewStringLoader(string(schemaContents))
	documentLoader := gojsonschema.NewStringLoader(ctx.lastResponse.Body)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)

	if err != nil {
		return err
	}

	if !result.Valid() {
		var schemaErrors []string
		for _, error := range result.Errors() {
			schemaErrors = append(schemaErrors, error.String())
		}

		return fmt.Errorf("The document is not valid according to the specified schema %s\n %v", path, schemaErrors)
	}

	return nil
}

// TheResponseHeaderShouldHaveValue Verify the value of a response header
func (ctx *ApiContext) TheResponseHeaderShouldHaveValue(name string, expectedValue string) error {
	actualValue := ctx.lastResponse.ResponseObj.Header.Get(name)

	if actualValue != expectedValue {
		return fmt.Errorf("expected header to have value %s. actual : %s", expectedValue, actualValue)
	}

	return nil
}

// logRequest Helper function to log the request
func (ctx *ApiContext) logRequest(request *http.Request) {
	if !ctx.debug {
		return
	}

	dump, _ := httputil.DumpRequestOut(request, true)
	log.Println(string(dump))
}

// // logResponse Helper function to log the response
func (ctx *ApiContext) logResponse(response *http.Response) {
	if !ctx.debug {
		return
	}

	dump, _ := httputil.DumpResponse(response, true)
	log.Println(string(dump))
}
