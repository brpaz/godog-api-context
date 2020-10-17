package apicontext

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v10"
	"github.com/stretchr/testify/assert"
)

func setupTestContext() *ApiContext {
	ctx := New("https://example.com").
		WithDebug(true).
		WithJSONSchemasPath("testdata/schemas")

	return ctx
}
func TestApiContext_New(t *testing.T) {

	ctx := setupTestContext()

	assert.Equal(t, "https://example.com", ctx.baseURL)
	assert.True(t, ctx.debug)
	assert.Equal(t, "testdata/schemas", ctx.jSONSchemasPath)
}

func TestApiContext_ISetHeadersTo(t *testing.T) {

	ctx := setupTestContext()

	dt := &godog.Table{
		Rows: []*messages.PickleStepArgument_PickleTable_PickleTableRow{
			{
				Cells: []*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{
					{
						Value: "X-Header-1",
					},
					{
						Value: "value 1",
					},
				},
			},
			{
				Cells: []*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{
					{
						Value: "X-Header-2",
					},
					{
						Value: "value 2",
					},
				},
			},
		},
	}

	err := ctx.ISetHeadersTo(dt)

	assert.Nil(t, err)
	assert.Equal(t, "value 1", ctx.headers["X-Header-1"])
	assert.Equal(t, "value 2", ctx.headers["X-Header-2"])
}

func TestApiContext_ISetHeaderWithValue(t *testing.T) {
	ctx := setupTestContext()
	err := ctx.ISetHeaderWithValue("Content-Type", "application/json")

	assert.Nil(t, err)
	assert.Equal(t, 1, len(ctx.headers))
	assert.Equal(t, "application/json", ctx.headers["Content-Type"])
}

func TestApiContext_ISetQueryParamWithValue(t *testing.T) {

	ctx := setupTestContext()
	err := ctx.ISetQueryParamWithValue("page", "1")

	assert.Nil(t, err)
	assert.Equal(t, 1, len(ctx.queryParams))
	assert.Equal(t, "1", ctx.queryParams["page"])
}

func TestApiContext_ISetQueryParamsTo(t *testing.T) {

	ctx := setupTestContext()

	dt := &godog.Table{
		Rows: []*messages.PickleStepArgument_PickleTable_PickleTableRow{
			{
				Cells: []*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{
					{
						Value: "q1",
					},
					{
						Value: "v1",
					},
				},
			},
			{
				Cells: []*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{
					{
						Value: "q2",
					},
					{
						Value: "v2",
					},
				},
			},
		},
	}

	err := ctx.ISetQueryParamsTo(dt)

	assert.Nil(t, err)
	assert.Equal(t, "v1", ctx.queryParams["q1"])
	assert.Equal(t, "v2", ctx.queryParams["q2"])
}

func TestApiContext_ISendRequestTo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := make(map[string]string)
		p["result"] = "success"
		if err := json.NewEncoder(w).Encode(p); err != nil {
			w.WriteHeader(500)
		}
	}))

	defer ts.Close()

	ctx := setupTestContext().
		WithBaseURL(ts.URL)

	if err := ctx.ISetQueryParamWithValue("page", "1"); err != nil {
		t.Fatal(err)
	}

	if err := ctx.ISetHeaderWithValue("Content-Type", "application/json"); err != nil {
		t.Fatal(err)
	}

	err := ctx.ISendRequestTo("GET", "/")

	assert.Nil(t, err)
	assert.NotNil(t, ctx.lastResponse)
	assert.Equal(t, 200, ctx.lastResponse.StatusCode)
	assert.NotNil(t, ctx.TheResponseCodeShouldBe(400))
	assert.Nil(t, ctx.TheResponseShouldBeAValidJSON())
	assert.Nil(t, ctx.TheResponseShouldMatchJSON(&godog.DocString{
		Content: "{\"result\": \"success\"}",
	}))
}

func TestApiContext_ISendRequestToWithBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := make(map[string]string)
		p["result"] = "success"
		if err := json.NewEncoder(w).Encode(p); err != nil {
			w.WriteHeader(500)
		}
	}))

	defer ts.Close()

	ctx := setupTestContext().
		WithBaseURL(ts.URL)

	if err := ctx.ISetQueryParamWithValue("page", "1"); err != nil {
		t.Fatalf("cannot set query param on the request %v", err)
	}

	if err := ctx.ISetHeaderWithValue("Content-Type", "application/json"); err != nil {
		t.Fatalf("cannot set header on the request %v", err)
	}

	reqBody := "{ \"name\": \"Bruno\"}"
	err := ctx.ISendRequestToWithBody("POST", "/", &godog.DocString{
		Content: reqBody,
	})

	assert.Nil(t, err)
	assert.NotNil(t, ctx.lastResponse)
	assert.Equal(t, 200, ctx.lastResponse.StatusCode)
	assert.Equal(t, "POST", ctx.lastRequest.Method)
}

func TestApiContext_TheResponseHeaderShouldHaveValue(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Some-Header", "hello")
	}))

	defer ts.Close()
	ctx := setupTestContext().
		WithBaseURL(ts.URL)

	err := ctx.ISendRequestTo("GET", "/")

	assert.Nil(t, err)
	assert.Nil(t, ctx.TheResponseHeaderShouldHaveValue("X-Some-Header", "hello"))
	assert.NotNil(t, ctx.TheResponseHeaderShouldHaveValue("non-existing-header", "hello"))
}

func TestApiContext_TheResponseShouldMatchJsonSchema(t *testing.T) {

	p := make(map[string]interface{})
	p["firstName"] = "Bruno"
	p["lastName"] = "Paz"
	p["age"] = 30

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Some-Header", "hello")
		if err := json.NewEncoder(w).Encode(p); err != nil {
			w.WriteHeader(500)
		}
	}))

	defer ts.Close()
	ctx := setupTestContext().
		WithBaseURL(ts.URL)

	err := ctx.ISendRequestTo("GET", "/")

	assert.Nil(t, err)
	assert.Nil(t, ctx.TheResponseShouldMatchJsonSchema("person.json"))
	assert.NotNil(t, ctx.TheResponseShouldMatchJsonSchema("coordinates.json"))
}

func TestApiContext_TheJSONPathShouldHaveValue(t *testing.T) {

	f, err := ioutil.ReadFile(filepath.Join("testdata", "test_json_path.json"))

	if err != nil {
		t.Error(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_, err := w.Write(f)

		if err != nil {
			w.WriteHeader(500)
		}
	}))

	defer ts.Close()
	ctx := setupTestContext().
		WithBaseURL(ts.URL).
		WithDebug(false)

	err = ctx.ISendRequestTo("GET", "/")

	assert.Nil(t, err)
	assert.NotNil(t, ctx.lastResponse)
	assert.Nil(t, ctx.TheJSONPathShouldHaveValue("$.a", "a"))
	assert.Nil(t, ctx.TheJSONPathShouldHaveValue("$.b", "2"))
	assert.Nil(t, ctx.TheJSONPathShouldHaveValue("$.c", "3.5"))
	assert.Nil(t, ctx.TheJSONPathShouldHaveValue("$.d", "true"))
}

func TestApiContext_TheJSONPathShouldMatch(t *testing.T) {

	var respBody = map[string]string{
		"name": "Bruno",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(respBody); err != nil {
			w.WriteHeader(500)
		}
	}))

	defer ts.Close()
	ctx := setupTestContext().
		WithBaseURL(ts.URL).
		WithDebug(false)

	err := ctx.ISendRequestTo("GET", "/")

	assert.Nil(t, err)
	assert.Nil(t, ctx.TheJSONPathShouldMatch("$.name", "^[a-zA-Z]+$"))
	assert.Error(t, ctx.TheJSONPathShouldMatch("$.name", "^[0-9]+$"))
}

func TestApiContext_TheJSONPathHaveCount(t *testing.T) {

	f, err := ioutil.ReadFile(filepath.Join("testdata", "test_json_path.json"))

	if err != nil {
		t.Error(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_, err := w.Write(f)

		if err != nil {
			w.WriteHeader(500)
		}
	}))

	defer ts.Close()
	ctx := setupTestContext().
		WithBaseURL(ts.URL).
		WithDebug(false)

	err = ctx.ISendRequestTo("GET", "/")

	assert.Nil(t, err)
	assert.NotNil(t, ctx.lastResponse)
	assert.Nil(t, ctx.TheJSONPathHaveCount("$.list", 2))
}

func TestApiContext_TheJSONPathHaveCountRoot(t *testing.T) {

	f, err := ioutil.ReadFile(filepath.Join("testdata", "test_root_array.json"))

	if err != nil {
		t.Error(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_, err := w.Write(f)

		if err != nil {
			w.WriteHeader(500)
		}
	}))

	defer ts.Close()
	ctx := setupTestContext().
		WithBaseURL(ts.URL).
		WithDebug(false)

	err = ctx.ISendRequestTo("GET", "/")

	assert.Nil(t, err)
	assert.NotNil(t, ctx.lastResponse)
	assert.Nil(t, ctx.TheJSONPathHaveCount("$", 2))
}

func TestApiContext_TheResponseBodyShouldContain(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello world!"))
	}))

	defer ts.Close()
	ctx := setupTestContext().
		WithBaseURL(ts.URL)

	err := ctx.ISendRequestTo("GET", "/")

	assert.Nil(t, err)
	assert.Nil(t, ctx.TheResponseBodyShouldContain("hello"))
	assert.Error(t, ctx.TheResponseBodyShouldContain("abc"))
}

func TestApiContext_TheResponseBodyShouldBePresent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{ \"hello\": \"value\" }"))
	}))

	defer ts.Close()
	ctx := setupTestContext().
		WithBaseURL(ts.URL)

	err := ctx.ISendRequestTo("GET", "/")

	assert.Nil(t, err)
	assert.Nil(t, ctx.TheJSONPathShouldBePresent("$.hello"))
}

func TestApiContext_TheResponseBodyShouldMatch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello world!"))
	}))

	defer ts.Close()
	ctx := setupTestContext().
		WithBaseURL(ts.URL)

	err := ctx.ISendRequestTo("GET", "/")

	assert.Nil(t, err)
	assert.Nil(t, ctx.TheResponseBodyShouldMatch("^hello\\sworld!"))
}

func TestReset(t *testing.T) {
	ctx := setupTestContext()

	p := &messages.Pickle{}
	ctx.headers = map[string]string{
		"Content-Type": "application/json",
	}
	ctx.queryParams = map[string]string{
		"param": "test",
	}

	ctx.reset(p)

	assert.Empty(t, ctx.headers)
	assert.Empty(t, ctx.queryParams)
}
