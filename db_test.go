package dicebot

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

var testScopes = []JsonScope{
	{
		Name: "test",
		Variables: []JsonVariable{
			{"a", "1"},
			{"b", "2"},
		},
	},
}

func TestJsonDatabase_ReadValue(t *testing.T) {
	db := &JsonDatabase{
		scopes: append([]JsonScope(nil), testScopes...),
	}

	tests := []struct {
		name  string
		scope string
		value string
		ok    bool
	}{
		{"a", "test", "1", true},
		{"b", "test", "2", true},
		{"x", "test", "", false},
		{"a", "x", "", false},
	}

	for _, test := range tests {
		value, ok := db.ReadValue(test.name, test.scope)
		if value != test.value || ok != test.ok {
			t.Errorf("ReadValue(%v, %v) got %v %v expected %v %v", test.name, test.scope, value, ok, test.value, test.ok)
		}
	}
}

func TestJsonDatabase_StoreValue(t *testing.T) {
	db := &JsonDatabase{
		scopes: append([]JsonScope(nil), testScopes...),
	}

	tests := []struct {
		name  string
		scope string
		value string
	}{
		{"a", "test", "1"},
		{"b", "test", "2"},
		{"x", "test", "3"},
		{"a", "new", "4"},
	}

	for _, test := range tests {
		err := db.StoreValue(test.name, test.scope, test.value)
		if err != nil {
			t.Errorf("StoreValue(%v, %v, %v) unexpected %v", test.name, test.scope, test.value, err)
			continue
		}

		value, ok := db.ReadValue(test.name, test.scope)
		if value != test.value || !ok {
			t.Errorf("StoreValue(%v, %v, %v) got %v %v", test.name, test.scope, test.value, value, ok)
		}
	}
}

func WriteTempFile(t *testing.T, template, contents string) string {
	file, err := ioutil.TempFile("", template)
	if err != nil {
		t.Fatalf("Unable to create temporary file %v: %v", template, err)
	}

	if _, err := file.Write([]byte(contents)); err != nil {
		t.Fatalf("Unable to write temporary file %v: %v", template, err)
	}

	if err := file.Close(); err != nil {
		t.Fatalf("Unable to close temporary file %v: %v", template, err)
	}

	return file.Name()
}

func TestNewJsonDatabase(t *testing.T) {
	json := `[
  {
    "name": "test",
    "variables": [
      {"name": "a", "value": "1"},
      {"name": "b", "value": "2"}
    ]
  }
]`

	filename := WriteTempFile(t, "test*.json", json)
	defer os.Remove(filename)

	db, err := NewJsonDatabase(filename)
	if err != nil {
		t.Fatalf("NewJsonDatabase(): %v", err)
	}

	scopes := db.(*JsonDatabase).scopes
	if !reflect.DeepEqual(scopes, testScopes) {
		t.Errorf("NewJsonDatabase(): expected %+v got %+v", testScopes, scopes)
	}
}

func TestNewJsonDatabaseEmpty(t *testing.T) {
	_, err := NewJsonDatabase("does-not-exist.json")
	if err != nil {
		t.Fatalf("NewJsonDatabase(): %v", err)
	}
}

func TestJsonDatabase_StoreValueSaves(t *testing.T) {
	filename := WriteTempFile(t, "test*.json", "")
	defer os.Remove(filename)

	db := &JsonDatabase{
		filename: filename,
	}
	err := db.StoreValue("a", "test", "1")
	if err != nil {
		t.Errorf("StoreValue(): %v", err)
	}

	data, err := ioutil.ReadFile(db.filename)
	if err != nil {
		t.Errorf("ReadFile(): %v", err)
	}

	json := `[
  {
    "name": "test",
    "variables": [
      {
        "name": "a",
        "value": "1"
      }
    ]
  }
]`
	if string(data) != json {
		t.Errorf("ReadFile(): expected %v\ngot %v", json, string(data))
	}
}
