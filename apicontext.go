// Package apicontext defines common step definitions for testing REST APIs using Cucumber and Godog
package apicontext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/cucumber/godog"
	"github.com/xeipuuv/gojsonschema"
)

// The defaults path to json schema files for validating the responses.
const defaultSchemasPath = "schemas"

// ApiContext main struct
type ApiContext struct {
	baseURL         string
	jSONSchemasPath string
	debug           bool
	client          *http.Client
	headers         map[string]string
	queryParams     map[string]string
	lastResponse    *ApiResponse
	lastRequest     *http.Request
	scope           map[string]string
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
		scope:           map[string]string{},
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
	s.Step(`^I send "([^"]*)" request to "([^"]*)" with form body::$`, ctx.ISendRequestToWithFormBody)
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
	s.Step(`^The json path "([^"]*)" should match "([^"]*)"$`, ctx.TheJSONPathShouldMatch)
	s.Step(`^The json path "([^"]*)" should have count "([^"]*)"$`, ctx.TheJSONPathHaveCount)
	s.Step(`^The json path "([^"]*)" should be present"$`, ctx.TheJSONPathShouldBePresent)
	s.Step(`^The response body should contain "([^"]*)"$`, ctx.TheResponseBodyShouldContain)
	s.Step(`^The response body should match "([^"]*)"$`, ctx.TheResponseBodyShouldMatch)
	s.Step(`^I wait for (\d+) seconds$`, ctx.WaitForSomeTime)
	s.Step(`^I store data in scope variable "([^"]*)" with value "([^"]*)"`, ctx.StoreScopeData)
	s.Step(`^I store the value of response header "([^"]*)" as "([^"]*)" in scenario scope$`, ctx.StoreResponseHeader)
	s.Step(`^I store the value of body path "([^"]*)" as "([^"]*)" in scenario scope$`, ctx.StoreJsonPathValue)
	s.Step(`^The scope variable "([^"]*)" should have value "([^"]*)"$`, ctx.TheScopeVariableShouldHaveValue)
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
func (ctx *ApiContext) ISetHeadersTo(dt *godog.Table) error {
	for i := 0; i < len(dt.Rows); i++ {
		ctx.headers[dt.Rows[i].Cells[0].Value] = ctx.ReplaceScopeVariables(dt.Rows[i].Cells[1].Value)
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
	ctx.queryParams[name] = ctx.ReplaceScopeVariables(value)
	return nil
}

// ISetQueryParamsTo Set query params from a Data Table
func (ctx *ApiContext) ISetQueryParamsTo(dt *godog.Table) error {
	for i := 0; i < len(dt.Rows); i++ {
		ctx.queryParams[dt.Rows[i].Cells[0].Value] = ctx.ReplaceScopeVariables(dt.Rows[i].Cells[1].Value)
	}

	return nil
}

// ISendRequestTo Sends a request to the specified endpoint using the specified method.
func (ctx *ApiContext) ISendRequestTo(method, uri string) error {
	reqURL := ctx.buildRequestURL(uri)
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

// ISendRequestToWithFormBody Send a request with json body. Ex: a POST request.
func (ctx *ApiContext) ISendRequestToWithFormBody(method, uri string, requestBodyTable *godog.Table) error {
	reqURL := ctx.buildRequestURL(uri)
	reqBody := &bytes.Buffer{}
	w := multipart.NewWriter(reqBody)

	for i := 0; i < len(requestBodyTable.Rows); i++ {
		key := requestBodyTable.Rows[i].Cells[0].Value
		value := ctx.ReplaceScopeVariables(requestBodyTable.Rows[i].Cells[1].Value)
		typeOfField := requestBodyTable.Rows[i].Cells[2].Value

		var fw io.Writer
		var err error

		if typeOfField == "text" {
			if err = w.WriteField(key, value); err != nil {
				return err
			}
		} else if typeOfField == "file" {
			file, err := os.Open(value)
			if err != nil {
				panic(err)
			}
			_, fileName := filepath.Split(value)
			fw, err = w.CreateFormFile(key, fileName)
			if err != nil {
				return err
			}
			_, err = io.Copy(fw, file)
			if err != nil {
				return err
			}
		}
	}
	contentType := w.FormDataContentType()
	err := w.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, reqURL, bytes.NewReader(reqBody.Bytes()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	for name, value := range ctx.headers {
		req.Header.Set(name, value)
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

// ISendRequestToWithBody Send a request with json body. Ex: a POST request.
func (ctx *ApiContext) ISendRequestToWithBody(method, uri string, requestBody *godog.DocString) error {
	reqURL := ctx.buildRequestURL(uri)
	jsonBody := ctx.ReplaceScopeVariables(requestBody.Content)
	//todo
	var jsonBodyBytes = []byte(jsonBody)
	req, err := http.NewRequest(method, reqURL, bytes.NewBuffer(jsonBodyBytes))

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
func (ctx *ApiContext) TheResponseCodeShouldBe(statusCode int) error {
	if statusCode != ctx.lastResponse.StatusCode {
		return fmt.Errorf("expected status code to be %d, but actual is %d.\n Response body: %s", statusCode, ctx.lastResponse.StatusCode, ctx.lastResponse.Body)
	}
	return nil
}

// TheResponseShouldBeAValidJSON checks if the response is a valid JSON.
func (ctx *ApiContext) TheResponseShouldBeAValidJSON() error {
	var data interface{}
	return json.Unmarshal([]byte(ctx.lastResponse.Body), &data)
}

// TheJSONPathShouldHaveValue Validates if the json object have the expected value at the specified path.
func (ctx *ApiContext) TheJSONPathShouldHaveValue(pathExpr string, expectedValue string) error {
	var jsonData interface{}
	expectedValue = ctx.ReplaceScopeVariables(expectedValue)
	if err := json.Unmarshal([]byte(ctx.lastResponse.Body), &jsonData); err != nil {
		return err
	}

	actualValue, err := jsonpath.Get(pathExpr, jsonData)

	if err != nil {
		return err
	}

	var expectedParsedValue interface{}
	switch reflect.TypeOf(actualValue).Kind() {
	case reflect.Bool:
		expectedParsedValue, err = strconv.ParseBool(expectedValue)

		if err != nil {
			return err
		}

	case reflect.Float64:
		expectedParsedValue, err = strconv.ParseFloat(expectedValue, 64)

		if err != nil {
			return err
		}
	case reflect.Int32:
		expectedParsedValue, err = strconv.ParseInt(expectedValue, 10, 64)

		if err != nil {
			return err
		}
	case reflect.Int64:
		expectedParsedValue, err = strconv.ParseInt(expectedValue, 10, 64)

		if err != nil {
			return err
		}

	default:
		expectedParsedValue = expectedValue
	}

	if actualValue != expectedParsedValue {
		return fmt.Errorf("expected json path to have value %v but it is %v", expectedParsedValue, actualValue)
	}

	return nil
}

// TheJSONPathShouldMatch Validates Checks if the the value from the specified json path matches the specified pattern.
func (ctx *ApiContext) TheJSONPathShouldMatch(pathExpr string, pattern string) error {
	var jsonData interface{}

	if err := json.Unmarshal([]byte(ctx.lastResponse.Body), &jsonData); err != nil {
		return err
	}

	value, err := jsonpath.Get(pathExpr, jsonData)

	if err != nil {
		return err
	}

	match, err := regexp.MatchString(pattern, value.(string))

	if err != nil {
		return err
	}

	if !match {
		return fmt.Errorf("%s does not match: %s", value, pattern)
	}

	return nil
}

// TheJSONPathShouldBePresent checks if the specified json path exists in the response body
func (ctx *ApiContext) TheJSONPathShouldBePresent(pathExpr string) error {
	var jsonData interface{}

	if err := json.Unmarshal([]byte(ctx.lastResponse.Body), &jsonData); err != nil {
		return err
	}

	value, err := jsonpath.Get(pathExpr, jsonData)

	if err != nil {
		return err
	}

	if value == nil {
		return fmt.Errorf("the json path %s was not present in the response", pathExpr)
	}

	return nil
}

// TheJSONPathHaveCount Validates if the field at the specified json path have the expected length
func (ctx *ApiContext) TheJSONPathHaveCount(pathExpr string, expectedCount int) error {
	var jsonData interface{}

	if err := json.Unmarshal([]byte(ctx.lastResponse.Body), &jsonData); err != nil {
		return err

	}

	value, err := jsonpath.Get(pathExpr, jsonData)

	if err != nil {
		return err
	}

	s := reflect.ValueOf(value)

	if s.Kind() != reflect.Slice {
		return fmt.Errorf("the json path %s is not an array. Found %v", pathExpr, value)
	}

	if s.Len() != expectedCount {
		return fmt.Errorf("the value %v doenst have count %d but %d", value, expectedCount, s.Len())
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

// TheResponseBodyShouldContain Checks if the response body contains the specified string
func (ctx *ApiContext) TheResponseBodyShouldContain(s string) error {
	bodyContent := strings.Trim(ctx.lastResponse.Body, "\n")

	if !strings.Contains(bodyContent, ctx.ReplaceScopeVariables(s)) {
		return fmt.Errorf("%s does not contain %s", bodyContent, s)
	}
	return nil
}

// TheResponseBodyMatch Checks if the response body matches the specified pattern
func (ctx *ApiContext) TheResponseBodyShouldMatch(pattern string) error {

	bodyContents := ctx.lastResponse.Body
	match, err := regexp.MatchString(pattern, bodyContents)

	if err != nil {
		return err
	}

	if !match {
		return fmt.Errorf("%s does not match pattern: %s", bodyContents, pattern)
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

	schemaLoader := gojsonschema.NewBytesLoader(schemaContents)
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

		return fmt.Errorf("The response is not valid according to the specified schema %s\n %v", path, schemaErrors)
	}

	return nil
}

// TheResponseHeaderShouldHaveValue Verify the value of a response header
func (ctx *ApiContext) TheResponseHeaderShouldHaveValue(name string, expectedValue string) error {
	actualValue := ctx.lastResponse.ResponseObj.Header.Get(name)

	if actualValue != ctx.ReplaceScopeVariables(expectedValue) {
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

// buildRequestURL concatenates the given URI and the base URL after replacing scope variables.
func (ctx *ApiContext) buildRequestURL(uri string) string {
	uri = ctx.ReplaceScopeVariables(uri)
	return ctx.baseURL + uri
}

// WaitForSomeTime halt for some time.
func (ctx *ApiContext) WaitForSomeTime(timeToWait int) error {
	duration := time.Duration(timeToWait) * time.Second
	time.Sleep(duration)
	return nil
}

// StoreScopeData Store data in scope map.
func (ctx *ApiContext) StoreScopeData(scopeKeyName string, value string) error {
	ctx.scope[scopeKeyName] = value
	return nil
}

// StoreResponseHeader Store header value to scope map.
func (ctx *ApiContext) StoreResponseHeader(name string, scopeKeyName string) error {
	actualValue := ctx.lastResponse.ResponseObj.Header.Get(name)
	ctx.scope[scopeKeyName] = actualValue
	return nil
}

// StoreJsonPathValue Store value from json body path to scope map.
func (ctx *ApiContext) StoreJsonPathValue(pathExpr string, scopeKeyName string) error {
	var jsonData interface{}

	if err := json.Unmarshal([]byte(ctx.lastResponse.Body), &jsonData); err != nil {
		return err
	}

	actualValue, err := jsonpath.Get(pathExpr, jsonData)

	if err != nil {
		return err
	}
	ctx.scope[scopeKeyName] = actualValue.(string)
	return nil
}

// TheScopeVariableShouldHaveValue Verify the value of a scope variable
func (ctx *ApiContext) TheScopeVariableShouldHaveValue(scopeKeyName string, expectedValue string) error {
	if ctx.scope[scopeKeyName] != expectedValue {
		return fmt.Errorf("expected scope variable to have value %s. actual : %s", expectedValue, ctx.scope[scopeKeyName])
	}

	return nil
}

func (ctx *ApiContext) ReplaceScopeVariables(data string) string {
	re := regexp.MustCompile("`##(.*?)`")
	matches := re.FindStringSubmatch(data)
	if len(matches) < 1 || len(matches[1]) < 1 {
		return data
	}
	dataToReplace := matches[0]
	scopeKey := matches[1]
	return strings.Replace(data, dataToReplace, ctx.scope[scopeKey], 1)
}
