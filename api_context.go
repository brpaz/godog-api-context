package context

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/xeipuuv/gojsonschema"
)

type ApiContext struct {
	BaseURL         string
	JSONSchemasPath string
	Debug           bool
	client          *http.Client
	headers         map[string]string
	queryParams     map[string]string
	lastResponse    *ApiResponse
	lastRequest     *http.Request
}

type ApiResponse struct {
	StatusCode  int
	Body        string
	ResponseObj *http.Response
}

// NewAPIContext Creates a new API context for testing
func NewAPIContext(s *godog.Suite, baseURL string) *ApiContext {
	ctx := &ApiContext{
		BaseURL:         baseURL,
		client:          &http.Client{},
		headers:         make(map[string]string, 0),
		queryParams:     make(map[string]string, 0),
		Debug:           false,
		JSONSchemasPath: "schemas",
	}

	ctx.registerSteps(s)

	return ctx
}

// Register steps into the suite
func (ctx *ApiContext) registerSteps(s *godog.Suite) {
	s.BeforeScenario(ctx.ResetContext)

	s.Step(`^I send "([^"]*)" request to "([^"]*)"$`, ctx.ISendRequestTo)
	s.Step(`^The response code should be (\d+)$`, ctx.TheResponseCodeShouldBe)
	s.Step(`^The response should match json:$`, ctx.TheResponseShouldMatchJson)
	s.Step(`^I set header "([^"]*)" with value "([^"]*)"$`, ctx.ISetHeaderWithValue)
	//s.Step(`^The json path "([^"]*)" should have value "([^"]*)"$`, ctx.TheJsonPathShouldHaveValue)
	s.Step(`^The response should match json schema "([^"]*)"$`, ctx.TheResponseShouldMatchJsonSchema)
	s.Step(`^I send "([^"]*)" request to "([^"]*)" with body:$`, ctx.ISendRequestToWithBody)
	s.Step(`^The response should be a valid json$`, ctx.TheResponseShouldBeAValidJson)
	s.Step(`^I set query param "([^"]*)" with value "([^"]*)"$`, ctx.ISetQueryParamWithValue)
	s.Step(`^I set headers to:$`, ctx.ISetHeadersTo)
	s.Step(`^I set query params to:$`, ctx.ISetQueryParamsTo)
}

// ResetContexts Reset the internal stored context data
func (ctx *ApiContext) ResetContext(interface{}) {
	ctx.headers = make(map[string]string, 0)
	ctx.queryParams = make(map[string]string, 0)
	ctx.lastResponse = nil
}

// ISetHeadersTo Set headers from a Data Table
func (ctx *ApiContext) ISetHeadersTo(data *gherkin.DataTable) error {
	for i := 0; i < len(data.Rows); i++ {
		ctx.headers[data.Rows[i].Cells[0].Value] = data.Rows[i].Cells[1].Value
	}

	return nil
}

// IAddHeaderWithValue Step that add a new header to the current request.
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
	reqURL := fmt.Sprintf("%s%s", ctx.BaseURL, uri)

	req, _ := http.NewRequest(method, reqURL, nil)

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

	if ctx.Debug {
		requestDump, _ := httputil.DumpRequestOut(req, true)
		log.Printf("New Request:\n%q", requestDump)
	}

	resp, err := ctx.client.Do(req)

	if err != nil {
		return err
	}

	if ctx.Debug {
		dump, _ := httputil.DumpResponse(resp, true)
		log.Printf("Received response:\n%q", dump)
	}

	body, err2 := ioutil.ReadAll(resp.Body)

	if err2 != nil {
		return err2
	}

	ctx.lastRequest = req

	ctx.lastResponse = &ApiResponse{
		StatusCode:  resp.StatusCode,
		ResponseObj: resp,
		Body:        string(body),
	}

	return nil
}

// ISendRequestToWithBody Send a request with json body. Ex: a POST request.
func (ctx *ApiContext) ISendRequestToWithBody(method, uri string, requestBody *gherkin.DocString) error {

	reqURL := fmt.Sprintf("%s%s", ctx.BaseURL, uri)

	var jsonStr = []byte(requestBody.Content)
	req, err := http.NewRequest(method, reqURL, bytes.NewBuffer(jsonStr))

	for name, value := range ctx.headers {
		req.Header.Set(name, value)
	}

	if err != nil {
		return err
	}

	if ctx.Debug {
		requestDump, _ := httputil.DumpRequestOut(req, false)
		log.Printf("New Request:\n%q", requestDump)
	}

	resp, err := ctx.client.Do(req)

	if err != nil {
		return err
	}

	if ctx.Debug {
		dump, _ := httputil.DumpResponse(resp, true)
		log.Printf("Received response:\n%q", dump)
	}

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

// TheResponseShouldBeAValidJson checks if the response is a valid JSON.
func (ctx *ApiContext) TheResponseShouldBeAValidJson() error {
	var data interface{}
	if err := json.Unmarshal([]byte(ctx.lastResponse.Body), &data); err != nil {
		return err
	}

	return nil
}

// TheJsonPathShouldHaveValue Validates if the json object have the expected value at the specified path.
// TODO this method is not working well.
//func (ctx *ApiContext) TheJsonPathShouldHaveValue(path string, value interface{}) error {
//	var jsonData interface{}
//
//	json.Unmarshal([]byte(ctx.lastResponse.Body), &jsonData)
//
//	res, err := jsonpath.JsonPathLookup(jsonData, path)
//
//	if err != nil {
//		return err
//	}
//
//	if res != value {
//		return fmt.Errorf("expected json %v, does not match actual: %v", res, value)
//	}
//
//	return nil
//}

// TheResponseShouldMatchJson Check that response matches the expected JSON.
func (ctx *ApiContext) TheResponseShouldMatchJson(body *gherkin.DocString) error {
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

// TheResponseShouldMatchJsonSchema Checks if the response matches the specified JSON schemctx.
func (ctx *ApiContext) TheResponseShouldMatchJsonSchema(path string) error {

	path = strings.Trim(path, "/")

	schemaPath := fmt.Sprintf("%s/%s", ctx.JSONSchemasPath, path)

	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		return fmt.Errorf("JSON schema file does not exist: %s", schemaPath)
	}

	schemaContents, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("Cannot open json schema file: %s", err)
	}

	schemaLoader := gojsonschema.NewStringLoader(string(schemaContents))
	documentLoader := gojsonschema.NewStringLoader(ctx.lastResponse.Body)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)

	if err != nil {
		return err
	}

	if !result.Valid() {
		fmt.Printf("The document is not valid according to the specified schema %s:", path)
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}

		return errors.New("The document is not valid according to the specified schema")
	}

	return nil
}
