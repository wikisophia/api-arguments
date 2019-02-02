package endpoints_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveGetRoundtrip(t *testing.T) {
	expected := parseArgument(t, readFile(t, "../samples/save-request.json"))
	server := newServerForTests()
	id := doSaveObject(t, server, expected)
	rr := doGetArgument(server, id)
	assertSuccessfulJSON(t, rr)
	actual := parseArgument(t, rr.Body.Bytes())
	assertArgumentsMatch(t, expected, actual)
}

func TestSaveNoConclusion(t *testing.T) {
	rr := doSaveArgument(newServerForTests(), `{"premises":["Socrates is a man","All men are mortal"]}`)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "text/plain; charset=utf-8", rr.Header().Get("Content-Type"))
}

func TestSaveNoPremises(t *testing.T) {
	rr := doSaveArgument(newServerForTests(), `{"conclusion":"Socrates is mortal"}`)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "text/plain; charset=utf-8", rr.Header().Get("Content-Type"))
}

func TestSaveNotJSON(t *testing.T) {
	rr := doSaveArgument(newServerForTests(), `bad payload`)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "text/plain; charset=utf-8", rr.Header().Get("Content-Type"))
}