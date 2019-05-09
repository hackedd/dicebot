package dicebot

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Database interface {
	ReadValue(name, scope string) (string, bool)
	StoreValue(name, scope, value string) error
}

type JsonVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type JsonScope struct {
	Name      string         `json:"name"`
	Variables []JsonVariable `json:"variables"`
}

type JsonDatabase struct {
	filename string
	scopes   []JsonScope
}

func NewJsonDatabase(filename string) (Database, error) {
	db := &JsonDatabase{filename: filename}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return db, nil
		}
		return nil, err
	}

	err = json.Unmarshal(data, &db.scopes)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *JsonDatabase) getScope(name string) *JsonScope {
	for i := range db.scopes {
		if db.scopes[i].Name == name {
			return &db.scopes[i]
		}
	}
	return nil
}

func (db *JsonDatabase) getVariable(scope *JsonScope, name string) *JsonVariable {
	for i := range scope.Variables {
		if scope.Variables[i].Name == name {
			return &scope.Variables[i]
		}
	}
	return nil
}

func (db *JsonDatabase) save() error {
	if db.filename == "" {
		return nil
	}

	data, err := json.MarshalIndent(db.scopes, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(db.filename, data, 0644)
}

func (db *JsonDatabase) ReadValue(name, scope string) (string, bool) {
	s := db.getScope(scope)
	if s == nil {
		return "", false
	}
	v := db.getVariable(s, name)
	if v == nil {
		return "", false
	}
	return v.Value, true
}

func (db *JsonDatabase) StoreValue(name, scope, value string) error {
	s := db.getScope(scope)
	if s == nil {
		s = &JsonScope{Name: scope}
		db.scopes = append(db.scopes, JsonScope{Name: scope})
		s = &db.scopes[len(db.scopes)-1]
	}
	v := db.getVariable(s, name)
	if v == nil {
		s.Variables = append(s.Variables, JsonVariable{name, value})
		v = &s.Variables[len(s.Variables)-1]
	} else {
		v.Value = value
	}
	return db.save()
}
