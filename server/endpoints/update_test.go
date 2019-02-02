package endpoints_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wikisophia/api-arguments/server/arguments"
)

func TestPatchLive(t *testing.T) {
	original := parseArgument(t, readFile(t, "../samples/save-request.json"))
	update := parseArgument(t, readFile(t, "../samples/update-request.json"))

	server := newServerForTests()
	id := doSaveObject(t, server, original)
	doValidUpdate(t, server, id, update.Premises)
	rr := doGetArgument(server, id)
	assertSuccessfulJSON(t, rr)
	actual := parseArgument(t, rr.Body.Bytes())
	assertArgumentsMatch(t, arguments.Argument{
		Conclusion: original.Conclusion,
		Premises:   update.Premises,
	}, actual)
}

func TestUpdateLocation(t *testing.T) {
	original := parseArgument(t, readFile(t, "../samples/save-request.json"))
	update := parseArgument(t, readFile(t, "../samples/update-request.json"))

	server := newServerForTests()
	id := doSaveObject(t, server, original)
	rr := doValidUpdate(t, server, id, update.Premises)
	assert.Equal(t, "/arguments/1/version/2", rr.Header().Get("Location"))
}

func TestPatchUnknown(t *testing.T) {
	server := newServerForTests()
	payload := `{"premises":["Socrates is a man", "All men are mortal"]}`
	rr := doRequest(server, httptest.NewRequest("PATCH", "/arguments/1", strings.NewReader(payload)))
	assert.Equal(t, http.StatusNotFound, rr.Code, "body: %s", rr.Body.String())
	assert.Equal(t, "text/plain; charset=utf-8", rr.Header().Get("Content-Type"))
}

func TestMalformedPatch(t *testing.T) {
	assertPatchRejected(t, "not json")
}

func TestPatchConclusion(t *testing.T) {
	assertPatchRejected(t, `{"conclusion":"Socrates is mortal","premises":["Socrates is a man", "All men are mortal"]}`)
}

func TestPatchOnePremise(t *testing.T) {
	assertPatchRejected(t, `{"premises":["Socrates is a man"]}`)
}

func TestPatchEmpty(t *testing.T) {
	assertPatchRejected(t, `{"premises":["Socrates is a man", ""]}`)
}

func assertPatchRejected(t *testing.T, payload string) {
	t.Helper()
	original := parseArgument(t, readFile(t, "../samples/save-request.json"))

	server := newServerForTests()
	id := doSaveObject(t, server, original)
	rr := doRequest(server, httptest.NewRequest("PATCH", "/arguments/"+strconv.FormatInt(id, 10), strings.NewReader(payload)))
	assert.Equal(t, http.StatusBadRequest, rr.Code, "body: %s", rr.Body.String())
	assert.Equal(t, "text/plain; charset=utf-8", rr.Header().Get("Content-Type"))
}