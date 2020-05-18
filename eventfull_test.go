package eventfull

import (
	"testing"
)

func TestReadGenericJSONOK(t *testing.T) {
	if m, err := readGenericJSON("test_resources/test.json"); err != nil {
		t.Error("Failed reading generic json")
	} else {
		if m["string"] != "string" || m["number"] != 100.0 || m["boolean"] != true {
			t.Error("Error reading one of the fields")
		}
	}
}

func TestReadGenericJSONFail(t *testing.T) {
	if _, err := readGenericJSON("doesnotexists.json"); err == nil {
		t.Error("Failed negative test")
	}
}
