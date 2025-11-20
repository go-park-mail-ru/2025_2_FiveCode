package apiutils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteErrorAndJSON(t *testing.T) {
    w := httptest.NewRecorder()
    WriteError(w, http.StatusBadRequest, "oops")
    if w.Result().StatusCode != http.StatusBadRequest {
        t.Fatalf("expected 400 got %d", w.Result().StatusCode)
    }
    var out ErrorResponse
    if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
        t.Fatalf("json decode: %v", err)
    }
    if out.Error != "oops" {
        t.Fatalf("unexpected error text: %s", out.Error)
    }
}

func TestGetHelpersAndStrictUnmarshal(t *testing.T) {
    b := true
    if !GetBool(&b, false) { t.Fatalf("GetBool failed") }
    s := "x"
    if GetString(&s, "def") != "x" { t.Fatalf("GetString failed") }
    i := 5
    if GetInt(&i, 1) != 5 { t.Fatalf("GetInt failed") }

    // StrictUnmarshal should error on unknown fields
    type X struct { A int `json:"a"` }
    data := []byte(`{"a":1,"b":2}`)
    var x X
    if err := StrictUnmarshal(data, &x); err == nil {
        t.Fatalf("expected error from StrictUnmarshal with unknown fields")
    }

    // parseGovalidatorError behaviour
    f, m := parseGovalidatorError("field: message")
    if f != "field" || m != "message" { t.Fatalf("parseGovalidatorError mismatch: %s %s", f, m) }
}
