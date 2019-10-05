package context

import (
	"encoding/json"
	"github.com/DATA-DOG/godog/gherkin"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/godog"
	"github.com/stretchr/testify/assert"
)

func TestSetSingleHeader(t *testing.T) {
	s := &godog.Suite{}
	ctx := NewAPIContext(s, "http://example.com")
	ctx.ISetHeaderWithValue("Content-Type", "application/json")

	assert.Equal(t, 1, len(ctx.headers))
	assert.Equal(t, "application/json", ctx.headers["Content-Type"])
}

func TestSetSingleQueryParam(t *testing.T) {
	s := &godog.Suite{}
	ctx := NewAPIContext(s, "http://example.com")
	ctx.ISetQueryParamWithValue("page", "1")

	assert.Equal(t, 1, len(ctx.queryParams))
	assert.Equal(t, "1", ctx.queryParams["page"])
}

func TestSendRequest(t *testing.T) {
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
	assert.Nil(t, ctx.TheResponseShouldBeAValidJson())
	assert.Nil(t, ctx.TheResponseShouldMatchJson(&gherkin.DocString{
		Content: "{\"result\": \"success\"}",
	}))
}

//func TestJsonPathMatchers(t *testing.T) {
//
//	/*f, err := ioutil.ReadFile(filepath.Join("testdata","test_json_path.json"))
//
//	if err != nil {
//		t.Error(err)
//	}*/
//
//	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.Header().Set("Content-Type", "application/json")
//		json.NewEncoder(w).Encode("{\"a\": \"a\"}")
//	}))
//
//	defer ts.Close()
//	s := &godog.Suite{}
//
//	ctx := NewAPIContext(s, ts.URL)
//
//	ctx.ISendRequestTo("GET", "/")
//
//	assert.NotNil(t, ctx.lastResponse)
//	assert.Nil(t, ctx.TheJsonPathShouldHaveValue("$.a", "a"))
//	assert.Nil(t, ctx.TheJsonPathShouldHaveValue("$.b", 1))
//	assert.Nil(t, ctx.TheJsonPathShouldHaveValue("$.c", 3.50))
//}
