package cstore

import (
	"errors"
	"os"
	"testing"
)

const (
	BASE_DIR = "./testing_work/dir2"
)

func removeBaseDir(t *testing.T) {
	if _, err := os.Stat(BASE_DIR); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}

	if err := os.RemoveAll(BASE_DIR); err != nil {
		t.Fatal(err)
	}
}

type Text struct {
	Text string
}

func TestToml(t *testing.T) {
	removeBaseDir(t)

	m, err := NewManager("testing", BASE_DIR)
	if err != nil {
		t.Fatal(err)
	}

	cs, err := m.New("text.toml", TOML)
	if err != nil {
		t.Fatal(err)
	}

	sText := Text{
		Text: "this message",
	}

	err = cs.SaveWithoutValidate(&sText)
	if err != nil {
		t.Fatal(err)
	}

	gText := Text{}
	err = cs.GetWithoutValidate(&gText)
	if err != nil {
		t.Fatal(err)
	}

	err = cs.Remove()
	if err != nil {
		t.Fatal(err)
	}

	err = cs.GetWithoutValidate(&Text{})
	if !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

type Sample struct {
	Name string
}

func (s *Sample) Validate() error {
	if s.Name == "" {
		return errors.New("name should not empty")
	}

	return nil
}

func TestJSON(t *testing.T) {
	removeBaseDir(t)

	m, err := NewManager("testing", BASE_DIR)
	if err != nil {
		t.Fatal(err)
	}

	cs, err := m.New("sample.json", JSON)
	if err != nil {
		t.Fatal(err)
	}
	s := Sample{Name: "sample name"}

	err = cs.Save(&s)
	if err != nil {
		t.Fatal(err)
	}

	s2 := Sample{}
	err = cs.Get(&s2)
	if err != nil {
		t.Fatal(err)
	}

	if s2.Name != "sample name" {
		t.Errorf("expect:%s but %s", "sample name", s2.Name)
	}

	err = cs.Remove()
	if err != nil {
		t.Fatal(err)
	}

	err = cs.Get(&Sample{})
	if !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func TestManager(t *testing.T) {
	removeBaseDir(t)

	m, err := NewManager("testing", BASE_DIR)
	if err != nil {
		t.Fatal(err)
	}

	name := "TestManager.json"
	_, err = m.New(name, JSON)
	if err != nil {
		t.Fatal(err)
	}

	if cs := m.Get(name); cs == nil {
		t.Fatalf("Get() should return %s, because called New()", name)
	}

	if cs := m.Remove(name); cs == nil {
		t.Fatalf("Remove() should return %s, because called New()", name)
	}

	if cs := m.Get(name); cs != nil {
		t.Fatalf("Get() should not return %s, because called Remove()", name)
	}
}
