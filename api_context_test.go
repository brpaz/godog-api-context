package context

import (
	"encoding/json"
	"github.com/DATA-DOG/godog/gherkin"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/godog"
	"github.com/stretchr/testify/assert"
)

func TestApiContext_NewAPIContextWithOptions(t *testing.T) {
	s := &godog.Suite{}

	ctx := NewAPIContextWithOptions(s, &ContextOptions{
		BaseURL:         "https://example.com",
		Debug:           true,
		JSONSchemasPath: "testdata/schemas",
	})

	assert.Equal(t, "https://example.com", ctx.BaseURL)
	assert.True(t, ctx.Debug)
	assert.Equal(t, "testdata/schemas", ctx.JSONSchemasPath)
}

func TestApiContext_NewAPIContext(t *testing.T) {
	s := &godog.Suite{}

	ctx := NewAPIContext(s, "https://example.com")

	assert.Equal(t, "https://example.com", ctx.BaseURL)
	assert.False(t, ctx.Debug)
	assert.Equal(t, DefaultSchemasPath, ctx.JSONSchemasPath)
}

func TestApiContext_ISetHeadersTo(t *testing.T) {
	s := &godog.Suite{}
	ctx := NewAPIContext(s, "http://example.com")

	dt := &gherkin.DataTable{
		Node: gherkin.Node{},
		Rows: []*gherkin.TableRow{
			{
				Node: gherkin.Node{},
				Cells: []*gherkin.TableCell{
					{
						Node:  gherkin.Node{},
						Value: "X-Header-1",
					},
					{
						Node:  gherkin.Node{},
						Value: "value 1",
					},
				},
			},
			{
				Node: gherkin.Node{},
				Cells: []*gherkin.TableCell{
					{
						Node:  gherkin.Node{},
						Value: "X-Header-2",
					},
					{
						Node:  gherkin.Node{},
						Value: "value 2",
					},
				},
			},
		},
	}

	ctx.ISetHeadersTo(dt)

	assert.Equal(t, "value 1", ctx.headers["X-Header-1"])
	assert.Equal(t, "value 2", ctx.headers["X-Header-2"])
}

func TestApiContext_ISetHeaderWithValue(t *testing.T) {
	s := &godog.Suite{}
	ctx := NewAPIContext(s, "http://example.com")
	ctx.ISetHeaderWithValue("Content-Type", "application/json")

	assert.Equal(t, 1, len(ctx.headers))
	assert.Equal(t, "application/json", ctx.headers["Content-Type"])
}

func TestApiContext_ISetQueryParamWithValue(t *testing.T) {
	s := &godog.Suite{}
	ctx := NewAPIContext(s, "http://example.com")
	ctx.ISetQueryParamWithValue("page", "1")

	assert.Equal(t, 1, len(ctx.queryParams))
	assert.Equal(t, "1", ctx.queryParams["page"])
}

func TestApiContext_ISetQueryParamsTo(t *testing.T) {
	s := &godog.Suite{}
	ctx := NewAPIContext(s, "http://example.com")

	dt := &gherkin.DataTable{
		Node: gherkin.Node{},
		Rows: []*gherkin.TableRow{
			{
				Node: gherkin.Node{},
				Cells: []*gherkin.TableCell{
					{
						Node:  gherkin.Node{},
						Value: "q1",
					},
					{
						Node:  gherkin.Node{},
						Value: "v1",
					},
				},
			},
			{
				Node: gherkin.Node{},
				Cells: []*gherkin.TableCell{
					{
						Node:  gherkin.Node{},
						Value: "q2",
					},
					{
						Node:  gherkin.Node{},
						Value: "v2",
					},
				},
			},
		},
	}

	ctx.ISetQueryParamsTo(dt)

	assert.Equal(t, "v1", ctx.queryParams["q1"])
	assert.Equal(t, "v2", ctx.queryParams["q2"])
}

func TestApiContext_ISendRequestTo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := make(map[string]string, 0)
		p["result"] = "success"
		json.NewEncoder(w).Encode(p)
	}))

	defer ts.Close()
	s := &godog.Suite{}

	ctx := NewAPIContext(s, ts.URL)

	ctx.ISendRequestTo("GET", "/")

	assert.NotNil(t, ctx.lastResponse)
	assert.Equal(t, 200, ctx.lastResponse.StatusCode)
	assert.NotNil(t, ctx.TheResponseCodeShouldBe(400))
	assert.Nil(t, ctx.TheResponseShouldBeAValidJSON())
	assert.Nil(t, ctx.TheResponseShouldMatchJson(&gherkin.DocString{
		Content: "{\"result\": \"success\"}",
	}))
}

func TestVerifyResponseHeaderValue(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Some-Header", "hello")
	}))

	defer ts.Close()
	s := &godog.Suite{}

	ctx := NewAPIContext(s, ts.URL)

	ctx.ISendRequestTo("GET", "/")

	assert.Nil(t, ctx.TheResponseHeaderShouldHaveValue("X-Some-Header", "hello"))
	assert.NotNil(t, ctx.TheResponseHeaderShouldHaveValue("non-existing-header", "hello"))
}

func TestApiContext_TheResponseShouldMatchJsonSchema(t *testing.T) {

	p := make(map[string]interface{}, 0)
	p["firstName"] = "Bruno"
	p["lastName"] = "PAZ"
	p["age"] = 30

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Some-Header", "hello")
		json.NewEncoder(w).Encode(p)
	}))

	defer ts.Close()
	s := &godog.Suite{}

	ctx := NewAPIContext(s, ts.URL)
	ctx.JSONSchemasPath = "testdata/schemas"

	ctx.ISendRequestTo("GET", "/")

	assert.Nil(t, ctx.TheResponseShouldMatchJsonSchema("person.json"))
	assert.NotNil(t, ctx.TheResponseShouldMatchJsonSchema("coordinates.json"))
}

func TestJsonPathMatchers(t *testing.T) {

	f, err := ioutil.ReadFile(filepath.Join("testdata", "test_json_path.json"))

	if err != nil {
		t.Error(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(f)
	}))

	defer ts.Close()
	s := &godog.Suite{}

	ctx := NewAPIContext(s, ts.URL)

	ctx.ISendRequestTo("GET", "/")

	assert.NotNil(t, ctx.lastResponse)
	assert.Nil(t, ctx.TheJsonPathShouldHaveValue("$.a", "a"))
	assert.Nil(t, ctx.TheJsonPathShouldHaveValue("$.b", 2.0))
	assert.Nil(t, ctx.TheJsonPathShouldHaveValue("$.c", 3.5))
	assert.Nil(t, ctx.TheJsonPathShouldHaveValue("$.d", true))
}
