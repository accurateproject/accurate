
package console

import (
	"encoding/json"
	"testing"
)

func TestToJSON(t *testing.T) {
	jsn := ToJSON(`TimeStart="Test"     Crazy = 1 Mama=true coco Test=1`)
	expected := `{"TimeStart":"Test","Crazy":1,"Mama":true,"Test":1}`
	if string(jsn) != expected {
		t.Errorf("Expected: %s got: %s", expected, jsn)
	}
}

func TestToJSONValid(t *testing.T) {
	jsn := ToJSON(`TimeStart="Test"     Crazy = 1 Mama=true coco Test=1`)
	a := make(map[string]interface{})
	if err := json.Unmarshal(jsn, &a); err != nil {
		t.Error("Error unmarshaling generated json: ", err)
	}
}

func TestToJSONEmpty(t *testing.T) {
	jsn := ToJSON("")
	if string(jsn) != `{"Item":""}` {
		t.Error("Error empty: ", string(jsn))
	}
}

func TestToJSONString(t *testing.T) {
	jsn := ToJSON("1002")
	if string(jsn) != `{"Item":"1002"}` {
		t.Error("Error string: ", string(jsn))
	}
}

func TestToJSONArrayNoSpace(t *testing.T) {
	jsn := ToJSON(`Param=["id1","id2","id3"] Another="Patram"`)
	if string(jsn) != `{"Param":["id1","id2","id3"],"Another":"Patram"}` {
		t.Error("Error string: ", string(jsn))
	}
}

func TestToJSONArraySpace(t *testing.T) {
	jsn := ToJSON(`Param=["id1", "id2", "id3"]  Another="Patram"`)
	if string(jsn) != `{"Param":["id1", "id2", "id3"],"Another":"Patram"}` {
		t.Error("Error string: ", string(jsn))
	}
}

func TestFromJSON(t *testing.T) {
	line := FromJSON([]byte(`{"TimeStart":"Test","Crazy":1,"Mama":true,"Test":1}`), []string{"TimeStart", "Crazy", "Mama", "Test"})
	expected := `TimeStart="Test" Crazy=1 Mama=true Test=1`
	if line != expected {
		t.Errorf("Expected: %s got: '%s'", expected, line)
	}
}

func TestFromJSONInterestingFields(t *testing.T) {
	line := FromJSON([]byte(`{"TimeStart":"Test","Crazy":1,"Mama":true,"Test":1}`), []string{"TimeStart", "Test"})
	expected := `TimeStart="Test" Test=1`
	if line != expected {
		t.Errorf("Expected: %s got: '%s'", expected, line)
	}
}

func TestFromJSONString(t *testing.T) {
	line := FromJSON([]byte(`1002`), []string{"string"})
	expected := `"1002"`
	if line != expected {
		t.Errorf("Expected: %s got: '%s'", expected, line)
	}
}

func TestFromJSONArrayNoSpace(t *testing.T) {
	line := FromJSON([]byte(`{"Param":["id1","id2","id3"], "TimeStart":"Test", "Test":1}`), []string{"Param", "TimeStart", "Test"})
	expected := `Param=["id1","id2","id3"] TimeStart="Test" Test=1`
	if line != expected {
		t.Errorf("Expected: %s got: '%s'", expected, line)
	}
}

func TestFromJSONArraySpace(t *testing.T) {
	line := FromJSON([]byte(`{"Param":["id1", "id2", "id3"], "TimeStart":"Test", "Test":1}`), []string{"Param", "TimeStart", "Test"})
	expected := `Param=["id1", "id2", "id3"] TimeStart="Test" Test=1`
	if line != expected {
		t.Errorf("Expected: %s got: '%s'", expected, line)
	}
}
