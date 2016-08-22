
package utils

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

var testLocal = flag.Bool("local", false, "Perform the tests only on local test environment, not by default.") // This flag will be passed here via "go test -local" args

type TestContent struct {
	Var1 string
	Var2 string
}

func TestHttpJsonPoster(t *testing.T) {
	if !*testLocal {
		return
	}
	content := &TestContent{Var1: "Val1", Var2: "Val2"}
	jsn, _ := json.Marshal(content)
	filePath := "/tmp/cgr_test_http_poster.json"
	if _, _, err := HttpPoster("http://localhost:8080/invalid", true, jsn, CONTENT_JSON, 3, filePath, false); err != nil {
		t.Error(err)
	}
	if readBytes, err := ioutil.ReadFile(filePath); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(jsn, readBytes) {
		t.Errorf("Expecting: %q, received: %q", string(jsn), string(readBytes))
	}
	if err := os.Remove(filePath); err != nil {
		t.Error("Failed removing file: ", filePath)
	}
}

func TestHttpBytesPoster(t *testing.T) {
	if !*testLocal {
		return
	}
	content := []byte(`Test
		Test2
		`)
	filePath := "/tmp/test_http_poster.http"
	if _, _, err := HttpPoster("http://localhost:8080/invalid", true, content, CONTENT_TEXT, 3, filePath, false); err != nil {
		t.Error(err)
	}
	if readBytes, err := ioutil.ReadFile(filePath); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(content, readBytes) {
		t.Errorf("Expecting: %q, received: %q", string(content), string(readBytes))
	}
	if err := os.Remove(filePath); err != nil {
		t.Error("Failed removing file: ", filePath)
	}
}
